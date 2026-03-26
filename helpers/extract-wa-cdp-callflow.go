package main

import (
    "bufio"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "os"
    "regexp"
    "sort"
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

type event struct {
    Line    int               `json:"line"`
    Kind    string            `json:"kind"`
    Tag     string            `json:"tag"`
    Type    string            `json:"type,omitempty"`
    ID      string            `json:"id,omitempty"`
    From    string            `json:"from,omitempty"`
    To      string            `json:"to,omitempty"`
    CallID  string            `json:"call_id,omitempty"`
    XML     string            `json:"xml"`
    Attrs   map[string]string `json:"attrs,omitempty"`
}

type group struct {
    CallID string  `json:"call_id"`
    Events []event `json:"events"`
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

func flattenCallID(n *wabinary.Node) string {
    if n == nil {
        return ""
    }
    if v, ok := n.Attrs["call-id"]; ok {
        return fmt.Sprint(v)
    }
    switch c := n.Content.(type) {
    case []wabinary.Node:
        for i := range c {
            if id := flattenCallID(&c[i]); id != "" {
                return id
            }
        }
    }
    return ""
}

func attrsToStrings(m wabinary.Attrs) map[string]string {
    if len(m) == 0 {
        return nil
    }
    out := make(map[string]string, len(m))
    for k, v := range m {
        out[k] = fmt.Sprint(v)
    }
    return out
}

func isCallRelevant(n *wabinary.Node) bool {
    if n == nil {
        return false
    }
    if n.Tag == "call" || n.Tag == "receipt" || n.Tag == "ack" {
        if n.Tag != "ack" {
            return true
        }
        if cls, ok := n.Attrs["class"]; ok && fmt.Sprint(cls) == "call" {
            return true
        }
    }
    switch c := n.Content.(type) {
    case []wabinary.Node:
        for i := range c {
            if isCallRelevant(&c[i]) {
                return true
            }
        }
    }
    return false
}

func extractEvent(lineNo int, kind string, n *wabinary.Node) *event {
    if !isCallRelevant(n) {
        return nil
    }
    ev := &event{
        Line:   lineNo,
        Kind:   kind,
        Tag:    n.Tag,
        XML:    n.XMLString(),
        Attrs:  attrsToStrings(n.Attrs),
        CallID: flattenCallID(n),
    }
    if v, ok := n.Attrs["type"]; ok {
        ev.Type = fmt.Sprint(v)
    }
    if v, ok := n.Attrs["id"]; ok {
        ev.ID = fmt.Sprint(v)
    }
    if v, ok := n.Attrs["from"]; ok {
        ev.From = fmt.Sprint(v)
    }
    if v, ok := n.Attrs["to"]; ok {
        ev.To = fmt.Sprint(v)
    }
    return ev
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

    byCall := map[string][]event{}
    noCall := []event{}

    sc := bufio.NewScanner(f)
    sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
    lineNo := 0
    for sc.Scan() {
        lineNo++
        line := sc.Text()

        var node *wabinary.Node
        var kind string

        if m := reResult.FindStringSubmatch(line); m != nil {
            kind = "result." + m[1]
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
            n, err := decodeHex(hexStr)
            if err != nil {
                continue
            }
            node = n
        } else if m := reDirect.FindStringSubmatch(line); m != nil {
            kind = m[1] + "." + m[2]
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
            n, err := decodeHex(hexStr)
            if err != nil {
                continue
            }
            node = n
        } else {
            continue
        }

        ev := extractEvent(lineNo, kind, node)
        if ev == nil {
            continue
        }
        if ev.CallID == "" {
            noCall = append(noCall, *ev)
        } else {
            byCall[ev.CallID] = append(byCall[ev.CallID], *ev)
        }
    }
    if err := sc.Err(); err != nil {
        fmt.Println("scan_error:", err)
        os.Exit(1)
    }

    callIDs := make([]string, 0, len(byCall))
    for k := range byCall {
        callIDs = append(callIDs, k)
    }
    sort.Strings(callIDs)

    out := struct {
        Calls  []group `json:"calls"`
        NoCall []event `json:"no_call,omitempty"`
    }{}

    for _, id := range callIDs {
        out.Calls = append(out.Calls, group{CallID: id, Events: byCall[id]})
    }
    if len(noCall) > 0 {
        out.NoCall = noCall
    }

    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    _ = enc.Encode(out)
}
