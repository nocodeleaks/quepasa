param(
  [string]$EtlPath = "",
  [string]$OutDir = ".dist/pcaps",
  [int]$MaxPackets = 500,
  [string]$OnlyMsgTypeHex = "0x0003"
)

$ErrorActionPreference = "Stop"

function Require-Command([string]$Name) {
  $cmd = Get-Command $Name -ErrorAction SilentlyContinue
  if (-not $cmd) {
    throw "Missing required command '$Name'."
  }
  return $cmd
}

Require-Command pktmon | Out-Null

$repoRoot = Split-Path -Parent $PSScriptRoot
$fullOutDir = Join-Path $repoRoot $OutDir
New-Item -ItemType Directory -Force -Path $fullOutDir | Out-Null

if (-not $EtlPath -or $EtlPath.Trim() -eq "") {
  $latest = Get-ChildItem -LiteralPath $fullOutDir -Filter "wa_desktop_*.etl" -ErrorAction SilentlyContinue |
    Sort-Object LastWriteTime -Descending |
    Select-Object -First 1
  if (-not $latest) {
    throw "No ETL capture found in $fullOutDir. Run helpers/capture-wa-desktop-turn.ps1 first."
  }
  $EtlPath = $latest.FullName
}

if (-not (Test-Path $EtlPath)) {
  throw "ETL not found: $EtlPath"
}

$baseName = [System.IO.Path]::GetFileNameWithoutExtension($EtlPath)
$txtPath = Join-Path $fullOutDir ($baseName + ".txt")

$onlyLabel = $OnlyMsgTypeHex
if (-not $onlyLabel -or $onlyLabel.Trim() -eq "") {
  $onlyLabel = "0x0003"
}

$passOnly = $OnlyMsgTypeHex
if ($onlyLabel.Trim().ToLower() -eq "all") {
  $onlyLabel = "all"
  $passOnly = "all"
}

$safeLabel = ($onlyLabel -replace '[^0-9A-Za-zx]+', '_')
$jsonPath = Join-Path $fullOutDir ($baseName + "_stun_" + $safeLabel + ".json")

Write-Host "[1/2] pktmon etl2txt -> $txtPath" -ForegroundColor Cyan
pktmon etl2txt "$EtlPath" --out "$txtPath" --brief --hex --timestamp --no-ethernet | Out-Null

$parser = Join-Path $repoRoot "helpers/parse-pktmon-stun.ps1"
if (-not (Test-Path $parser)) {
  throw "Parser not found: $parser"
}

Write-Host "[2/2] Export STUN/TURN msgType=$onlyLabel -> $jsonPath" -ForegroundColor Cyan
powershell -ExecutionPolicy Bypass -File "$parser" -InputTxt "$txtPath" -MaxPackets $MaxPackets -OnlyMsgTypeHex $passOnly -OutJson "$jsonPath" -Quiet | Out-Null

Write-Host "Done." -ForegroundColor Green
Write-Host "- ETL : $EtlPath" -ForegroundColor DarkGray
Write-Host "- TXT : $txtPath" -ForegroundColor DarkGray
Write-Host "- JSON: $jsonPath" -ForegroundColor DarkGray
