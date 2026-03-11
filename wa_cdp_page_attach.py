import json
import pathlib
import sys
import time
import urllib.request
import base64

try:
    import websocket
except Exception:
    print("missing dependency: websocket-client")
    print("install with: python -m pip install websocket-client")
    sys.exit(1)


ROOT = pathlib.Path(__file__).resolve().parent
PAYLOAD_PATH = ROOT / "wa_cdp_inject.js"
CDP_BASE = "http://127.0.0.1:9222"


def decode_binary_payload(payload):
    try:
        raw_bytes = base64.b64decode(payload, validate=False)
    except Exception:
        return None

    info = {
        "bin_len": len(raw_bytes),
        "bin_head_hex": raw_bytes[:48].hex(),
    }

    if len(raw_bytes) >= 3:
        len_hdr = (raw_bytes[0] << 16) | (raw_bytes[1] << 8) | raw_bytes[2]
        info["prefix3_hex"] = raw_bytes[:3].hex()
        info["len_hdr"] = len_hdr
        info["len_hdr_matches"] = len_hdr == (len(raw_bytes) - 3)

    if len(raw_bytes) <= 256:
        info["bin_full_hex"] = raw_bytes.hex()
    else:
        info["bin_tail_hex"] = raw_bytes[-48:].hex()

    return info


def load_targets():
    with urllib.request.urlopen(CDP_BASE + "/json/list", timeout=3) as resp:
        return json.load(resp)


def pick_page_target(targets):
    scored = []
    for t in targets:
        title = str(t.get("title", ""))
        url = str(t.get("url", ""))
        ttype = str(t.get("type", ""))
        dbg = str(t.get("webSocketDebuggerUrl", ""))
        if ttype != "page" or not dbg:
            continue
        score = 0
        if "whatsapp" in title.lower():
            score += 20
        if "whatsapp" in url.lower():
            score += 20
        score += 5
        scored.append((score, t))
    scored.sort(key=lambda x: x[0], reverse=True)
    return scored[0][1] if scored else None


class CDP:
    def __init__(self, ws_url, label, timeout=15):
        self.ws = websocket.create_connection(ws_url, timeout=timeout, suppress_origin=True)
        self.next_id = 1
        self.label = label

    def send(self, method, params=None):
        msg_id = self.next_id
        self.next_id += 1
        payload = {"id": msg_id, "method": method}
        if params is not None:
            payload["params"] = params
        self.ws.send(json.dumps(payload))
        return msg_id

    def recv_until(self, wanted_id=None, timeout=30):
        deadline = time.time() + timeout
        while time.time() < deadline:
            self.ws.settimeout(max(0.1, deadline - time.time()))
            raw = self.ws.recv()
            msg = json.loads(raw)
            if wanted_id is None:
                self._handle_event(msg)
                return msg
            if msg.get("id") == wanted_id:
                return msg
            self._handle_event(msg)
        raise TimeoutError("timeout waiting for response")

    def _handle_event(self, msg):
        method = msg.get("method")
        if method == "Runtime.consoleAPICalled":
            args = msg.get("params", {}).get("args", [])
            rendered = []
            for a in args:
                if "value" in a:
                    rendered.append(str(a["value"]))
                elif "description" in a:
                    rendered.append(str(a["description"]))
            if rendered:
                print(f"[console:{self.label}]", " ".join(rendered))
        elif method in ("Network.webSocketFrameSent", "Network.webSocketFrameReceived"):
            params = msg.get("params", {}) or {}
            resp = params.get("response", {}) or {}
            payload = str(resp.get("payloadData", ""))
            info = {
                "opcode": resp.get("opcode"),
                "payload_len": len(payload),
                "payload_head": payload[:80],
            }
            if resp.get("opcode") == 2 and payload:
                decoded = decode_binary_payload(payload)
                if decoded:
                    info.update(decoded)
            print(f"[cdp:{self.label}] {method.split('.')[-1]} {json.dumps(info, ensure_ascii=False)}")
        elif method == "Network.webSocketCreated":
            params = msg.get("params", {}) or {}
            print(f"[cdp:{self.label}] WebSocketCreated {json.dumps({'url': params.get('url')}, ensure_ascii=False)}")


def main():
    payload = PAYLOAD_PATH.read_text(encoding="utf-8")

    target = None
    while target is None:
        try:
            target = pick_page_target(load_targets())
        except Exception as exc:
            print("waiting for page target:", str(exc))
        if target is None:
            time.sleep(1)

    label = f"page:{target.get('title', '')}"
    cdp = CDP(target["webSocketDebuggerUrl"], label)
    print("target:", target.get("title", ""), target.get("url", ""))

    for method, params in (
        ("Runtime.enable", None),
        ("Page.enable", None),
        ("Network.enable", None),
        ("Page.addScriptToEvaluateOnNewDocument", {"source": payload}),
        ("Runtime.evaluate", {"expression": payload, "awaitPromise": False}),
    ):
        try:
            cdp.recv_until(cdp.send(method, params), timeout=30)
        except Exception as exc:
            print("setup-failed:", method, str(exc))

    print("page injection installed; waiting for events")
    while True:
        try:
            cdp.recv_until(timeout=2)
        except Exception:
            time.sleep(0.25)


if __name__ == "__main__":
    main()
