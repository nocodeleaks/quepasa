package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	wabinary "go.mau.fi/whatsmeow/binary"
)

var reResult = regexp.MustCompile(`crypto\.result\.(decrypt|encrypt) (\{.*\})`)
var reDirect = regexp.MustCompile(`crypto\.(direct|proto)\.(decrypt|encrypt) (\{.*\})`)

type cryptoLine struct {
	Len     int    `json:"len"`
	HeadHex string `json:"head_hex"`
	FullHex string `json:"full_hex"`
}

type directLine struct {
	Method string     `json:"method"`
	Data   cryptoLine `json:"data"`
}

func decodeHex(s string) (*wabinary.Node, error) {
	data, err := hex.DecodeString(strings.TrimSpace(s))
	if err != nil {
		return nil, err
	}
	if len(data) < 2 {
		return nil, fmt.Errorf("too short")
	}
	return wabinary.Unmarshal(data[1:])
}

func main() {
	path := ".dist\\wa_cdp_page_console.log"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("open_error:", err)
		os.Exit(1)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		if m := reResult.FindStringSubmatch(line); m != nil {
			kind := "result." + m[1]
			if m[1] != "decrypt" {
				continue
			}
			var item cryptoLine
			if err := json.Unmarshal([]byte(m[2]), &item); err != nil {
				continue
			}
			hexStr := item.FullHex
			if hexStr == "" && item.Len > 0 && len(item.HeadHex) == item.Len*2 {
				hexStr = item.HeadHex
			}
			if hexStr == "" {
				continue
			}
			node, err := decodeHex(hexStr)
			if err != nil {
				fmt.Printf("line=%d kind=%s len=%d decode_err=%v\n", lineNo, kind, item.Len, err)
				continue
			}
			typ, _ := node.Attrs["type"]
			id, _ := node.Attrs["id"]
			from, _ := node.Attrs["from"]
			to, _ := node.Attrs["to"]
			fmt.Printf("line=%d kind=%s len=%d tag=%s type=%v id=%v from=%v to=%v xml=%s\n",
				lineNo, kind, item.Len, node.Tag, typ, id, from, to, node.XMLString())
			continue
		}
		if m := reDirect.FindStringSubmatch(line); m != nil {
			kind := m[1] + "." + m[2]
			var item directLine
			if err := json.Unmarshal([]byte(m[3]), &item); err != nil {
				continue
			}
			hexStr := item.Data.FullHex
			if hexStr == "" && item.Data.Len > 0 && len(item.Data.HeadHex) == item.Data.Len*2 {
				hexStr = item.Data.HeadHex
			}
			if hexStr == "" {
				continue
			}
			node, err := decodeHex(hexStr)
			if err != nil {
				fmt.Printf("line=%d kind=%s len=%d decode_err=%v\n", lineNo, kind, item.Data.Len, err)
				continue
			}
			typ, _ := node.Attrs["type"]
			id, _ := node.Attrs["id"]
			from, _ := node.Attrs["from"]
			to, _ := node.Attrs["to"]
			fmt.Printf("line=%d kind=%s len=%d tag=%s type=%v id=%v from=%v to=%v xml=%s\n",
				lineNo, kind, item.Data.Len, node.Tag, typ, id, from, to, node.XMLString())
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Println("scan_error:", err)
		os.Exit(1)
	}
}
