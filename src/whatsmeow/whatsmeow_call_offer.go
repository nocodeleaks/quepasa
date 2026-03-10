package whatsmeow

// Structures and helpers to decode the CallOffer "Data" payload
// derived from the sample JSON files in docs/WHATSMEOW-call_offer_example (live / history).
//
// Two variants have been observed for the "voip_settings" node:
//  1) Content is an object with fields { "base64": "...", "decoded": { ... }, "_note": "..." }
//  2) Content is already the decoded object (no wrapper, plain object form).
//
// The root Data object itself has the shape:
// {
//   "Tag": "offer",
//   "Attrs": { "call-id": "...", ... },
//   "Content": [ Node, Node, ... ]
// }
// Each Node has: Tag, Attrs (map) and Content which can be:
//  - null
//  - string (e.g. base64 tokens, enc, rte, hbh_key, etc.)
//  - another array of Nodes (case Tag == "relay")
//  - special object (voip_settings wrapper or already decoded voip_settings)
//
// We keep flexibility by using json.RawMessage in Content and providing helpers
// to interpret specific nodes.

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow/types/events"
)

// WhatsmeowCallOffer embeds the original events.CallOffer and adds a normalized Data tree.
// Data is decoded according to documented JSON examples (Tag/Attrs/Content nodes).
type WhatsmeowCallOffer struct {
	TimestampRaw string `json:"Timestamp,omitempty"`

	// * /docs/WHATSMEOW-call_offer_example.json
	events.CallOffer

	// Data is the normalized structure of the "Data" field from CallOffer.
	Data OfferDataNode `json:"Data"`

	// lazy caches
	voipOnce           sync.Once
	voipDecoded        map[string]interface{}
	voipWrapped        bool
	relayOnce          sync.Once
	relayTokens        []string
	relayBlockOnce     sync.Once
	relayBlock         *RelayBlock
	relayCandOnce      sync.Once
	relayHasCandidates bool
	dataOnce           sync.Once
}

// RelayBlock contains relay-only call material (tokens, relay candidates, keys).
// This is critical for relay media-plane work (SRTP/relay).
type RelayBlock struct {
	UUID      string
	SelfPID   string
	PeerPID   string
	Tokens    []RelayToken
	Auth      []RelayToken
	Key       string
	HBHKey    string
	TE2       []RelayTE2
	Protocols []string
}

// EncBlock contains the opaque <enc> payload from offers (often base64-encoded).
// This is a prime suspect for relay-only media-plane/TURN short-term key derivation.
// Never log Raw/RawB64 directly outside of explicit dump files.
type EncBlock struct {
	Type        string
	V           string
	Raw         []byte `json:"-"`
	RawLen      int
	ContentKind string
}

type RelayToken struct {
	ID string
	// Value is a stable base64 representation of the token bytes.
	// Do not assume this is the raw token/auth bytes.
	Value string
	// Raw contains the original bytes (not marshaled) when available.
	Raw []byte `json:"-"`
}

func (t RelayToken) Bytes() []byte {
	if len(t.Raw) > 0 {
		return t.Raw
	}
	s := strings.TrimSpace(t.Value)
	if s == "" {
		return nil
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil && len(b) > 0 {
		return b
	}
	return []byte(s)
}

type binaryContentWrapped struct {
	Base64 string `json:"base64"`
	Len    int    `json:"len"`
}

type RelayTE2 struct {
	RelayName   string
	RelayID     string
	TokenID     string
	AuthTokenID string
	Protocol    string
	C2RRtt      string

	// Payload is the raw bytes content of the <te2> node when available.
	// This is often opaque binary data and should not be logged directly.
	Payload    []byte `json:"-"`
	PayloadB64 string
	PayloadLen int
	IPv6Prefix string
	MarkerHex  string
	RelayTailHex string
	SuffixHex  string
}

func formatRelayTE2IPv6Prefix(b []byte) string {
	if len(b) != 8 {
		return ""
	}
	return fmt.Sprintf("%02x%02x:%02x%02x:%02x%02x:%02x%02x",
		b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7],
	)
}

func parseRelayTE2Payload(te *RelayTE2) {
	if te == nil || len(te.Payload) != 18 {
		return
	}
	te.IPv6Prefix = formatRelayTE2IPv6Prefix(te.Payload[:8])
	te.MarkerHex = hex.EncodeToString(te.Payload[8:12])
	te.RelayTailHex = hex.EncodeToString(te.Payload[12:16])
	te.SuffixHex = hex.EncodeToString(te.Payload[16:18])
}

// OfferDataNode / RawNode definitions
type OfferDataNode struct {
	Tag     string            `json:"Tag"`
	Attrs   map[string]string `json:"Attrs"`
	Content []RawNode         `json:"Content"`
}

type RawNode struct {
	Tag     string            `json:"Tag"`
	Attrs   map[string]string `json:"Attrs"`
	Content json.RawMessage   `json:"Content"`
}

type VoipSettingsWrapped struct {
	Note    string                 `json:"_note,omitempty"`
	Base64  string                 `json:"base64"`
	Decoded map[string]interface{} `json:"decoded"`
}

type VoipSettingsPlain map[string]interface{}

// NewWhatsmeowCallOffer builds instance embedding evt and populates a minimal normalized Data tree.
func NewWhatsmeowCallOffer(evt *events.CallOffer) *WhatsmeowCallOffer {
	w := &WhatsmeowCallOffer{}
	if evt != nil {
		w.CallOffer = *evt
	}
	// Derive TimestampRaw
	if f := reflect.ValueOf(&w.CallOffer).Elem().FieldByName("Timestamp"); f.IsValid() && f.CanInterface() {
		if t, ok := f.Interface().(time.Time); ok && !t.IsZero() {
			w.TimestampRaw = t.UTC().Format(time.RFC3339)
		}
	}
	if w.Data.Tag == "" {
		w.Data.Tag = "offer"
	}
	if w.Data.Attrs == nil {
		w.Data.Attrs = map[string]string{}
	}
	if w.CallID != "" {
		w.Data.Attrs["call-id"] = w.CallID
	}
	// Attempt to copy CallCreator / Joinable via reflection
	rv := reflect.ValueOf(&w.CallOffer).Elem()
	if f := rv.FieldByName("CallCreator"); f.IsValid() && f.Kind() == reflect.String && f.String() != "" {
		w.Data.Attrs["call-creator"] = f.String()
	}
	if f := rv.FieldByName("Joinable"); f.IsValid() {
		val := ""
		switch f.Kind() {
		case reflect.String:
			s := f.String()
			if s == "1" || strings.ToLower(s) == "true" {
				val = "1"
			} else {
				val = s
			}
		case reflect.Bool:
			if f.Bool() {
				val = "1"
			} else {
				val = "0"
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if f.Int() != 0 {
				val = "1"
			} else {
				val = "0"
			}
		}
		if val != "" {
			w.Data.Attrs["joinable"] = val
		}
	}
	// Early best-effort hydration; GetData() will ensure idempotent finalization.
	w.hydrateFromEventData()
	w.normalizeVoipSettings()
	return w
}

// hydrateFromEventData tries to reflect the embedded events.CallOffer.Data field (if present)
// and convert its node tree into our OfferDataNode/RawNode representation.
// This is best-effort: if structures differ or already populated we skip.
func (w *WhatsmeowCallOffer) hydrateFromEventData() {
	if len(w.Data.Content) > 0 { // already filled (maybe from JSON in older design)
		return
	}
	rv := reflect.ValueOf(&w.CallOffer).Elem()
	f := rv.FieldByName("Data")
	if !f.IsValid() || f.IsZero() {
		return
	}
	// Dereference pointers
	fval := f
	if fval.Kind() == reflect.Ptr && !fval.IsNil() {
		fval = fval.Elem()
	}
	// Expect root struct with Tag/Attrs/Content
	rootNode, ok := toRawNode(fval)
	if !ok {
		return
	}
	// rootNode represents the <offer>; adopt its Attrs if ours are minimal.
	if len(rootNode.Attrs) > 0 {
		for k, v := range rootNode.Attrs {
			if _, exists := w.Data.Attrs[k]; !exists {
				w.Data.Attrs[k] = v
			}
		}
	}
	// Unmarshal its Content (children) into []RawNode and assign.
	if len(rootNode.Content) > 0 {
		var children []RawNode
		if err := json.Unmarshal(rootNode.Content, &children); err == nil {
			w.Data.Content = children
		}
	}
}

// normalizeVoipSettings finds the voip_settings node (if any) and canonicalizes its Content:
// - If Content is a base64 string containing JSON -> wrap as { base64: original, decoded: <object> }
// - If Content is a JSON string representing an object -> decoded directly (base64 field empty)
// - If already a wrapper with base64/decoded -> leave as-is
// This makes later decoding simpler and consistent.
func (w *WhatsmeowCallOffer) normalizeVoipSettings() {
	node := w.Data.FindFirst("voip_settings")
	if node == nil || len(node.Content) == 0 {
		return
	}
	trimmed := strings.TrimSpace(string(node.Content))
	if len(trimmed) == 0 {
		return
	}
	// Already wrapper?
	var existing VoipSettingsWrapped
	if json.Unmarshal(node.Content, &existing) == nil && (existing.Base64 != "" || existing.Decoded != nil) {
		return
	}
	// Quoted string?
	var asString string
	if json.Unmarshal(node.Content, &asString) == nil {
		s := strings.TrimSpace(asString)
		if strings.HasPrefix(s, "{") { // JSON object in string
			var obj map[string]interface{}
			if json.Unmarshal([]byte(s), &obj) == nil {
				wrapper := VoipSettingsWrapped{Base64: "", Decoded: obj, Note: "decoded from JSON string (already uncompressed)"}
				if b, err := json.Marshal(wrapper); err == nil {
					node.Content = b
				}
			}
			return
		}
		if raw, err := base64.StdEncoding.DecodeString(s); err == nil { // base64 JSON
			var obj map[string]interface{}
			if json.Unmarshal(raw, &obj) == nil {
				wrapper := VoipSettingsWrapped{Base64: s, Decoded: obj, Note: "decoded from base64 string"}
				if b, err := json.Marshal(wrapper); err == nil {
					node.Content = b
				}
				return
			}
		}
		return
	}
	// Plain object (not wrapper)
	var plain map[string]interface{}
	if json.Unmarshal(node.Content, &plain) == nil && len(plain) > 0 {
		wrapper := VoipSettingsWrapped{Base64: "", Decoded: plain, Note: "decoded from plain object"}
		if b, err := json.Marshal(wrapper); err == nil {
			node.Content = b
		}
	}
}

// toRawNode converts a reflected struct (expected fields Tag, Attrs, Content) into RawNode recursively.
func toRawNode(v reflect.Value) (RawNode, bool) {
	var rn RawNode
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return rn, false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return rn, false
	}
	tagField := v.FieldByName("Tag")
	if !tagField.IsValid() || tagField.Kind() != reflect.String {
		return rn, false
	}
	rn.Tag = tagField.String()
	// Attrs
	attrsField := v.FieldByName("Attrs")
	if attrsField.IsValid() {
		rn.Attrs = map[string]string{}
		switch attrsField.Kind() {
		case reflect.Map:
			for _, key := range attrsField.MapKeys() {
				val := attrsField.MapIndex(key)
				rn.Attrs[fmt.Sprint(key.Interface())] = fmt.Sprint(val.Interface())
			}
		}
	}
	// Content
	contentField := v.FieldByName("Content")
	if contentField.IsValid() && contentField.Kind() != reflect.Invalid && !contentField.IsZero() {
		cf := contentField
		if cf.Kind() == reflect.Interface && !cf.IsNil() {
			cf = cf.Elem()
		}
		switch cf.Kind() {
		case reflect.String:
			b, _ := json.Marshal(cf.String())
			rn.Content = b
		case reflect.Slice, reflect.Array:
			// Could be []byte or []Node
			if cf.Type().Elem().Kind() == reflect.Uint8 { // []byte -> preserve as base64 to avoid UTF-8 replacement
				raw := cf.Bytes()
				wrap := binaryContentWrapped{Base64: base64.StdEncoding.EncodeToString(raw), Len: len(raw)}
				if b, err := json.Marshal(wrap); err == nil {
					rn.Content = b
				}
			} else {
				// Iterate children
				children := make([]RawNode, 0, cf.Len())
				for i := 0; i < cf.Len(); i++ {
					childVal := cf.Index(i)
					if cn, okc := toRawNode(childVal); okc {
						children = append(children, cn)
					}
				}
				if len(children) > 0 {
					if b, err := json.Marshal(children); err == nil {
						rn.Content = b
					}
				}
			}
		case reflect.Struct:
			if cn, okc := toRawNode(cf); okc {
				// Wrap single child as array for uniformity.
				children := []RawNode{cn}
				if b, err := json.Marshal(children); err == nil {
					rn.Content = b
				}
			}
		default:
			// Fallback to string representation
			b, _ := json.Marshal(fmt.Sprint(cf.Interface()))
			rn.Content = b
		}
	}
	return rn, true
}

// Helper: find first child node by tag.
func (o *OfferDataNode) FindFirst(tag string) *RawNode {
	for i := range o.Content {
		if o.Content[i].Tag == tag {
			return &o.Content[i]
		}
	}
	return nil
}

func (o *OfferDataNode) ExtractEncBlock() *EncBlock {
	n := o.FindFirst("enc")
	if n == nil {
		return nil
	}
	enc := &EncBlock{}
	if n.Attrs != nil {
		enc.Type = strings.TrimSpace(n.Attrs["type"])
		enc.V = strings.TrimSpace(n.Attrs["v"])
	}
	if len(n.Content) == 0 {
		return enc
	}

	// 1) Wrapped bytes { base64, len }
	{
		var w binaryContentWrapped
		if json.Unmarshal(n.Content, &w) == nil && strings.TrimSpace(w.Base64) != "" {
			b64 := strings.TrimSpace(w.Base64)
			if raw, err := base64.StdEncoding.DecodeString(b64); err == nil {
				enc.Raw = raw
				enc.RawLen = len(raw)
				enc.ContentKind = "wrapped_b64"
				return enc
			}
		}
	}

	// 2) String content: can be raw base64 (often without padding)
	{
		var s string
		if json.Unmarshal(n.Content, &s) == nil {
			s = strings.TrimSpace(s)
			if s == "" {
				return enc
			}
			if raw, ok := decodeB64Loose(s); ok {
				enc.Raw = raw
				enc.RawLen = len(raw)
				enc.ContentKind = "string_b64"
				return enc
			}
			enc.Raw = []byte(s)
			enc.RawLen = len(enc.Raw)
			enc.ContentKind = "string_raw"
			return enc
		}
	}

	enc.RawLen = len(n.Content)
	enc.ContentKind = "json_raw"
	return enc
}

func decodeB64Loose(s string) ([]byte, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, false
	}
	// Try as-is first.
	decoders := []*base64.Encoding{base64.RawStdEncoding, base64.StdEncoding, base64.RawURLEncoding, base64.URLEncoding}
	for _, enc := range decoders {
		if b, err := enc.DecodeString(s); err == nil && len(b) > 0 {
			return b, true
		}
	}
	// Try with padding for the padded encodings.
	if m := len(s) % 4; m != 0 {
		padded := s + strings.Repeat("=", 4-m)
		for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.URLEncoding} {
			if b, err := enc.DecodeString(padded); err == nil && len(b) > 0 {
				return b, true
			}
		}
	}
	return nil, false
}

// ExtractRelayBlock parses the <relay> node once into a structured form.
// Values may contain sensitive material; do not log raw values outside redacted paths.
func (o *OfferDataNode) ExtractRelayBlock() *RelayBlock {
	relay := o.FindFirst("relay")
	if relay == nil {
		return nil
	}

	b := &RelayBlock{
		UUID:    strings.TrimSpace(relay.Attrs["uuid"]),
		SelfPID: strings.TrimSpace(relay.Attrs["self_pid"]),
		PeerPID: strings.TrimSpace(relay.Attrs["peer_pid"]),
	}

	if len(relay.Content) == 0 {
		return b
	}

	var nodes []RawNode
	if json.Unmarshal(relay.Content, &nodes) != nil {
		return b
	}

	protocolUniq := map[string]struct{}{}
	for _, n := range nodes {
		switch n.Tag {
		case "token":
			id := strings.TrimSpace(n.Attrs["id"])
			var w binaryContentWrapped
			if json.Unmarshal(n.Content, &w) == nil && strings.TrimSpace(w.Base64) != "" {
				if raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(w.Base64)); err == nil {
					b.Tokens = append(b.Tokens, RelayToken{ID: id, Value: strings.TrimSpace(w.Base64), Raw: raw})
					break
				}
			}
			var s string
			if json.Unmarshal(n.Content, &s) == nil && s != "" {
				raw := []byte(s)
				b.Tokens = append(b.Tokens, RelayToken{ID: id, Value: base64.StdEncoding.EncodeToString(raw), Raw: raw})
			}
		case "auth_token":
			id := strings.TrimSpace(n.Attrs["id"])
			var w binaryContentWrapped
			if json.Unmarshal(n.Content, &w) == nil && strings.TrimSpace(w.Base64) != "" {
				if raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(w.Base64)); err == nil {
					b.Auth = append(b.Auth, RelayToken{ID: id, Value: strings.TrimSpace(w.Base64), Raw: raw})
					break
				}
			}
			var s string
			if json.Unmarshal(n.Content, &s) == nil && s != "" {
				raw := []byte(s)
				b.Auth = append(b.Auth, RelayToken{ID: id, Value: base64.StdEncoding.EncodeToString(raw), Raw: raw})
			}
		case "key":
			if b.Key != "" {
				break
			}
			// Content may be a raw string or wrapped bytes {base64,len}.
			var w binaryContentWrapped
			if json.Unmarshal(n.Content, &w) == nil && strings.TrimSpace(w.Base64) != "" {
				if raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(w.Base64)); err == nil {
					b.Key = strings.TrimSpace(string(raw))
					break
				}
			}
			var s string
			if json.Unmarshal(n.Content, &s) == nil {
				b.Key = strings.TrimSpace(s)
			}
		case "hbh_key":
			if b.HBHKey != "" {
				break
			}
			// Content may be a raw string or wrapped bytes {base64,len}.
			var w binaryContentWrapped
			if json.Unmarshal(n.Content, &w) == nil && strings.TrimSpace(w.Base64) != "" {
				if raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(w.Base64)); err == nil {
					b.HBHKey = strings.TrimSpace(string(raw))
					break
				}
			}
			var s string
			if json.Unmarshal(n.Content, &s) == nil {
				b.HBHKey = strings.TrimSpace(s)
			}
		case "te2":
			te := RelayTE2{
				RelayName:   strings.TrimSpace(n.Attrs["relay_name"]),
				RelayID:     strings.TrimSpace(n.Attrs["relay_id"]),
				TokenID:     strings.TrimSpace(n.Attrs["token_id"]),
				AuthTokenID: strings.TrimSpace(n.Attrs["auth_token_id"]),
				Protocol:    strings.TrimSpace(n.Attrs["protocol"]),
				C2RRtt:      strings.TrimSpace(n.Attrs["c2r_rtt"]),
			}
			// Content may be wrapped bytes {base64,len}.
			var w binaryContentWrapped
			if json.Unmarshal(n.Content, &w) == nil && strings.TrimSpace(w.Base64) != "" {
				b64 := strings.TrimSpace(w.Base64)
				te.PayloadB64 = b64
				te.PayloadLen = w.Len
				if raw, err := base64.StdEncoding.DecodeString(b64); err == nil {
					te.Payload = raw
					if te.PayloadLen <= 0 {
						te.PayloadLen = len(raw)
					}
					parseRelayTE2Payload(&te)
				}
			} else {
				// Fallback: plain string content (treat as base64 when decodable; else bytes).
				var s string
				if json.Unmarshal(n.Content, &s) == nil {
					s = strings.TrimSpace(s)
					if s != "" {
						te.PayloadB64 = s
						if raw, err := base64.StdEncoding.DecodeString(s); err == nil {
							te.Payload = raw
							te.PayloadLen = len(raw)
							parseRelayTE2Payload(&te)
						} else {
							te.Payload = []byte(s)
							te.PayloadLen = len(te.Payload)
						}
					}
				}
			}
			b.TE2 = append(b.TE2, te)
			if te.Protocol != "" {
				protocolUniq[te.Protocol] = struct{}{}
			}
		}
	}

	if len(protocolUniq) > 0 {
		b.Protocols = make([]string, 0, len(protocolUniq))
		for p := range protocolUniq {
			b.Protocols = append(b.Protocols, p)
		}
		sort.Strings(b.Protocols)
	}

	return b
}

// Decode voip_settings variants into a generic map.
func (o *OfferDataNode) DecodeVoipSettings() (map[string]interface{}, bool, error) {
	n := o.FindFirst("voip_settings")
	if n == nil || len(n.Content) == 0 {
		return nil, false, nil
	}
	var w VoipSettingsWrapped
	if err := json.Unmarshal(n.Content, &w); err == nil && (w.Base64 != "" || w.Decoded != nil) {
		if w.Decoded == nil && w.Base64 != "" {
			if b, derr := base64.StdEncoding.DecodeString(w.Base64); derr == nil {
				var obj map[string]interface{}
				if json.Unmarshal(b, &obj) == nil {
					w.Decoded = obj
				}
			}
		}
		return w.Decoded, true, nil
	}
	var plain VoipSettingsPlain
	if err := json.Unmarshal(n.Content, &plain); err == nil && len(plain) > 0 {
		return plain, false, nil
	}
	var asString string
	if err := json.Unmarshal(n.Content, &asString); err == nil && asString != "" {
		if raw, derr := base64.StdEncoding.DecodeString(asString); derr == nil {
			var obj map[string]interface{}
			if json.Unmarshal(raw, &obj) == nil {
				return obj, false, nil
			}
		}
	}
	return nil, false, nil
}

// Relay token extraction
func (o *OfferDataNode) ExtractRelayTokens() []string {
	b := o.ExtractRelayBlock()
	if b == nil || len(b.Tokens) == 0 {
		return nil
	}
	tokens := make([]string, 0, len(b.Tokens))
	for _, t := range b.Tokens {
		if strings.TrimSpace(t.Value) != "" {
			// Value is base64 of the raw bytes.
			tokens = append(tokens, strings.TrimSpace(t.Value))
		}
	}
	return tokens
}

// Relay candidate presence
func (o *OfferDataNode) HasRelayCandidates() bool {
	b := o.ExtractRelayBlock()
	return b != nil && len(b.TE2) > 0
}

// Valid heuristic
// IsValid tells if we can still join now: joinable==1, not expired, structural offer.
func (c *WhatsmeowCallOffer) IsValid() bool {
	if c == nil {
		return false
	}
	if c.CallID == "" {
		return false
	}
	d := c.GetData()
	if strings.ToLower(d.Tag) != "offer" {
		return false
	}
	if d.Attrs["joinable"] != "1" {
		return false
	}
	const ttl = 90 * time.Second
	t := time.Time{}
	rv := reflect.ValueOf(&c.CallOffer).Elem()
	if f := rv.FieldByName("Timestamp"); f.IsValid() && f.CanInterface() {
		if tt, ok := f.Interface().(time.Time); ok {
			t = tt
		}
	}
	if !t.IsZero() {
		if time.Since(t) > ttl {
			return false
		}
	}
	return true
}

// Summary helper
func (c *WhatsmeowCallOffer) Summary() string {
	d := c.GetData()
	return fmt.Sprintf("callid=%s joinable=%s ts=%s", c.CallID, d.Attrs["joinable"], c.TimestampRaw)
}

// IsValid performs a lightweight structural validation of the offer.
// Criteria (can be expanded later):
//  - Non-nil receiver
//  - CallID present
//  - Root Data tag == "offer"
//  - At least one media-indicating child (audio or enc)
// (structural validation folded into IsValid)

// GetVoipSettings returns (decodedMap, wrappedFlag) with lazy decode & cache
func (c *WhatsmeowCallOffer) GetVoipSettings() (map[string]interface{}, bool) {
	c.voipOnce.Do(func() {
		decoded, wrapped, _ := c.Data.DecodeVoipSettings()
		c.voipDecoded = decoded
		c.voipWrapped = wrapped
	})
	return c.voipDecoded, c.voipWrapped
}

// GetData returns a pointer to the normalized OfferDataNode ensuring hydration/normalization once.
func (c *WhatsmeowCallOffer) GetData() *OfferDataNode {
	c.dataOnce.Do(func() {
		if len(c.Data.Content) == 0 {
			c.hydrateFromEventData()
		}
		c.normalizeVoipSettings()
	})
	return &c.Data
}

// GetRelayTokens returns relay tokens cached lazily.
func (c *WhatsmeowCallOffer) GetRelayTokens() []string {
	c.relayOnce.Do(func() {
		b := c.GetRelayBlock()
		if b == nil || len(b.Tokens) == 0 {
			c.relayTokens = c.Data.ExtractRelayTokens()
			return
		}
		tokens := make([]string, 0, len(b.Tokens))
		for _, t := range b.Tokens {
			if t.Value != "" {
				tokens = append(tokens, t.Value)
			}
		}
		c.relayTokens = tokens
	})
	return c.relayTokens
}

// GetRelayBlock returns parsed relay metadata cached lazily.
func (c *WhatsmeowCallOffer) GetRelayBlock() *RelayBlock {
	c.relayBlockOnce.Do(func() { c.relayBlock = c.Data.ExtractRelayBlock() })
	return c.relayBlock
}

// HasRelayCandidatesCached returns true if any te2 candidate exists (cached lazily).
func (c *WhatsmeowCallOffer) HasRelayCandidatesCached() bool {
	c.relayCandOnce.Do(func() {
		b := c.GetRelayBlock()
		if b != nil {
			c.relayHasCandidates = len(b.TE2) > 0
			return
		}
		c.relayHasCandidates = c.Data.HasRelayCandidates()
	})
	return c.relayHasCandidates
}

// IsP2PDisabledCached returns true if voip_settings.options.disable_p2p == "1".
func (c *WhatsmeowCallOffer) IsP2PDisabledCached() bool {
	voip, _ := c.GetVoipSettings()
	if len(voip) == 0 {
		return false
	}
	optionsAny, ok := voip["options"]
	if !ok || optionsAny == nil {
		return false
	}
	options, ok := optionsAny.(map[string]interface{})
	if !ok {
		return false
	}
	v, ok := options["disable_p2p"]
	if !ok || v == nil {
		return false
	}
	return fmt.Sprint(v) == "1" || strings.ToLower(fmt.Sprint(v)) == "true"
}

// RelayNamesCached extracts unique relay_name values from relay/te2 nodes.
func (c *WhatsmeowCallOffer) RelayNamesCached() []string {
	d := c.GetData()
	relay := d.FindFirst("relay")
	if relay == nil || len(relay.Content) == 0 {
		return nil
	}
	uniq := map[string]struct{}{}
	if b := c.GetRelayBlock(); b != nil {
		for _, te := range b.TE2 {
			name := strings.TrimSpace(te.RelayName)
			if name != "" {
				uniq[name] = struct{}{}
			}
		}
	} else {
		var nodes []RawNode
		if json.Unmarshal(relay.Content, &nodes) != nil {
			return nil
		}
		for _, n := range nodes {
			if n.Tag != "te2" {
				continue
			}
			if n.Attrs == nil {
				continue
			}
			name := strings.TrimSpace(n.Attrs["relay_name"])
			if name != "" {
				uniq[name] = struct{}{}
			}
		}
	}
	if len(uniq) == 0 {
		return nil
	}
	names := make([]string, 0, len(uniq))
	for k := range uniq {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
