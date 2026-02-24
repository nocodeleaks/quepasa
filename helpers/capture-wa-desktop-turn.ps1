param(
  [int]$MaxSeconds = 90,
  [string]$OutDir = ".dist/pcaps",
  [switch]$NoTshark,
  [switch]$Include443,
  [int[]]$ExtraPorts = @(),
  [switch]$NoFilters,
  [switch]$AutoStop,
  [switch]$SkipProcess,
  [int]$ProcessMaxPackets = 800
)

$ErrorActionPreference = "Stop"

function Require-Admin {
  $currentIdentity = [Security.Principal.WindowsIdentity]::GetCurrent()
  $principal = New-Object Security.Principal.WindowsPrincipal($currentIdentity)
  if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw "This script must run in an elevated PowerShell (Run as Administrator) because pktmon requires admin."
  }
}

function Require-Command([string]$Name) {
  $cmd = Get-Command $Name -ErrorAction SilentlyContinue
  if (-not $cmd) {
    throw "Missing required command '$Name'."
  }
  return $cmd
}

Require-Admin
Require-Command pktmon | Out-Null

$repoRoot = Split-Path -Parent $PSScriptRoot
$fullOutDir = Join-Path $repoRoot $OutDir
New-Item -ItemType Directory -Path $fullOutDir -Force | Out-Null

$stamp = Get-Date -Format "yyyyMMdd_HHmmss"
$etlPath = Join-Path $fullOutDir "wa_desktop_$stamp.etl"
$pcapPath = Join-Path $fullOutDir "wa_desktop_$stamp.pcapng"

Write-Host "[1/4] Starting pktmon capture..." -ForegroundColor Cyan
Write-Host "- Output ETL  : $etlPath" -ForegroundColor DarkGray
Write-Host "- MaxSeconds  : $MaxSeconds" -ForegroundColor DarkGray

# Clean up any prior filters (best-effort)
try { pktmon filter remove | Out-Null } catch { }

if (-not $NoFilters) {
  # Default: reduce noise. WhatsApp TURN/STUN is typically UDP/3478; signaling/media may also use 443 (QUIC/TCP).
  pktmon filter add -p 3478 | Out-Null
  if ($Include443) {
    pktmon filter add -p 443 | Out-Null
  }
  foreach ($p in $ExtraPorts) {
    if ($p -gt 0 -and $p -lt 65536) {
      pktmon filter add -p $p | Out-Null
    }
  }

  Write-Host "[INFO] Active pktmon filters:" -ForegroundColor DarkCyan
  try { pktmon filter list | Out-Host } catch { }
} else {
  Write-Host "[INFO] NoFilters set: capture will be broad/noisy." -ForegroundColor Yellow
}

pktmon start --capture --pkt-size 0 --file-name "$etlPath" | Out-Null
if ($LASTEXITCODE -ne 0) {
  throw "pktmon start failed (exit=$LASTEXITCODE). Ensure PowerShell is elevated and no other pktmon session is running."
}

Write-Host "[2/4] Now place/answer a WhatsApp Desktop call." -ForegroundColor Yellow
if ($AutoStop) {
  Write-Host "Capture will auto-stop in $MaxSeconds seconds (AutoStop)." -ForegroundColor Yellow
} else {
  Write-Host "Press ENTER to stop capture (auto-stops in $MaxSeconds seconds)." -ForegroundColor Yellow
}

$stopJob = Start-Job -ScriptBlock {
  param($Seconds)
  Start-Sleep -Seconds $Seconds
} -ArgumentList $MaxSeconds

if (-not $AutoStop) {
  $null = Read-Host
} else {
  Wait-Job $stopJob | Out-Null
}
try { Stop-Job $stopJob -ErrorAction SilentlyContinue | Out-Null } catch { }
try { Remove-Job $stopJob -Force -ErrorAction SilentlyContinue | Out-Null } catch { }

Write-Host "[3/4] Stopping pktmon and converting to pcapng..." -ForegroundColor Cyan
pktmon stop | Out-Null
if ($LASTEXITCODE -ne 0) {
  throw "pktmon stop failed (exit=$LASTEXITCODE)."
}
pktmon format "$etlPath" -o "$pcapPath" | Out-Null
if ($LASTEXITCODE -ne 0) {
  throw "pktmon format failed (exit=$LASTEXITCODE)."
}

Write-Host "- PCAPNG: $pcapPath" -ForegroundColor Green

# Quick validation: ensure the capture actually contains packets.
try {
  $tmpTxt = Join-Path $fullOutDir ("wa_desktop_{0}_validate.txt" -f $stamp)
  pktmon etl2txt "$etlPath" --out "$tmpTxt" --brief --hex --timestamp --no-ethernet | Out-Null
  $hasPackets = $false
  if (Test-Path $tmpTxt) {
    $first = Get-Content -Path $tmpTxt -TotalCount 200 -ErrorAction SilentlyContinue
    if ($first -and ($first | Select-String -Pattern 'PktGroupId' -SimpleMatch -Quiet)) {
      $hasPackets = $true
    }
  }
  if (-not $hasPackets) {
    Write-Host "[WARN] Capture validation: NO packet records found (no 'PktGroupId')." -ForegroundColor Yellow
    Write-Host "       This usually means pktmon did not capture (not elevated, or another session, or capture ended immediately)." -ForegroundColor Yellow
    Write-Host "       Re-run in an elevated PowerShell and keep WhatsApp Desktop active during capture." -ForegroundColor Yellow
  }
} catch {
  Write-Host "[WARN] Could not validate capture contents: $($_.Exception.Message)" -ForegroundColor Yellow
}

if (-not $SkipProcess) {
  $processor = Join-Path $repoRoot "helpers/process-wa-desktop-turn-capture.ps1"
  if (Test-Path $processor) {
    Write-Host "[3.1/4] Exporting TURN Allocate JSON (pktmon etl2txt + parser)..." -ForegroundColor Cyan
    powershell -ExecutionPolicy Bypass -File "$processor" -EtlPath "$etlPath" -OutDir "$OutDir" -MaxPackets $ProcessMaxPackets | Out-Null
  } else {
    Write-Host "[WARN] Processor script not found: $processor" -ForegroundColor Yellow
  }
}

if ($NoTshark) {
  Write-Host "[4/4] Skipping tshark summary (NoTshark set)." -ForegroundColor DarkGray
  exit 0
}

$tshark = Get-Command tshark -ErrorAction SilentlyContinue
if (-not $tshark) {
  Write-Host "[4/4] tshark not found. If Wireshark is installed, tshark is usually included." -ForegroundColor Yellow
  Write-Host "Open the file in Wireshark and filter with: stun || udp.port == 3478 || udp.port == 443 || tcp.port == 443" -ForegroundColor Yellow
  exit 0
}

Write-Host "[4/4] tshark summary (STUN/TURN packets)" -ForegroundColor Cyan

$displayFilter = "stun || udp.port == 3478 || udp.port == 443 || tcp.port == 443"

Write-Host "- Display filter: $displayFilter" -ForegroundColor DarkGray

# Best-effort: field names may vary by tshark version; this still gives a usable overview.
& $tshark.Path -r "$pcapPath" -Y "$displayFilter" -q -z io,stat,1 | Out-Host

Write-Host "---" -ForegroundColor DarkGray
Write-Host "First STUN/TURN-like frames (best-effort fields):" -ForegroundColor DarkGray

try {
  & $tshark.Path -r "$pcapPath" -Y "$displayFilter" -T fields `
    -e frame.number -e frame.time_relative -e ip.src -e ip.dst -e udp.srcport -e udp.dstport -e tcp.srcport -e tcp.dstport `
    -e stun.type -e stun.method -e stun.class -e stun.username `
    -E header=y -E separator=$"\t" -E quote=n `
    | Select-Object -First 40 | Out-Host
} catch {
  Write-Host "Could not extract STUN fields via tshark on this machine. The pcapng is still valid for Wireshark." -ForegroundColor Yellow
}
