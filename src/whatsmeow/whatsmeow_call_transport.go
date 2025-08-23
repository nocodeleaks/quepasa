package whatsmeow

// Structures and helpers to decode the CallTransport "Data" payload
// derived from the sample JSON files in docs/WHATSMEOW-call_transport_example.pretty.json.
//
// The root Data object itself has the shape:
// {
//   "Tag": "transport",
//   "Attrs": { "call-id": "...", ... },
//   "Content": [ Node, Node, ... ]
// }
// Each Node has: Tag, Attrs (map) and Content which can be:
//  - null
//  - string (e.g. base64 tokens, etc.)
//  - another array of Nodes
//
// We keep flexibility by using json.RawMessage in Content and providing helpers
// to interpret specific nodes.

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow/types/events"
)

// WhatsmeowCallTransport embeds the original events.CallTransport and adds a normalized Data tree.
type WhatsmeowCallTransport struct {
	TimestampRaw string `json:"Timestamp,omitempty"`
	
	events.CallTransport
	Data TransportDataNode `json:"Data"`
	dataOnce sync.Once
}

type TransportDataNode struct {
	Tag     string            `json:"Tag"`
	Attrs   map[string]string `json:"Attrs"`
	Content []RawNode         `json:"Content"`
}

type RawNode struct {
	Tag     string            `json:"Tag"`
	Attrs   map[string]string `json:"Attrs"`
	Content json.RawMessage   `json:"Content"`
}

func NewWhatsmeowCallTransport(evt *events.CallTransport) *WhatsmeowCallTransport {
	w := &WhatsmeowCallTransport{}
	if evt != nil {
		w.CallTransport = *evt
	}
	// Derive TimestampRaw
	if f := reflect.ValueOf(&w.CallTransport).Elem().FieldByName("Timestamp"); f.IsValid() && f.CanInterface() {
		if t, ok := f.Interface().(time.Time); ok && !t.IsZero() {
			w.TimestampRaw = t.UTC().Format(time.RFC3339)
		}
	}
	if w.Data.Tag == "" {
		w.Data.Tag = "transport"
	}
	if w.Data.Attrs == nil {
		w.Data.Attrs = map[string]string{}
	}
	if w.CallID != "" {
		w.Data.Attrs["call-id"] = w.CallID
	}
	// Attempt to copy CallCreator via reflection
	rv := reflect.ValueOf(&w.CallTransport).Elem()
	if f := rv.FieldByName("CallCreator"); f.IsValid() && f.Kind() == reflect.String && f.String() != "" {
		w.Data.Attrs["call-creator"] = f.String()
	}
	w.hydrateFromEventData()
	return w
}

func (w *WhatsmeowCallTransport) hydrateFromEventData() {
	if len(w.Data.Content) > 0 {
		return
	}
	rv := reflect.ValueOf(&w.CallTransport).Elem()
	f := rv.FieldByName("Data")
	if !f.IsValid() || f.IsZero() {
		return
	}
	fval := f
	if fval.Kind() == reflect.Ptr && !fval.IsNil() {
		fval = fval.Elem()
	}
	rootNode, ok := toRawNode(fval)
	if !ok {
		return
	}
	if len(rootNode.Attrs) > 0 {
		for k, v := range rootNode.Attrs {
			if _, exists := w.Data.Attrs[k]; !exists {
				w.Data.Attrs[k] = v
			}
		}
	}
	if len(rootNode.Content) > 0 {
		var children []RawNode
		if err := json.Unmarshal(rootNode.Content, &children); err == nil {
			w.Data.Content = children
		}
	}
}

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
	contentField := v.FieldByName("Content")
	if contentField.IsValid() && contentField.Kind() != reflect.Invalid && !contentField.IsZero() {
		cf := contentField
		if cf.Kind() == reflect.Interface && !cf.IsNil() { cf = cf.Elem() }
		switch cf.Kind() {
		case reflect.String:
			b, _ := json.Marshal(cf.String())
			rn.Content = b
		case reflect.Slice, reflect.Array:
			if cf.Type().Elem().Kind() == reflect.Uint8 {
				b, _ := json.Marshal(string(cf.Bytes()))
				rn.Content = b
			} else {
				children := make([]RawNode, 0, cf.Len())
				for i := 0; i < cf.Len(); i++ {
					childVal := cf.Index(i)
					if cn, okc := toRawNode(childVal); okc { children = append(children, cn) }
				}
				if len(children) > 0 { if b, err := json.Marshal(children); err == nil { rn.Content = b } }
			}
		case reflect.Struct:
			if cn, okc := toRawNode(cf); okc {
				children := []RawNode{cn}
				if b, err := json.Marshal(children); err == nil { rn.Content = b }
			}
		default:
			b, _ := json.Marshal(fmt.Sprint(cf.Interface()))
			rn.Content = b
		}
	}
	return rn, true
}

// GetData returns a pointer to the normalized TransportDataNode ensuring hydration once.
func (c *WhatsmeowCallTransport) GetData() *TransportDataNode {
	c.dataOnce.Do(func() {
		if len(c.Data.Content) == 0 { c.hydrateFromEventData() }
	})
	return &c.Data
}

// IsValid tells if the transport is structurally valid and not expired (TTL 90s).
func (c *WhatsmeowCallTransport) IsValid() bool {
	if c == nil { return false }
	if c.CallID == "" { return false }
	d := c.GetData()
	if strings.ToLower(d.Tag) != "transport" { return false }
	const ttl = 90 * time.Second
	t := time.Time{}
	rv := reflect.ValueOf(&c.CallTransport).Elem()
	if f := rv.FieldByName("Timestamp"); f.IsValid() && f.CanInterface() { if tt, ok := f.Interface().(time.Time); ok { t = tt } }
	if !t.IsZero() {
		if time.Since(t) > ttl { return false }
	}
	return true
}

func (c *WhatsmeowCallTransport) Summary() string {
	d := c.GetData()
	return fmt.Sprintf("callid=%s ts=%s", c.CallID, c.TimestampRaw)
}
