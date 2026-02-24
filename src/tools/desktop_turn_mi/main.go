package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pion/stun"
)

type turnExport struct {
	OnlyMsgType string           `json:"only_msg_type"`
	Exported    int              `json:"exported"`
	Items       []turnExportItem `json:"items"`
	Input       string           `json:"input"`
}

type turnExportItem struct {
	MsgType string           `json:"msg_type"`
	Len     int              `json:"len"`
	TxID    string           `json:"txid"`
	MI      bool             `json:"mi"`
	PktTS   string           `json:"pkt_ts"`
	IPLine  string           `json:"ip"`
	Attrs   []turnExportAttr `json:"attrs"`
}

type turnExportAttr struct {
	Type string `json:"type"`
	Len  int    `json:"len"`
	Hex  string `json:"hex"`
}

func main() {
	var jsonPath string
	var index int
	var list bool
	var keyHex string
	var keyB64 string
	var keyText string
	var keyFromAttr string
	var tryCommon bool
	var tryWindows bool
	var windowLens string
	var windowFrom string
	flag.StringVar(&jsonPath, "json", "", "Path to wa_desktop_*_turn_allocate.json")
	flag.IntVar(&index, "index", 0, "Item index within JSON items[]")
	flag.BoolVar(&list, "list", false, "List items and exit")
	flag.StringVar(&keyHex, "keyHex", "", "HMAC key as hex")
	flag.StringVar(&keyB64, "keyB64", "", "HMAC key as base64 (std or url, padded or raw)")
	flag.StringVar(&keyText, "keyText", "", "HMAC key as UTF-8 text")
	flag.StringVar(&keyFromAttr, "keyFromAttr", "", "Use attribute bytes as HMAC key (e.g. 0x4024, 0x4000, 0x0016)")
	flag.BoolVar(&tryCommon, "tryCommon", false, "Try a small set of common/derived keys from attrs and report hits")
	flag.BoolVar(&tryWindows, "tryWindows", false, "Try sliding-window keys inside attr blobs (for embedded secrets)")
	flag.StringVar(&windowLens, "windowLens", "16,20,32", "Comma-separated window lengths for --tryWindows")
	flag.StringVar(&windowFrom, "windowFrom", "0x4024,0x4000", "Comma-separated attr types to scan for --tryWindows")
	flag.Parse()

	if strings.TrimSpace(jsonPath) == "" {
		fmt.Fprintln(os.Stderr, "Missing --json")
		os.Exit(2)
	}

	payload, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Read JSON:", err)
		os.Exit(2)
	}
	// PowerShell may write UTF-8 with BOM; json.Unmarshal will fail unless we strip it.
	payload = bytes.TrimPrefix(payload, []byte{0xEF, 0xBB, 0xBF})

	var exp turnExport
	if err := json.Unmarshal(payload, &exp); err != nil {
		fmt.Fprintln(os.Stderr, "Parse JSON:", err)
		os.Exit(2)
	}

	if len(exp.Items) == 0 {
		fmt.Fprintln(os.Stderr, "No items found in JSON")
		os.Exit(2)
	}

	if list {
		for i, it := range exp.Items {
			endpoint := "-"
			for _, a := range it.Attrs {
				if strings.EqualFold(a.Type, "0x0016") {
					endpoint = "(has 0x0016)"
					break
				}
			}
			fmt.Printf("[%d] msg_type=%s txid=%s pkt_ts=%s len=%d mi=%v attrs=%d %s\n",
				i, it.MsgType, it.TxID, it.PktTS, it.Len, it.MI, len(it.Attrs), endpoint,
			)
		}
		return
	}

	if index < 0 || index >= len(exp.Items) {
		fmt.Fprintf(os.Stderr, "Invalid --index %d (items=%d)\n", index, len(exp.Items))
		os.Exit(2)
	}
	it := exp.Items[index]

	raw, err := buildStunRaw(it)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Build STUN raw:", err)
		os.Exit(2)
	}

	var msg stun.Message
	msg.Raw = raw
	if err := msg.Decode(); err != nil {
		fmt.Fprintln(os.Stderr, "stun Decode:", err)
		os.Exit(2)
	}

	miAttr, ok := findAttr(it, 0x0008)
	if !ok {
		fmt.Fprintln(os.Stderr, "Item has no 0x0008 MESSAGE-INTEGRITY attribute")
		os.Exit(2)
	}
	miBytes, err := hex.DecodeString(strings.TrimSpace(miAttr.Hex))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Decode MI hex:", err)
		os.Exit(2)
	}

	fmt.Printf("Item index=%d msg_type=%s txid=%s pkt_ts=%s\n", index, it.MsgType, it.TxID, it.PktTS)
	fmt.Printf("RawLen=%d HeaderLen=%d (json.len=%d)\n", len(raw), len(raw)-stunHeaderSize, it.Len)
	fmt.Printf("MI(0x0008)=%s\n", strings.ToLower(hex.EncodeToString(miBytes)))

	if tryCommon {
		hits := 0
		cands := buildCommonCandidates(it)
		for _, cand := range cands {
			checker := stun.MessageIntegrity(cand.key)
			err := checker.Check(&msg)
			if err == nil {
				hits++
				fmt.Printf("MI CHECK: OK (%s)\n", cand.label)
			}
		}
		if hits == 0 {
			fmt.Printf("MI CHECK: no hits (tried %d candidates)\n", len(cands))
			os.Exit(1)
		}
		return
	}

	if tryWindows {
		lens, err := parseIntList(windowLens)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid --windowLens:", err)
			os.Exit(2)
		}
		fromTypes, err := parseHexTypeList(windowFrom)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid --windowFrom:", err)
			os.Exit(2)
		}
		cands := buildWindowCandidates(it, fromTypes, lens, 5000)
		hits := 0
		for _, cand := range cands {
			checker := stun.MessageIntegrity(cand.key)
			if checker.Check(&msg) == nil {
				hits++
				fmt.Printf("MI CHECK: OK (%s)\n", cand.label)
			}
		}
		if hits == 0 {
			fmt.Printf("MI CHECK: no hits (tried %d window candidates)\n", len(cands))
			os.Exit(1)
		}
		return
	}

	key, keyLabel, err := parseKeyWithAttr(it, keyHex, keyB64, keyText, keyFromAttr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "Provide exactly one of: --keyHex, --keyB64, --keyText, --keyFromAttr, or use --tryCommon")
		os.Exit(2)
	}

	checker := stun.MessageIntegrity(key)
	if err := checker.Check(&msg); err != nil {
		fmt.Printf("MI CHECK: FAIL (%s) err=%v\n", keyLabel, err)
		os.Exit(1)
	}
	fmt.Printf("MI CHECK: OK (%s)\n", keyLabel)
}

func buildStunRaw(it turnExportItem) ([]byte, error) {
	msgType, err := parseHexU16(it.MsgType)
	if err != nil {
		return nil, fmt.Errorf("parse msg_type %q: %w", it.MsgType, err)
	}

	txidHex := strings.TrimSpace(it.TxID)
	txid, err := hex.DecodeString(txidHex)
	if err != nil {
		return nil, fmt.Errorf("parse txid %q: %w", it.TxID, err)
	}
	if len(txid) != 12 {
		return nil, fmt.Errorf("txid must be 12 bytes, got %d", len(txid))
	}

	attrsBuf := &bytes.Buffer{}
	for _, a := range it.Attrs {
		typ, err := parseHexU16(a.Type)
		if err != nil {
			return nil, fmt.Errorf("parse attr.type %q: %w", a.Type, err)
		}
		valHex := strings.TrimSpace(a.Hex)
		val, err := hex.DecodeString(valHex)
		if err != nil {
			return nil, fmt.Errorf("decode attr %s hex: %w", a.Type, err)
		}
		if a.Len != len(val) {
			return nil, fmt.Errorf("attr %s len mismatch: json.len=%d decoded=%d", a.Type, a.Len, len(val))
		}

		_ = binary.Write(attrsBuf, binary.BigEndian, uint16(typ))
		_ = binary.Write(attrsBuf, binary.BigEndian, uint16(len(val)))
		_, _ = attrsBuf.Write(val)

		if pad := (4 - (len(val) % 4)) % 4; pad != 0 {
			_, _ = attrsBuf.Write(make([]byte, pad))
		}
	}

	attrsRaw := attrsBuf.Bytes()
	if len(attrsRaw) > 0xFFFF {
		return nil, errors.New("attributes too large")
	}

	raw := make([]byte, 0, stunHeaderSize+len(attrsRaw))
	w := bytes.NewBuffer(raw)

	_ = binary.Write(w, binary.BigEndian, uint16(msgType))
	_ = binary.Write(w, binary.BigEndian, uint16(len(attrsRaw)))
	_ = binary.Write(w, binary.BigEndian, uint32(stunMagicCookie))
	_, _ = w.Write(txid)
	_, _ = w.Write(attrsRaw)

	out := w.Bytes()
	if got := len(out) - stunHeaderSize; got != it.Len {
		// Not fatal; Desktop export has been consistent, but keep this as a useful signal.
		// We still return the reconstructed raw bytes.
	}
	return out, nil
}

const (
	// STUN fixed header is always 20 bytes.
	stunHeaderSize = 20
	// STUN magic cookie per RFC5389.
	stunMagicCookie = 0x2112A442
)

func parseHexU16(s string) (uint16, error) {
	t := strings.TrimSpace(strings.ToLower(s))
	if strings.HasPrefix(t, "0x") {
		t = t[2:]
	}
	if t == "" {
		return 0, errors.New("empty")
	}
	u, err := strconv.ParseUint(t, 16, 16)
	if err != nil {
		return 0, err
	}
	return uint16(u), nil
}

func findAttr(it turnExportItem, typ uint16) (turnExportAttr, bool) {
	for _, a := range it.Attrs {
		u, err := parseHexU16(a.Type)
		if err != nil {
			continue
		}
		if u == typ {
			return a, true
		}
	}
	return turnExportAttr{}, false
}

func parseKey(keyHex, keyB64, keyText string) ([]byte, string, error) {
	set := 0
	if strings.TrimSpace(keyHex) != "" {
		set++
	}
	if strings.TrimSpace(keyB64) != "" {
		set++
	}
	if strings.TrimSpace(keyText) != "" {
		set++
	}
	if set != 1 {
		return nil, "", errors.New("must provide exactly one key input")
	}

	if strings.TrimSpace(keyHex) != "" {
		b, err := hex.DecodeString(strings.TrimSpace(keyHex))
		if err != nil {
			return nil, "", fmt.Errorf("invalid --keyHex: %w", err)
		}
		return b, fmt.Sprintf("keyHex(%dB)", len(b)), nil
	}

	if strings.TrimSpace(keyB64) != "" {
		b, err := decodeAnyBase64(strings.TrimSpace(keyB64))
		if err != nil {
			return nil, "", fmt.Errorf("invalid --keyB64: %w", err)
		}
		return b, fmt.Sprintf("keyB64(%dB)", len(b)), nil
	}

	// keyText
	b := []byte(keyText)
	if len(b) == 0 {
		return nil, "", errors.New("--keyText is empty")
	}
	// Keep a short label.
	sum := sha1.Sum(b)
	label := fmt.Sprintf("keyText(%dB,sha1=%s)", len(b), hex.EncodeToString(sum[:4]))
	return b, label, nil
}

func parseKeyWithAttr(it turnExportItem, keyHex, keyB64, keyText, keyFromAttr string) ([]byte, string, error) {
	set := 0
	if strings.TrimSpace(keyHex) != "" {
		set++
	}
	if strings.TrimSpace(keyB64) != "" {
		set++
	}
	if strings.TrimSpace(keyText) != "" {
		set++
	}
	if strings.TrimSpace(keyFromAttr) != "" {
		set++
	}
	if set != 1 {
		return nil, "", errors.New("must provide exactly one key input")
	}
	if strings.TrimSpace(keyFromAttr) == "" {
		return parseKey(keyHex, keyB64, keyText)
	}
	attrType, err := parseHexU16(keyFromAttr)
	if err != nil {
		return nil, "", fmt.Errorf("invalid --keyFromAttr %q: %w", keyFromAttr, err)
	}
	attr, ok := findAttr(it, attrType)
	if !ok {
		return nil, "", fmt.Errorf("attr %s not found in item", strings.ToLower(keyFromAttr))
	}
	b, err := hex.DecodeString(strings.TrimSpace(attr.Hex))
	if err != nil {
		return nil, "", fmt.Errorf("decode attr %s hex: %w", keyFromAttr, err)
	}
	label := fmt.Sprintf("keyFromAttr(%s,%dB)", fmt.Sprintf("0x%04x", attrType), len(b))
	return b, label, nil
}

type candidateKey struct {
	label string
	key   []byte
}

func buildCommonCandidates(it turnExportItem) []candidateKey {
	var cands []candidateKey
	add := func(label string, key []byte) {
		if len(key) == 0 {
			return
		}
		// Defensive copy
		b := make([]byte, len(key))
		copy(b, key)
		cands = append(cands, candidateKey{label: label, key: b})
	}

	// Prefer call/session-level blob first.
	var b4024, b4000, b0016, b4002, b0020 []byte
	if a, ok := findAttr(it, 0x4024); ok {
		b4024, _ = hex.DecodeString(strings.TrimSpace(a.Hex))
		add("attr4024(raw)", b4024)
		s1 := sha1.Sum(b4024)
		add("attr4024(sha1)", s1[:])
		s256 := sha256.Sum256(b4024)
		add("attr4024(sha256)", s256[:])
		if len(b4024) >= 20 {
			add("attr4024(first20)", b4024[:20])
			add("attr4024(last20)", b4024[len(b4024)-20:])
		}
	}
	if a, ok := findAttr(it, 0x4000); ok {
		b4000, _ = hex.DecodeString(strings.TrimSpace(a.Hex))
		add("attr4000(raw)", b4000)
		s1 := sha1.Sum(b4000)
		add("attr4000(sha1)", s1[:])
		s256 := sha256.Sum256(b4000)
		add("attr4000(sha256)", s256[:])
		if len(b4000) >= 20 {
			add("attr4000(first20)", b4000[:20])
			add("attr4000(last20)", b4000[len(b4000)-20:])
		}
	}
	if a, ok := findAttr(it, 0x0016); ok {
		b0016, _ = hex.DecodeString(strings.TrimSpace(a.Hex))
		add("attr0016(raw)", b0016)
		s1 := sha1.Sum(b0016)
		add("attr0016(sha1)", s1[:])
	}
	if a, ok := findAttr(it, 0x4002); ok {
		b4002, _ = hex.DecodeString(strings.TrimSpace(a.Hex))
		add("attr4002(raw)", b4002)
		s1 := sha1.Sum(b4002)
		add("attr4002(sha1)", s1[:])
	}
	if a, ok := findAttr(it, 0x0020); ok {
		b0020, _ = hex.DecodeString(strings.TrimSpace(a.Hex))
		add("attr0020(raw)", b0020)
		s1 := sha1.Sum(b0020)
		add("attr0020(sha1)", s1[:])
	}

	// A few concatenations/KDFs that are cheap to test.
	if len(b4024) > 0 && len(b0016) > 0 {
		combo := append(append([]byte{}, b4024...), b0016...)
		add("attr4024+0016", combo)
		s1 := sha1.Sum(combo)
		add("sha1(attr4024+0016)", s1[:])
	}
	if len(b4024) > 0 && len(b4000) > 0 {
		combo := append(append([]byte{}, b4024...), b4000...)
		add("attr4024+4000", combo)
		s1 := sha1.Sum(combo)
		add("sha1(attr4024+4000)", s1[:])
	}

	// HMAC-based derived keys: HMAC-SHA1(key=A, msg=B) using only on-wire blobs.
	h1 := func(key, msg []byte) []byte {
		if len(key) == 0 || len(msg) == 0 {
			return nil
		}
		h := hmac.New(sha1.New, key)
		_, _ = h.Write(msg)
		return h.Sum(nil)
	}
	if len(b4000) > 0 && len(b4024) > 0 {
		add("hmac1(key=4000,msg=4024)", h1(b4000, b4024))
		add("hmac1(key=4024,msg=4000)", h1(b4024, b4000))
	}
	if len(b4000) > 0 && len(b0016) > 0 {
		add("hmac1(key=4000,msg=0016)", h1(b4000, b0016))
		add("hmac1(key=0016,msg=4000)", h1(b0016, b4000))
	}
	if len(b4024) > 0 && len(b0016) > 0 {
		add("hmac1(key=4024,msg=0016)", h1(b4024, b0016))
		add("hmac1(key=0016,msg=4024)", h1(b0016, b4024))
	}
	if len(b4002) > 0 && len(b4000) > 0 {
		add("hmac1(key=4000,msg=4002)", h1(b4000, b4002))
		add("hmac1(key=4002,msg=4000)", h1(b4002, b4000))
	}
	if len(b4002) > 0 && len(b4024) > 0 {
		add("hmac1(key=4024,msg=4002)", h1(b4024, b4002))
		add("hmac1(key=4002,msg=4024)", h1(b4002, b4024))
	}
	if len(b0020) > 0 && len(b4000) > 0 {
		add("hmac1(key=4000,msg=0020)", h1(b4000, b0020))
		add("hmac1(key=0020,msg=4000)", h1(b0020, b4000))
	}

	return cands
}

func parseIntList(s string) ([]int, error) {
	var out []int
	for _, part := range strings.Split(s, ",") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		if n <= 0 {
			return nil, fmt.Errorf("non-positive length: %d", n)
		}
		out = append(out, n)
	}
	if len(out) == 0 {
		return nil, errors.New("empty list")
	}
	return out, nil
}

func parseHexTypeList(s string) ([]uint16, error) {
	var out []uint16
	for _, part := range strings.Split(s, ",") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		u, err := parseHexU16(p)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	if len(out) == 0 {
		return nil, errors.New("empty list")
	}
	return out, nil
}

func buildWindowCandidates(it turnExportItem, fromTypes []uint16, lens []int, max int) []candidateKey {
	var cands []candidateKey
	add := func(label string, key []byte) {
		if len(cands) >= max {
			return
		}
		b := make([]byte, len(key))
		copy(b, key)
		cands = append(cands, candidateKey{label: label, key: b})
	}

	for _, t := range fromTypes {
		attr, ok := findAttr(it, t)
		if !ok {
			continue
		}
		blob, err := hex.DecodeString(strings.TrimSpace(attr.Hex))
		if err != nil {
			continue
		}
		for _, l := range lens {
			if l > len(blob) {
				continue
			}
			for i := 0; i+l <= len(blob); i++ {
				add(fmt.Sprintf("win(%s)[%d:%d]", fmt.Sprintf("0x%04x", t), i, i+l), blob[i:i+l])
				if len(cands) >= max {
					return cands
				}
			}
		}
	}
	return cands
}

func decodeAnyBase64(s string) ([]byte, error) {
	// Small helper to accept both std and url alphabets and optional padding.
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil, errors.New("empty")
	}

	// Try common decoders in order.
	decoders := []struct {
		name string
		enc  *base64.Encoding
	}{
		{"RawStd", base64.RawStdEncoding},
		{"RawURL", base64.RawURLEncoding},
		{"Std", base64.StdEncoding},
		{"URL", base64.URLEncoding},
	}

	var lastErr error
	for _, d := range decoders {
		b, err := d.enc.DecodeString(trimmed)
		if err == nil {
			return b, nil
		}
		lastErr = err
	}

	// If input is unpadded, try padding for Std/URL.
	if !strings.Contains(trimmed, "=") {
		pad := (4 - (len(trimmed) % 4)) % 4
		if pad != 0 {
			padded := trimmed + strings.Repeat("=", pad)
			for _, d := range []struct {
				name string
				enc  *base64.Encoding
			}{
				{"Std(padded)", base64.StdEncoding},
				{"URL(padded)", base64.URLEncoding},
			} {
				b, err := d.enc.DecodeString(padded)
				if err == nil {
					return b, nil
				}
				lastErr = err
			}
		}
	}

	if lastErr == nil {
		lastErr = errors.New("decode failed")
	}
	return nil, lastErr
}
