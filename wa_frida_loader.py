import os
import sys
import threading
import time

import frida

SCRIPT_PATH = os.path.join(os.getcwd(), "wa_sendto_stun.js")
WATCH_NAMES = {"WhatsApp.Root.exe", "msedgewebview2.exe"}
WATCH_MARKERS = (
    "WhatsApp.Root.exe",
    "WhatsAppDesktop",
    "--webview-exe-name=WhatsApp.Root.exe",
)
POLL_SECONDS = 1.0


class Loader:
    def __init__(self, device, script_source):
        self.device = device
        self.script_source = script_source
        self.sessions = {}
        self.scripts = {}
        self.lock = threading.Lock()

    def on_message(self, pid, name):
        def _handler(message, data):
            if message["type"] == "send":
                print(f"[{name}:{pid}] {message.get('payload')}")
            elif message["type"] == "error":
                print(f"[{name}:{pid}] SCRIPT-ERROR {message.get('stack') or message}", file=sys.stderr)
            else:
                print(f"[{name}:{pid}] {message}")
        return _handler

    def attach_and_load(self, proc):
        pid = proc.pid
        name = getattr(proc, "name", "proc")
        with self.lock:
            if pid in self.sessions:
                return
        try:
            session = self.device.attach(pid)
            script = session.create_script(self.script_source)
            script.on("message", self.on_message(pid, name))
            script.load()
            with self.lock:
                self.sessions[pid] = session
                self.scripts[pid] = script
            print(f"[attach] {name} pid={pid} ok")
        except Exception as exc:
            print(f"[attach] {name} pid={pid} failed: {exc}", file=sys.stderr)

    def scan_once(self):
        try:
            processes = self.device.enumerate_processes()
        except Exception as exc:
            print(f"[scan] enumerate_processes failed: {exc}", file=sys.stderr)
            return
        for proc in processes:
            name = getattr(proc, "name", "") or ""
            params = getattr(proc, "parameters", None) or {}
            identifier = str(params.get("identifier", "") or "")
            path = str(params.get("path", "") or "")
            cmdline = str(params.get("argv", "") or "")
            haystack = " | ".join([name, identifier, path, cmdline])
            if name in WATCH_NAMES or any(marker in haystack for marker in WATCH_MARKERS):
                self.attach_and_load(proc)


def main():
    if not os.path.exists(SCRIPT_PATH):
        print(f"script not found: {SCRIPT_PATH}", file=sys.stderr)
        sys.exit(1)

    with open(SCRIPT_PATH, "r", encoding="utf-8") as f:
        script_source = f.read()

    device = frida.get_local_device()
    loader = Loader(device, script_source)

    print("[ready] polling for WhatsApp/Desktop WebView processes")
    print("[ready] start or keep WhatsApp Desktop open, then make/answer a call")
    print("[ready] press Ctrl+C to stop")

    try:
        while True:
            loader.scan_once()
            time.sleep(POLL_SECONDS)
    except KeyboardInterrupt:
        print("[stop]")


if __name__ == "__main__":
    main()
