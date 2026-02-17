#!/usr/bin/env bash
set -euo pipefail

# Capture UDP traffic relevant to WhatsApp/WebRTC call setup on apoint-voip.
# - Focuses on STUN/TURN (3478/5349) and typical UDP media ports.
# - Produces a .pcap and a .log.
#
# Usage examples:
#   ./capture-apoint-voip-pcap.sh --duration 180 --callid A519... \
#     --filter "udp and (port 3478 or port 5349 or portrange 10000-65000)"
#
#   # Narrow to a specific relay from dumps:
#   ./capture-apoint-voip-pcap.sh --duration 180 --callid A519... \
#     --filter "udp and host 57.144.179.54 and port 3478"

DURATION=180
IFACE=eth0
OUTDIR=/opt/quepasa/.dist/pcaps
PREFIX=webrtc
CALLID=""
FILTER="udp and (port 3478 or port 5349 or portrange 10000-65000)"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --duration)
      DURATION="$2"; shift 2 ;;
    --iface)
      IFACE="$2"; shift 2 ;;
    --outdir)
      OUTDIR="$2"; shift 2 ;;
    --prefix)
      PREFIX="$2"; shift 2 ;;
    --callid)
      CALLID="$2"; shift 2 ;;
    --filter)
      FILTER="$2"; shift 2 ;;
    -h|--help)
      sed -n '1,120p' "$0"; exit 0 ;;
    *)
      echo "Unknown arg: $1" >&2
      exit 2
      ;;
  esac
done

mkdir -p "$OUTDIR"

TS=$(date +%Y%m%d_%H%M%S)
SAFE_CALLID=$(printf '%s' "$CALLID" | tr -cd 'A-Za-z0-9_-')
SUFFIX=""
if [[ -n "$SAFE_CALLID" ]]; then
  SUFFIX="_${SAFE_CALLID}"
fi

PCAP="$OUTDIR/${PREFIX}${SUFFIX}_${TS}.pcap"
LOG="$OUTDIR/${PREFIX}${SUFFIX}_${TS}.log"

echo "[CAPTURE] iface=$IFACE duration=${DURATION}s"
echo "[CAPTURE] outdir=$OUTDIR"
echo "[CAPTURE] filter=$FILTER"
echo "[CAPTURE] pcap=$PCAP"
echo "[CAPTURE] log=$LOG"

# timeout returns 124 when it stops the command; that's expected.
set +e
timeout "$DURATION" tcpdump -i "$IFACE" -nn -s0 -w "$PCAP" "$FILTER" >"$LOG" 2>&1
CODE=$?
set -e

if [[ $CODE -ne 0 && $CODE -ne 124 ]]; then
  echo "[CAPTURE] tcpdump failed (exit=$CODE). See $LOG" >&2
  exit $CODE
fi

ls -lh "$PCAP" "$LOG"
echo "PCAP=$PCAP"
echo "LOG=$LOG"
