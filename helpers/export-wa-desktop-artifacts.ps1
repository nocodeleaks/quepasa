param(
  [string]$OutDir = ".dist/wa_desktop_artifacts",
  [int]$MaxRotatedLogs = 20,
  [switch]$IncludeSessions,
  [switch]$Quiet
)

$ErrorActionPreference = "Stop"

$pkgRoot = Join-Path $env:LOCALAPPDATA "Packages\5319275A.WhatsAppDesktop_cv1g1gvanyjgm"
if (-not (Test-Path $pkgRoot)) {
  throw "WhatsApp Desktop Store package not found at: $pkgRoot"
}

$localState = Join-Path $pkgRoot "LocalState"
if (-not (Test-Path $localState)) {
  throw "LocalState not found: $localState"
}

$repoRoot = Split-Path -Parent $PSScriptRoot
$fullOutDir = Join-Path $repoRoot $OutDir
New-Item -ItemType Directory -Force -Path $fullOutDir | Out-Null

$stamp = Get-Date -Format "yyyyMMdd_HHmmss"
$out = Join-Path $fullOutDir ("wa_desktop_" + $stamp)
New-Item -ItemType Directory -Force -Path $out | Out-Null

function Copy-IfExists([string]$src, [string]$dstDir) {
  if (Test-Path $src) {
    Copy-Item -LiteralPath $src -Destination $dstDir -Force
    return $true
  }
  return $false
}

# Core log/state files
$targets = @(
  "applog.txt",
  "reglog.txt",
  "wa_stats_v2.log",
  "wa_voip_history.json",
  "detailed_transport_record_record.json",
  "kCallingHistoryCallStateV1RecordType_record.json",
  "session.db",
  "session.db-wal",
  "session.db-shm"
)

$copied = New-Object System.Collections.Generic.List[string]
foreach ($t in $targets) {
  $src = Join-Path $localState $t
  if (Copy-IfExists $src $out) {
    $copied.Add($t) | Out-Null
  }
}

# Rotated logs
$rot = Join-Path $localState "rotatedLogs"
if (Test-Path $rot) {
  $dst = Join-Path $out "rotatedLogs"
  New-Item -ItemType Directory -Force -Path $dst | Out-Null
  Get-ChildItem -LiteralPath $rot -Force -File -Filter "applog.txt.*" |
    Sort-Object LastWriteTime -Descending |
    Select-Object -First ([Math]::Max(0, $MaxRotatedLogs)) |
    ForEach-Object { Copy-Item -LiteralPath $_.FullName -Destination $dst -Force }
}

# Sessions folder can be large; keep it opt-in.
if ($IncludeSessions) {
  $sessions = Join-Path $localState "sessions"
  if (Test-Path $sessions) {
    $dst = Join-Path $out "sessions"
    Copy-Item -LiteralPath $sessions -Destination $dst -Recurse -Force
  }
}

if (-not $Quiet) {
  Write-Host "Exported WhatsApp Desktop artifacts." -ForegroundColor Green
  Write-Host "- Source: $localState" -ForegroundColor DarkGray
  Write-Host "- Output: $out" -ForegroundColor DarkGray
  Write-Host "- Copied: $($copied -join ', ')" -ForegroundColor DarkGray
}
