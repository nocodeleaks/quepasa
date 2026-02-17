package whatsmeow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types/events"
)

type callOfferDump struct {
	Kind         string                 `json:"kind"`
	Captured     string                 `json:"captured"`
	CallID       string                 `json:"call_id"`
	From         string                 `json:"from"`
	Timestamp    string                 `json:"timestamp,omitempty"`
	Attrs        map[string]string      `json:"attrs,omitempty"`
	Data         *OfferDataNode         `json:"data"`
	VoipSettings map[string]interface{} `json:"voip_settings,omitempty"`
	RelayTokens  []string               `json:"relay_tokens,omitempty"`
}

func DumpCallOfferEvent(evt *events.CallOffer, normalized *WhatsmeowCallOffer) (string, error) {
	if evt == nil {
		return "", fmt.Errorf("nil CallOffer")
	}

	dumpDir := strings.TrimSpace(os.Getenv("QP_CALL_DUMP_DIR"))
	if dumpDir == "" {
		dumpDir = filepath.Join(".dist", "call_dumps")
	}
	if err := os.MkdirAll(dumpDir, 0o755); err != nil {
		return "", err
	}

	callIDPart := sanitizeFilenamePart(evt.CallID)
	if callIDPart == "" {
		callIDPart = "unknown"
	}
	timestampStr := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("call_offer_received_%s_%s.json", timestampStr, callIDPart)
	path := filepath.Join(dumpDir, filename)

	timestamp := ""
	if !evt.Timestamp.IsZero() {
		timestamp = evt.Timestamp.UTC().Format(time.RFC3339Nano)
	}

	data := &OfferDataNode{Tag: "offer", Attrs: map[string]string{}, Content: nil}
	voip := map[string]interface{}(nil)
	relay := []string(nil)
	if normalized != nil {
		data = normalized.GetData()
		if v, _ := normalized.GetVoipSettings(); len(v) > 0 {
			voip = v
		}
		relay = normalized.GetRelayTokens()
	}

	attrs := map[string]string{}
	for k, v := range data.Attrs {
		attrs[k] = v
	}

	dump := callOfferDump{
		Kind:         "CallOffer",
		Captured:     time.Now().UTC().Format(time.RFC3339Nano),
		CallID:       evt.CallID,
		From:         fmt.Sprint(evt.From),
		Timestamp:    timestamp,
		Attrs:        attrs,
		Data:         data,
		VoipSettings: voip,
		RelayTokens:  relay,
	}

	bytes, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, bytes, 0o644); err != nil {
		return "", err
	}

	return path, nil
}
