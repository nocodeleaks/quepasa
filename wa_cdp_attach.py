import json
import pathlib
import sys
import time
import urllib.request
from typing import Dict, Optional

try:
    import websocket
except Exception:
    print("missing dependency: websocket-client")
    print("install with: python -m pip install websocket-client")
    sys.exit(1)


ROOT = pathlib.Path(__file__).resolve().parent
PAYLOAD_PATH = ROOT / "wa_cdp_inject.js"
CDP_BASE = "http://127.0.0.1:9222"


def http_json(path: str):
    with urllib.request.urlopen(CDP_BASE + path, timeout=3) as resp:
        return json.load(resp)


def get_browser_ws_url() -> str:
    version = http_json("/json/version")
    ws_url = version.get("webSocketDebuggerUrl")
    if not ws_url:
      raise RuntimeError("browser websocket debugger URL not found")
    return str(ws_url)


def is_relevant_target_info(info: dict) -> bool:
    title = str(info.get("title", "")).lower()
    url = str(info.get("url", "")).lower()
    ttype = str(info.get("type", "")).lower()
    if ttype not in ("page", "worker", "service_worker", "shared_worker"):
        return False
    if "whatsapp" in title or "whatsapp" in url:
        return True
    if ttype in ("worker", "service_worker", "shared_worker"):
        return True
    return False


def render_console_args(args):
    rendered = []
    for a in args:
        if "value" in a:
            rendered.append(str(a["value"]))
        elif "description" in a:
            rendered.append(str(a["description"]))
    return " ".join(rendered)


class BrowserCDP:
    def __init__(self, ws_url: str):
        self.ws = websocket.create_connection(ws_url, timeout=10, suppress_origin=True)
        self.next_id = 1
        self.pending: Dict[int, dict] = {}
        self.session_meta: Dict[str, dict] = {}
        self.payload = PAYLOAD_PATH.read_text(encoding="utf-8")

    def send(self, method: str, params: Optional[dict] = None, session_id: Optional[str] = None) -> int:
        msg_id = self.next_id
        self.next_id += 1
        payload = {"id": msg_id, "method": method}
        if params is not None:
            payload["params"] = params
        if session_id is not None:
            payload["sessionId"] = session_id
        self.ws.send(json.dumps(payload))
        return msg_id

    def call(self, method: str, params: Optional[dict] = None, session_id: Optional[str] = None, timeout: float = 10.0):
        msg_id = self.send(method, params=params, session_id=session_id)
        deadline = time.time() + timeout
        while time.time() < deadline:
            msg = self.recv(timeout=max(0.1, deadline - time.time()))
            if msg.get("id") == msg_id:
                return msg
        raise TimeoutError(f"timeout waiting for {method}")

    def maybe_log_console(self, msg: dict):
        if msg.get("method") != "Runtime.consoleAPICalled":
            return
        params = msg.get("params", {}) or {}
        session_id = msg.get("sessionId")
        meta = self.session_meta.get(session_id or "", {})
        label = f"{meta.get('type', '?')}:{meta.get('title', '') or meta.get('targetId', '')}"
        rendered = render_console_args(params.get("args", []) or [])
        if rendered:
            print(f"[console:{label}] {rendered}")

    def maybe_log_network(self, msg: dict):
        method = msg.get("method")
        if method not in (
            "Network.webSocketCreated",
            "Network.webSocketFrameSent",
            "Network.webSocketFrameReceived",
            "Network.webSocketClosed",
        ):
            return
        session_id = msg.get("sessionId")
        meta = self.session_meta.get(session_id or "", {})
        label = f"{meta.get('type', '?')}:{meta.get('title', '') or meta.get('targetId', '')}"
        params = msg.get("params", {}) or {}
        if method == "Network.webSocketCreated":
            print(
                f"[cdp:{label}] WebSocketCreated",
                json.dumps(
                    {
                        "requestId": params.get("requestId"),
                        "url": params.get("url"),
                    },
                    ensure_ascii=False,
                ),
            )
            return
        if method == "Network.webSocketClosed":
            print(
                f"[cdp:{label}] WebSocketClosed",
                json.dumps(
                    {
                        "requestId": params.get("requestId"),
                        "timestamp": params.get("timestamp"),
                    },
                    ensure_ascii=False,
                ),
            )
            return
        response = params.get("response", {}) or {}
        payload = str(response.get("payloadData", ""))
        info = {
            "requestId": params.get("requestId"),
            "opcode": response.get("opcode"),
            "mask": response.get("mask"),
            "payload_len": len(payload),
            "payload_head": payload[:80],
        }
        print(
            f"[cdp:{label}] {method.split('.')[-1]}",
            json.dumps(info, ensure_ascii=False),
        )

    def bootstrap_session(self, session_id: str, target_info: dict):
        self.session_meta[session_id] = {
            "targetId": target_info.get("targetId"),
            "type": target_info.get("type"),
            "title": target_info.get("title"),
            "url": target_info.get("url"),
        }
        ttype = str(target_info.get("type", "")).lower()
        label = f"{ttype} {target_info.get('title', '')} {target_info.get('url', '')}".strip()
        session_timeout = 30 if ttype == "page" else 12
        try:
            try:
                self.call("Runtime.enable", session_id=session_id, timeout=session_timeout)
            except Exception:
                pass
            try:
                self.call("Network.enable", session_id=session_id, timeout=session_timeout)
            except Exception:
                pass
            if ttype == "page":
                try:
                    self.call("Page.enable", session_id=session_id, timeout=session_timeout)
                except Exception:
                    pass
                try:
                    self.call(
                        "Page.addScriptToEvaluateOnNewDocument",
                        {"source": self.payload},
                        session_id=session_id,
                        timeout=session_timeout,
                    )
                except Exception:
                    pass
            try:
                self.call(
                    "Runtime.evaluate",
                    {"expression": self.payload, "awaitPromise": False},
                    session_id=session_id,
                    timeout=session_timeout,
                )
            except Exception:
                try:
                    self.send(
                        "Runtime.evaluate",
                        {"expression": self.payload, "awaitPromise": False},
                        session_id=session_id,
                    )
                except Exception:
                    pass
            print("attached:", label)
        except Exception as exc:
            print("attach-failed:", label, str(exc))

    def recv(self, timeout: float = 1.0):
        self.ws.settimeout(timeout)
        raw = self.ws.recv()
        msg = json.loads(raw)
        self.maybe_log_console(msg)
        self.maybe_log_network(msg)
        if msg.get("method") == "Target.attachedToTarget":
            params = msg.get("params", {}) or {}
            session_id = params.get("sessionId")
            target_info = params.get("targetInfo", {}) or {}
            if session_id and is_relevant_target_info(target_info):
                self.bootstrap_session(session_id, target_info)
        elif msg.get("method") == "Target.targetCreated":
            params = msg.get("params", {}) or {}
            target_info = params.get("targetInfo", {}) or {}
            if is_relevant_target_info(target_info):
                print(
                    "target-created:",
                    target_info.get("type", ""),
                    target_info.get("title", ""),
                    target_info.get("url", ""),
                )
        return msg


def main():
    ws_url = None
    boot_reported = False
    while ws_url is None:
        try:
            ws_url = get_browser_ws_url()
        except Exception as exc:
            if not boot_reported:
                print("waiting for browser CDP on 127.0.0.1:9222:", str(exc))
                boot_reported = True
            time.sleep(1)

    cdp = BrowserCDP(ws_url)

    cdp.call("Target.setDiscoverTargets", {"discover": True}, timeout=10)
    cdp.call(
        "Target.setAutoAttach",
        {
            "autoAttach": True,
            "waitForDebuggerOnStart": False,
            "flatten": True,
            "filter": [
                {"type": "page", "exclude": False},
                {"type": "worker", "exclude": False},
                {"type": "service_worker", "exclude": False},
                {"type": "shared_worker", "exclude": False},
            ],
        },
        timeout=10,
    )

    try:
        infos = http_json("/json/list")
    except Exception:
        infos = []
    for info in infos:
        if not is_relevant_target_info(info):
            continue
        try:
            resp = cdp.call(
                "Target.attachToTarget",
                {"targetId": info.get("id"), "flatten": True},
                timeout=10,
            )
            session_id = (resp.get("result", {}) or {}).get("sessionId")
            if session_id:
                cdp.bootstrap_session(
                    session_id,
                    {
                        "targetId": info.get("id"),
                        "type": info.get("type"),
                        "title": info.get("title"),
                        "url": info.get("url"),
                    },
                )
        except Exception as exc:
            print("attach-failed:", info.get("type", ""), info.get("title", ""), str(exc))

    print("browser-level auto-attach active; waiting for console events")
    while True:
        try:
            cdp.recv(timeout=1.0)
        except Exception:
            time.sleep(0.25)


if __name__ == "__main__":
    main()
