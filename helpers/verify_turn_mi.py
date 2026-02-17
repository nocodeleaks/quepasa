#!/usr/bin/env python3
"""Verify TURN/STUN MESSAGE-INTEGRITY (HMAC-SHA1) for a specific request in a PCAP.

This is a debugging utility to prove whether the on-wire MESSAGE-INTEGRITY value
matches a locally-derived key for a given txid.

Usage:
  python helpers/verify_turn_mi.py <file.pcap> --txid <hex12bytes> --key-hex <hex>

Notes:
- Supports PCAP linktype Ethernet (1) and Linux cooked capture v2 (SLL2, 276).
- Only checks Allocate (method=0x0003) request messages.
"""

from __future__ import annotations

import argparse
import hmac
import struct
from pathlib import Path
from typing import Iterable, List, Optional, Tuple

MAGIC_COOKIE = 0x2112A442

ATTR_USERNAME = 0x0006
ATTR_MESSAGE_INTEGRITY = 0x0008


def _u16be(b: bytes) -> int:
    return struct.unpack(">H", b)[0]


def _u32be(b: bytes) -> int:
    return struct.unpack(">I", b)[0]


def _read_pcap_global_header(f) -> Tuple[str, int]:
    hdr = f.read(24)
    if len(hdr) != 24:
        raise ValueError("PCAP too short")

    magic = hdr[:4]
    if magic == b"\xd4\xc3\xb2\xa1":
        endian = "<"
    elif magic == b"\xa1\xb2\xc3\xd4":
        endian = ">"
    elif magic == b"\x4d\x3c\xb2\xa1":
        endian = "<"
    elif magic == b"\xa1\xb2\x3c\x4d":
        endian = ">"
    else:
        raise ValueError(f"Unknown PCAP magic: {magic.hex()}")

    _ver_major, _ver_minor, _thiszone, _sigfigs, _snaplen, network = struct.unpack(
        endian + "HHiiii", hdr[4:24]
    )
    return endian, network


def _iter_pcap_packets(f, endian: str) -> Iterable[bytes]:
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


def _decode_sll2(packet: bytes) -> Tuple[int, bytes]:
    if len(packet) < 20:
        raise ValueError("SLL2 packet too short")
    protocol = _u16be(packet[0:2])
    return protocol, packet[20:]


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
    if b[0] & 0xC0:
        return False
    return _u32be(b[4:8]) == MAGIC_COOKIE


def _stun_class(msg_type: int) -> int:
    c0 = (msg_type >> 4) & 0x1
    c1 = (msg_type >> 8) & 0x1
    return (c1 << 1) | c0


def _stun_method(msg_type: int) -> int:
    m0_3 = msg_type & 0x000F
    m4_6 = (msg_type >> 1) & 0x0070
    m7_11 = (msg_type >> 2) & 0x0F80
    return m0_3 | m4_6 | m7_11


def _nearest_padded_value_len(n: int) -> int:
    return (n + 3) & ~3


class ParsedStun:
    def __init__(self, raw: bytes):
        self.raw = raw
        self.msg_type = _u16be(raw[0:2])
        self.msg_len = _u16be(raw[2:4])
        self.txid = raw[8:20]
        self.attrs: List[Tuple[int, int, int, int]] = []
        # Each attr: (atype, alen, start_offset, padded_total_len_with_header)

        end = 20 + self.msg_len
        if end > len(raw):
            end = len(raw)

        off = 20
        while off + 4 <= end:
            at = _u16be(raw[off : off + 2])
            alen = _u16be(raw[off + 2 : off + 4])
            val_start = off + 4
            padded = _nearest_padded_value_len(alen)
            total = 4 + padded
            self.attrs.append((at, alen, off, total))
            off = val_start + padded

    def get_attr_value(self, atype: int) -> Optional[bytes]:
        for at, alen, off, _total in self.attrs:
            if at != atype:
                continue
            val_start = off + 4
            return self.raw[val_start : val_start + alen]
        return None


def verify_message_integrity(req: ParsedStun, key: bytes) -> Tuple[bool, str]:
    mi = req.get_attr_value(ATTR_MESSAGE_INTEGRITY)
    if mi is None:
        return False, "no MESSAGE-INTEGRITY attr"
    if len(mi) != 20:
        return False, f"MI wrong size: {len(mi)}"

    # Find MI attribute index.
    mi_index = None
    for i, (at, _alen, _off, _total) in enumerate(req.attrs):
        if at == ATTR_MESSAGE_INTEGRITY:
            mi_index = i
            break
    if mi_index is None:
        return False, "MI not found in attrs list"

    # Compute new length = original_length - bytes(after MI)
    bytes_after = 0
    for (at, _alen, _off, total) in req.attrs[mi_index + 1 :]:
        bytes_after += total

    new_len = req.msg_len - bytes_after
    if new_len < 0:
        return False, f"invalid new_len: {new_len}"

    # startOfHMAC is the first byte of the MI attribute header.
    start_of_hmac = 20 + new_len - (4 + 20)
    if start_of_hmac < 20 or start_of_hmac > len(req.raw):
        return False, f"invalid start_of_hmac: {start_of_hmac}"

    raw2 = bytearray(req.raw)
    raw2[2:4] = struct.pack(">H", new_len)

    expected = hmac.new(key, bytes(raw2[:start_of_hmac]), "sha1").digest()
    ok = hmac.compare_digest(mi, expected)
    detail = f"new_len={new_len} start={start_of_hmac} mi={mi.hex()} exp={expected.hex()}"
    return ok, detail


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("pcap")
    ap.add_argument("--txid", required=True, help="STUN txid as 24 hex chars (12 bytes)")
    ap.add_argument("--key-hex", required=True, help="HMAC key in hex")
    args = ap.parse_args()

    want_txid = bytes.fromhex(args.txid)
    key = bytes.fromhex(args.key_hex)

    data = Path(args.pcap).read_bytes()
    import io

    f = io.BytesIO(data)
    endian, linktype = _read_pcap_global_header(f)

    found_req = None
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

        if ethertype != 0x0800:
            continue
        dec = _decode_ipv4(l3)
        if not dec:
            continue
        proto, ip_payload = dec
        if proto != 17:
            continue
        udp_payload = _decode_udp(ip_payload)
        if not udp_payload or not _is_stun(udp_payload):
            continue

        msg_type = _u16be(udp_payload[0:2])
        cls = _stun_class(msg_type)
        method = _stun_method(msg_type)
        if method != 0x0003 or cls != 0:  # Allocate request only
            continue
        txid = udp_payload[8:20]
        if txid != want_txid:
            continue
        found_req = ParsedStun(udp_payload)
        break

    if not found_req:
        print("not found")
        return 1

    user = found_req.get_attr_value(ATTR_USERNAME)
    print(f"pcap={args.pcap} linktype={linktype}")
    print(f"txid={want_txid.hex()} user={(user.hex() if user else '-')}")
    ok, detail = verify_message_integrity(found_req, key)
    print("ok=" + str(ok))
    print(detail)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
