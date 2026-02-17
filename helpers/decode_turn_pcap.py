#!/usr/bin/env python3
"""Decode STUN/TURN messages from a .pcap without external dependencies.

Features:
- Supports libpcap files with linktype Ethernet (1) and Linux cooked capture v2 (SLL2, 276).
- Extracts UDP payloads and identifies STUN messages via RFC 5389 magic cookie.
- Prints per-transaction (txid) summary for TURN Allocate (method=0x0003).

Usage:
  python helpers/decode_turn_pcap.py .dist/pcaps/wa_turn_YYYYMMDD_HHMMSS.pcap

Notes:
- This is intentionally minimal and best-effort.
- It does not validate MESSAGE-INTEGRITY; it only reports presence and related attrs.
"""

from __future__ import annotations

import struct
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, Iterable, List, Optional, Tuple


MAGIC_COOKIE = 0x2112A442


@dataclass
class StunMsg:
    msg_type: int
    msg_len: int
    cookie: int
    txid: bytes
    attrs: List[Tuple[int, bytes]]


def _read_pcap_global_header(f) -> Tuple[str, int, int]:
    hdr = f.read(24)
    if len(hdr) != 24:
        raise ValueError("PCAP too short (missing global header)")

    magic = hdr[:4]
    if magic == b"\xd4\xc3\xb2\xa1":
        endian = "<"
    elif magic == b"\xa1\xb2\xc3\xd4":
        endian = ">"
    elif magic == b"\x4d\x3c\xb2\xa1":
        endian = "<"  # nanosecond-resolution (rare)
    elif magic == b"\xa1\xb2\x3c\x4d":
        endian = ">"
    else:
        raise ValueError(f"Unknown PCAP magic: {magic.hex()}")

    _ver_major, _ver_minor, _thiszone, _sigfigs, _snaplen, network = struct.unpack(
        endian + "HHiiii", hdr[4:24]
    )
    return endian, _snaplen, network


def _iter_pcap_packets(f, endian: str) -> Iterable[bytes]:
    # per-packet header: ts_sec, ts_usec, incl_len, orig_len
    ph_fmt = endian + "IIII"
    while True:
        ph = f.read(16)
        if not ph:
            return
        if len(ph) != 16:
            return
        _ts_sec, _ts_sub, incl_len, _orig_len = struct.unpack(ph_fmt, ph)
        data = f.read(incl_len)
        if len(data) != incl_len:
            return
        yield data


def _u16be(b: bytes) -> int:
    return struct.unpack(">H", b)[0]


def _u32be(b: bytes) -> int:
    return struct.unpack(">I", b)[0]


def _decode_sll2(packet: bytes) -> Tuple[int, bytes]:
    # Linux cooked capture v2 (SLL2), 20 bytes header.
    # https://www.tcpdump.org/linktypes/LINKTYPE_LINUX_SLL2.html
    if len(packet) < 20:
        raise ValueError("SLL2 packet too short")
    protocol = _u16be(packet[0:2])
    payload = packet[20:]
    return protocol, payload


def _decode_ethernet(packet: bytes) -> Tuple[int, bytes]:
    if len(packet) < 14:
        raise ValueError("Ethernet frame too short")
    ethertype = _u16be(packet[12:14])
    return ethertype, packet[14:]


def _decode_ipv4(payload: bytes) -> Optional[Tuple[int, bytes]]:
    if len(payload) < 20:
        return None
    ver_ihl = payload[0]
    ver = (ver_ihl >> 4) & 0xF
    if ver != 4:
        return None
    ihl = (ver_ihl & 0xF) * 4
    if ihl < 20 or len(payload) < ihl:
        return None
    proto = payload[9]
    total_len = _u16be(payload[2:4])
    if total_len < ihl:
        return None
    total_len = min(total_len, len(payload))
    return proto, payload[ihl:total_len]


def _decode_udp(ip_payload: bytes) -> Optional[bytes]:
    if len(ip_payload) < 8:
        return None
    udp_len = _u16be(ip_payload[4:6])
    if udp_len < 8:
        return None
    udp_len = min(udp_len, len(ip_payload))
    return ip_payload[8:udp_len]


def _is_stun(b: bytes) -> bool:
    if len(b) < 20:
        return False
    # STUN message: first two bits are 0
    if b[0] & 0xC0:
        return False
    return _u32be(b[4:8]) == MAGIC_COOKIE


def _parse_stun(b: bytes) -> Optional[StunMsg]:
    if not _is_stun(b):
        return None
    msg_type = _u16be(b[0:2])
    msg_len = _u16be(b[2:4])
    cookie = _u32be(b[4:8])
    txid = b[8:20]
    end = 20 + msg_len
    if end > len(b):
        end = len(b)
    attrs: List[Tuple[int, bytes]] = []
    off = 20
    while off + 4 <= end:
        at = _u16be(b[off : off + 2])
        alen = _u16be(b[off + 2 : off + 4])
        off += 4
        aval = b[off : min(off + alen, end)]
        attrs.append((at, aval))
        off += alen
        # padding to 32-bit
        pad = (-alen) % 4
        off += pad
    return StunMsg(msg_type=msg_type, msg_len=msg_len, cookie=cookie, txid=txid, attrs=attrs)


def _stun_class(msg_type: int) -> int:
    # STUN class bits (RFC 5389):
    # class = (C1<<1) | C0 where C0=bit4, C1=bit8
    c0 = (msg_type >> 4) & 0x1
    c1 = (msg_type >> 8) & 0x1
    return (c1 << 1) | c0


def _stun_method(msg_type: int) -> int:
    # method = M11..M0 reconstructed from bits
    m0_3 = msg_type & 0x000F
    m4_6 = (msg_type >> 1) & 0x0070
    m7_11 = (msg_type >> 2) & 0x0F80
    return m0_3 | m4_6 | m7_11


def _attr_get(attrs: List[Tuple[int, bytes]], atype: int) -> List[bytes]:
    return [v for (t, v) in attrs if t == atype]


def _decode_error_code(v: bytes) -> Optional[Tuple[int, str]]:
    if len(v) < 4:
        return None
    # 2 reserved bytes, then class (3 bits), number (8 bits)
    code_class = v[2] & 0x07
    code_number = v[3]
    code = code_class * 100 + code_number
    reason = ""
    if len(v) > 4:
        try:
            reason = v[4:].decode("utf-8", errors="replace").strip()
        except Exception:
            reason = ""
    return code, reason


def _b2hex(b: bytes) -> str:
    return b.hex()


def _attr_types_hex(attrs: List[Tuple[int, bytes]]) -> str:
    if not attrs:
        return ""
    types = [f"0x{t:04x}" for (t, _v) in attrs]
    if len(types) > 12:
        types = types[:12] + ["..."]
    return "[" + ",".join(types) + "]"


def _format_username(v: bytes) -> str:
    if not v:
        return ""
    try:
        s = v.decode("utf-8")
        if s and all(32 <= ord(ch) <= 126 for ch in s):
            return s
    except Exception:
        pass

    hx = v.hex()
    if len(hx) > 64:
        hx = hx[:64] + "..."
    return f"hex:{hx}"


def main() -> int:
    if len(sys.argv) != 2:
        print("Usage: python helpers/decode_turn_pcap.py <file.pcap>")
        return 2

    p = Path(sys.argv[1])
    data = p.read_bytes()
    # Use a file object for simplicity.
    import io

    f = io.BytesIO(data)
    endian, _snaplen, linktype = _read_pcap_global_header(f)

    stun_msgs: List[StunMsg] = []
    for pkt in _iter_pcap_packets(f, endian):
        try:
            if linktype == 1:
                ethertype, l3 = _decode_ethernet(pkt)
            elif linktype == 276:
                protocol, l3 = _decode_sll2(pkt)
                ethertype = protocol
            else:
                continue
        except Exception:
            continue

        # IPv4 ethertype 0x0800
        if ethertype != 0x0800:
            continue
        decoded = _decode_ipv4(l3)
        if not decoded:
            continue
        proto, ip_payload = decoded
        if proto != 17:  # UDP
            continue
        udp_payload = _decode_udp(ip_payload)
        if not udp_payload:
            continue
        msg = _parse_stun(udp_payload)
        if msg:
            stun_msgs.append(msg)

    print(f"pcap={p} linktype={linktype} stun_msgs={len(stun_msgs)}")

    # Group by txid
    by_txid: Dict[bytes, List[StunMsg]] = {}
    for m in stun_msgs:
        by_txid.setdefault(m.txid, []).append(m)

    # Focus on TURN Allocate (method 0x0003)
    alloc_method = 0x0003
    allocs: List[Tuple[str, int, int, Optional[int], str, bool, bool, Optional[str], int, int, str]] = []
    for txid, msgs in by_txid.items():
        for m in msgs:
            method = _stun_method(m.msg_type)
            if method != alloc_method:
                continue
            cls = _stun_class(m.msg_type)
            # username (0x0006)
            usernames = _attr_get(m.attrs, 0x0006)
            username = ""
            if usernames:
                username = _format_username(usernames[0])
            # error-code (0x0009)
            ec = None
            reason = ""
            ecs = _attr_get(m.attrs, 0x0009)
            if ecs:
                dec = _decode_error_code(ecs[0])
                if dec:
                    ec, reason = dec
            # realm (0x0014), nonce (0x0015)
            realm = ""
            nonce = ""
            realms = _attr_get(m.attrs, 0x0014)
            nonces = _attr_get(m.attrs, 0x0015)
            if realms:
                realm = realms[0].decode("utf-8", errors="replace")
            if nonces:
                nonce = nonces[0].decode("utf-8", errors="replace")
            realm_len = len(realms[0]) if realms else 0
            nonce_len = len(nonces[0]) if nonces else 0
            attr_types = _attr_types_hex(m.attrs)
            # presence flags
            has_mi = bool(_attr_get(m.attrs, 0x0008))
            has_fp = bool(_attr_get(m.attrs, 0x8028))

            allocs.append(
                (
                    _b2hex(txid),
                    cls,
                    m.msg_type,
                    ec,
                    reason,
                    has_mi,
                    has_fp,
                    username if username else None,
                    realm_len,
                    nonce_len,
                    attr_types,
                )
            )

    if not allocs:
        print("No TURN Allocate messages found (method=0x0003)")
        return 0

    # Print a compact list.
    # class: 0=request, 1=indication, 2=success, 3=error
    class_name = {0: "req", 1: "ind", 2: "ok", 3: "err"}
    print(f"allocate_msgs={len(allocs)}")
    for txid_hex, cls, msg_type, ec, reason, has_mi, has_fp, username, realm_len, nonce_len, attr_types in allocs[:80]:
        cn = class_name.get(cls, str(cls))
        u = username if username is not None else "-"
        e = f"{ec}" if ec is not None else "-"
        r = reason.replace("\n", " ")[:80] if reason else ""
        extra = ""
        if cn == "err":
            extra = f" realmLen={realm_len} nonceLen={nonce_len} attrs={attr_types}"
        print(f"txid={txid_hex} class={cn} type=0x{msg_type:04x} err={e} mi={int(has_mi)} fp={int(has_fp)} user={u} reason={r}{extra}")

    if len(allocs) > 80:
        print(f"... truncated ({len(allocs)} total allocate msgs)")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
