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
	"encoding/json"
	"fmt"
	"reflect"
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
	relayCandOnce      sync.Once
	relayHasCandidates bool
	dataOnce           sync.Once
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
			if cf.Type().Elem().Kind() == reflect.Uint8 { // []byte -> treat as base64? we just as string
				b, _ := json.Marshal(string(cf.Bytes()))
				rn.Content = b
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
	relay := o.FindFirst("relay")
	if relay == nil || len(relay.Content) == 0 {
		return nil
	}
	var nodes []RawNode
	if json.Unmarshal(relay.Content, &nodes) != nil {
		return nil
	}
	tokens := make([]string, 0, 3)
	for _, n := range nodes {
		if n.Tag == "token" && len(n.Content) > 0 {
			var s string
			if json.Unmarshal(n.Content, &s) == nil && s != "" {
				tokens = append(tokens, s)
			}
		}
	}
	return tokens
}

// Relay candidate presence
func (o *OfferDataNode) HasRelayCandidates() bool {
	relay := o.FindFirst("relay")
	if relay == nil || len(relay.Content) == 0 {
		return false
	}
	var nodes []RawNode
	if json.Unmarshal(relay.Content, &nodes) != nil {
		return false
	}
	for _, n := range nodes {
		if n.Tag == "te2" {
			return true
		}
	}
	return false
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
	c.relayOnce.Do(func() { c.relayTokens = c.Data.ExtractRelayTokens() })
	return c.relayTokens
}

// HasRelayCandidatesCached returns true if any te2 candidate exists (cached lazily).
func (c *WhatsmeowCallOffer) HasRelayCandidatesCached() bool {
	c.relayCandOnce.Do(func() { c.relayHasCandidates = c.Data.HasRelayCandidates() })
	return c.relayHasCandidates
}
