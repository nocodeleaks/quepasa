package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	wabinary "go.mau.fi/whatsmeow/binary"
)

func tryDecode(label, hexStr string) {
	data, err := hex.DecodeString(strings.TrimSpace(hexStr))
	if err != nil {
		fmt.Printf("%s hex_error=%v\n", label, err)
		return
	}
	for _, off := range []int{0, 1, 2, 3} {
		if off >= len(data) {
			continue
		}
		node, err := wabinary.Unmarshal(data[off:])
		if err != nil {
			fmt.Printf("%s off=%d err=%v\n", label, off, err)
			continue
		}
		fmt.Printf("%s off=%d tag=%q attrs=%v content_type=%T xml=%s\n", label, off, node.Tag, node.Attrs, node.Content, node.XMLString())
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run helpers/decode-wa-cdp-plaintext.go <hex> [<hex>...]")
		os.Exit(1)
	}
	for i, arg := range os.Args[1:] {
		tryDecode(fmt.Sprintf("arg%d", i+1), arg)
	}
}
