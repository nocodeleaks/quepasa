import json
import sys
from collections import Counter, defaultdict


def main() -> int:
    if len(sys.argv) < 2:
        print("Usage: python helpers/summarize-desktop-stun-json.py <path-to-stun-export.json>")
        return 2

    path = sys.argv[1]
    raw = open(path, "rb").read()
    # Strip UTF-8 BOM if present.
    if raw.startswith(b"\xef\xbb\xbf"):
        raw = raw[3:]
    data = json.loads(raw.decode("utf-8", errors="replace"))

    items = data.get("items") or []
    print(f"items={len(items)}")

    by_type = Counter()
    attr_by_type = defaultdict(Counter)  # msg_type -> Counter(attr_type)

    for it in items:
        mt = str(it.get("msg_type") or "-")
        by_type[mt] += 1
        for a in it.get("attrs") or []:
            at = str(a.get("type") or "-")
            attr_by_type[mt][at] += 1

    print("--- msg_type counts ---")
    for mt, n in by_type.most_common():
        print(f"{mt}: {n}")

    interesting_attrs = {
        "0x0006": "USERNAME",
        "0x0014": "REALM",
        "0x0015": "NONCE",
        "0x0009": "ERROR-CODE",
        "0x0020": "XOR-MAPPED-ADDRESS",
        "0x0016": "XOR-ADDRESS(?)",
        "0x0008": "MESSAGE-INTEGRITY",
        "0x8028": "FINGERPRINT",
        "0x4000": "PROPRIETARY_4000",
        "0x4024": "PROPRIETARY_4024",
    }

    print("--- interesting attrs by msg_type (presence) ---")
    for mt, _ in by_type.most_common():
        c = attr_by_type.get(mt) or Counter()
        present = []
        for at, label in interesting_attrs.items():
            if c.get(at):
                present.append(f"{at}({label})={c[at]}")
        if present:
            print(f"{mt}: " + ", ".join(present))

    # Show top attrs for the most common msg types.
    print("--- top attrs (first 5 msg types) ---")
    for mt, _ in by_type.most_common(5):
        c = attr_by_type.get(mt) or Counter()
        top = ", ".join([f"{k}={v}" for k, v in c.most_common(20)])
        print(f"{mt}: {top}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
