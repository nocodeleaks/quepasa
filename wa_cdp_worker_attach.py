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
PAYLOAD_PATH = ROOT / "wa_cdp_inject_worker.js"
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
    ttype = str(info.get("type", "")).lower()
    if ttype not in ("worker", "service_worker", "shared_worker"):
        return False
    title = str(info.get("title", "")).lower()
    url = str(info.get("url", "")).lower()
    if url.startswith("chrome-extension://"):
        return False
    if not title and not url:
        return True
    if "whatsapp" in title or "whatsapp" in url:
        return True
    if url.startswith("https://web.whatsapp.com"):
        return True
    if url.startswith("https://static.whatsapp.net"):
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

    def bootstrap_session(self, session_id: str, target_info: dict):
        self.session_meta[session_id] = {
            "targetId": target_info.get("targetId"),
            "type": target_info.get("type"),
            "title": target_info.get("title"),
            "url": target_info.get("url"),
        }
        label = f"{target_info.get('type', '')}:{target_info.get('title', '')} {target_info.get('url', '')}".strip()
        try:
            try:
                self.call("Runtime.enable", session_id=session_id, timeout=15)
            except Exception:
                pass
            try:
                self.call("Runtime.evaluate", {"expression": self.payload, "awaitPromise": False}, session_id=session_id, timeout=15)
            except Exception:
                try:
                    self.send("Runtime.evaluate", {"expression": self.payload, "awaitPromise": False}, session_id=session_id)
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
        if msg.get("method") == "Target.attachedToTarget":
            params = msg.get("params", {}) or {}
            session_id = params.get("sessionId")
            target_info = params.get("targetInfo", {}) or {}
            if session_id and is_relevant_target_info(target_info):
                self.bootstrap_session(session_id, target_info)
        return msg


def main():
    ws_url = None
    while ws_url is None:
        try:
            ws_url = get_browser_ws_url()
        except Exception as exc:
            print("waiting for browser CDP on 127.0.0.1:9222:", str(exc))
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
            resp = cdp.call("Target.attachToTarget", {"targetId": info.get("id"), "flatten": True}, timeout=10)
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

    print("worker-level auto-attach active; waiting for console events")
    while True:
        try:
            cdp.recv(timeout=1.0)
        except Exception:
            time.sleep(0.25)


if __name__ == "__main__":
    main()
