package whatsmeow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type callTurnProbeAttemptResult struct {
	Index   int    `json:"index"`
	TxID    string `json:"txid,omitempty"`
	Cand    string `json:"cand"`
	Algo    string `json:"algo"`
	UserLen int    `json:"user_len"`
	KeyLen  int    `json:"key_len"`

	Code     int    `json:"code"`
	Reason   string `json:"reason,omitempty"`
	NonceLen int    `json:"nonce_len"`
	RealmLen int    `json:"realm_len"`

	RespMsgType    string `json:"resp_msg_type,omitempty"`
	MappedEndpoint string `json:"mapped_endpoint,omitempty"`
	Extra4002Hex   string `json:"extra_4002_hex,omitempty"`

	Success  bool `json:"success"`
	LongTerm bool `json:"long_term"`
	TryMI256 bool `json:"try_mi_sha256"`

	RelayName string `json:"relay_name,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
}

type callTurnProbeRequestAttr struct {
	Type string `json:"type"`
	Len  int    `json:"len"`
	Hex  string `json:"hex"`
}

type callTurnProbeRequestDump struct {
	Stage       string                     `json:"stage"`
	TxID        string                     `json:"txid,omitempty"`
	MsgType     string                     `json:"msg_type,omitempty"`
	Len         int                        `json:"len"`
	PreimageHex string                     `json:"preimage_hex,omitempty"`
	RawHex      string                     `json:"raw_hex,omitempty"`
	Attrs       []callTurnProbeRequestAttr `json:"attrs,omitempty"`
}

type callTurnProbeDump struct {
	Kind     string `json:"kind"`
	Captured string `json:"captured"`
	CallID   string `json:"call_id"`

	RelayName string `json:"relay_name,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
	LocalAddr string `json:"local_addr,omitempty"`

	BaseAllocateTxID    string `json:"base_allocate_txid,omitempty"`
	BaseAllocateSuccess bool   `json:"base_allocate_success"`
	BaseRespMsgType     string `json:"base_resp_msg_type,omitempty"`
	BaseMappedEndpoint  string `json:"base_mapped_endpoint,omitempty"`
	BaseExtra4002Hex    string `json:"base_extra_4002_hex,omitempty"`

	BaseAllocateCode   int    `json:"base_allocate_code"`
	BaseAllocateReason string `json:"base_allocate_reason,omitempty"`
	BaseNonceLen       int    `json:"base_nonce_len"`
	BaseRealmLen       int    `json:"base_realm_len"`

	DiscoveryTxID     string `json:"discovery_txid,omitempty"`
	DiscoveryUser     string `json:"discovery_user,omitempty"`
	DiscoveryCode     int    `json:"discovery_code"`
	DiscoveryReason   string `json:"discovery_reason,omitempty"`
	DiscoveryNonceLen int    `json:"discovery_nonce_len"`
	DiscoveryRealmLen int    `json:"discovery_realm_len"`

	RelayUUID string `json:"relay_uuid,omitempty"`
	SelfPID   string `json:"self_pid,omitempty"`
	PeerPID   string `json:"peer_pid,omitempty"`
	HasKey    bool   `json:"has_key"`
	HasHBHKey bool   `json:"has_hbh_key"`

	Buckets map[string]int `json:"buckets,omitempty"`
	MaxTry  int            `json:"max_try"`

	Requests []callTurnProbeRequestDump `json:"requests,omitempty"`

	Attempts []callTurnProbeAttemptResult `json:"attempts,omitempty"`
}

func DumpCallTurnProbeSummary(callID string, dump callTurnProbeDump) (string, error) {
	callIDPart := sanitizeFilenamePart(callID)
	if callIDPart == "" {
		callIDPart = "unknown"
	}

	dumpDir := strings.TrimSpace(os.Getenv("QP_CALL_DUMP_DIR"))
	if dumpDir == "" {
		dumpDir = filepath.Join(".dist", "call_dumps")
	}
	if err := os.MkdirAll(dumpDir, 0o755); err != nil {
		return "", err
	}

	timestampStr := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("call_turn_probe_%s_%s.json", timestampStr, callIDPart)
	path := filepath.Join(dumpDir, filename)

	if strings.TrimSpace(dump.Kind) == "" {
		dump.Kind = "TurnProbe"
	}
	if strings.TrimSpace(dump.Captured) == "" {
		dump.Captured = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if strings.TrimSpace(dump.CallID) == "" {
		dump.CallID = callID
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
