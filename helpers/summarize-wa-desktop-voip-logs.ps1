param(
  [string]$ArtifactsDir = "",
  [string]$OutDir = "",
  [int]$MaxLines = 200000,
  [int]$Tail = 2000,
  [int]$MaxEvents = 3000,
  [int]$MaxMessageChars = 600,
  [switch]$Quiet
)

$ErrorActionPreference = "Stop"

function Find-LatestArtifactsDir {
  $base = Join-Path (Get-Location) ".dist\wa_desktop_artifacts"
  if (-not (Test-Path $base)) { return "" }
  $d = Get-ChildItem -LiteralPath $base -Directory -Filter "wa_desktop_*" | Sort-Object LastWriteTime -Descending | Select-Object -First 1
  if (-not $d) { return "" }
  return $d.FullName
}

if ([string]::IsNullOrWhiteSpace($ArtifactsDir)) {
  $ArtifactsDir = Find-LatestArtifactsDir
}
if ([string]::IsNullOrWhiteSpace($ArtifactsDir) -or -not (Test-Path $ArtifactsDir)) {
  throw "ArtifactsDir not found. Run helpers/export-wa-desktop-artifacts.ps1 first or provide -ArtifactsDir."
}

if ([string]::IsNullOrWhiteSpace($OutDir)) {
  $OutDir = $ArtifactsDir
}
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

$rot = Join-Path $ArtifactsDir "rotatedLogs"
if (-not (Test-Path $rot)) {
  throw "rotatedLogs not found under: $ArtifactsDir"
}

$logFiles = @()
$logFiles += @(Get-ChildItem -LiteralPath $rot -File -Filter "applog.txt.*" -ErrorAction SilentlyContinue)

# Also include the current (non-rotated) applog.txt if present. Recent VoIP activity may be there.
$currentAppLog = Join-Path $ArtifactsDir "applog.txt"
if (Test-Path $currentAppLog) {
  $logFiles += @(Get-Item -LiteralPath $currentAppLog -ErrorAction SilentlyContinue)
}

$logFiles = @($logFiles | Sort-Object LastWriteTime)
if (-not $logFiles -or $logFiles.Count -eq 0) {
  throw "No rotated applog files found in: $rot"
}

# Parse lines like:
# [?? W P:9904 T:012464 D:-1 16-02-26 09:51:33.956 WARN] voip > wa_transport.cc A555DC  bind record not found...
$rx = [regex]'^\[(?<q>[^\]]+)\s+(?<dt>\d{2}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}\.\d{3})\s+(?<lvl>[A-Z]+)\]\s+(?<chan>[^>]+)\s*>\s*(?<msg>.*)$'
$rxCall = [regex]'\bA[0-9A-F]{5,12}\b'

$events = New-Object System.Collections.Generic.List[object]
$linesRead = 0

foreach ($f in $logFiles) {
  $i = 0
  foreach ($line in (Get-Content -LiteralPath $f.FullName -ErrorAction SilentlyContinue)) {
    $linesRead++
    if ($MaxLines -gt 0 -and $linesRead -gt $MaxLines) { break }

    $m = $rx.Match($line)
    if (-not $m.Success) { continue }

    $chan = $m.Groups['chan'].Value.Trim()
    if ($chan -notmatch '^(?i)voip$') { continue }

    $msg = $m.Groups['msg'].Value
    $callTag = ""
    $mc = $rxCall.Match($msg)
    if ($mc.Success) { $callTag = $mc.Value }

    $dtRaw = $m.Groups['dt'].Value
    $dt = $null
    try {
      # dtRaw is yy-mm-dd HH:mm:ss.mmm
      $dt = [DateTime]::ParseExact($dtRaw, 'yy-MM-dd HH:mm:ss.fff', [System.Globalization.CultureInfo]::InvariantCulture)
    } catch {
      $dt = $dtRaw
    }

    $events.Add([PSCustomObject]@{
      file = $f.Name
      ts = $dt
      level = $m.Groups['lvl'].Value
      call_tag = $callTag
      message = $msg.Trim()
      line = $line
    }) | Out-Null

    $i++
  }

  if ($MaxLines -gt 0 -and $linesRead -gt $MaxLines) { break }
}

# Save lightweight summary + NDJSON events to avoid ConvertTo-Json OOM on Windows PowerShell.
$summaryPath = Join-Path $OutDir "wa_desktop_voip_log_summary.json"
$eventsPath = Join-Path $OutDir "wa_desktop_voip_events.ndjson"

$total = $events.Count
$kept = $events
if ($MaxEvents -gt 0 -and $total -gt $MaxEvents) {
  $kept = $events | Select-Object -Last $MaxEvents
}

$byLevel = @{}
$byCall = @{}
foreach ($e in $kept) {
  $lvl = [string]$e.level
  if (-not $byLevel.ContainsKey($lvl)) { $byLevel[$lvl] = 0 }
  $byLevel[$lvl] = [int]$byLevel[$lvl] + 1

  $ct = [string]$e.call_tag
  if (-not [string]::IsNullOrWhiteSpace($ct)) {
    if (-not $byCall.ContainsKey($ct)) { $byCall[$ct] = 0 }
    $byCall[$ct] = [int]$byCall[$ct] + 1
  }
}

$topCalls = $byCall.GetEnumerator() | Sort-Object Value -Descending | Select-Object -First 30 |
  ForEach-Object { [PSCustomObject]@{ call_tag = $_.Key; count = $_.Value } }

$summary = [PSCustomObject]@{
  artifacts_dir = $ArtifactsDir
  rotated_logs = $rot
  files = @($logFiles | Select-Object Name,LastWriteTime,Length)
  parsed_lines = $linesRead
  events_total = $total
  events_kept = $kept.Count
  by_level = $byLevel
  top_calls = $topCalls
  events_file = [System.IO.Path]::GetFileName($eventsPath)
}
$summary | ConvertTo-Json -Depth 6 | Out-File -FilePath $summaryPath -Encoding utf8

# Write NDJSON events (minimal fields)
$eventsTempPath = $eventsPath + ".tmp"
Remove-Item -LiteralPath $eventsTempPath -Force -ErrorAction SilentlyContinue | Out-Null

$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
$sw = New-Object System.IO.StreamWriter($eventsTempPath, $false, $utf8NoBom)
try {
  foreach ($e in $kept) {
    $ts = $e.ts
    $tsStr = if ($ts -is [DateTime]) { $ts.ToString('o') } else { [string]$ts }
    $msg = [string]$e.message
    if ($MaxMessageChars -gt 0 -and $msg.Length -gt $MaxMessageChars) {
      $msg = $msg.Substring(0, $MaxMessageChars)
    }
    $row = [PSCustomObject]@{
      ts = $tsStr
      level = [string]$e.level
      call_tag = [string]$e.call_tag
      file = [string]$e.file
      message = $msg
    }
    $sw.WriteLine(($row | ConvertTo-Json -Compress))
  }
} finally {
  $sw.Close()
}

try {
  Move-Item -LiteralPath $eventsTempPath -Destination $eventsPath -Force
} catch {
  Write-Warning "Failed to replace NDJSON file (in use?): $eventsPath. Keeping temp file: $eventsTempPath"
}

# Print tail
if (-not $Quiet) {
  Write-Host "Saved: $summaryPath" -ForegroundColor Green
  Write-Host "Saved: $eventsPath" -ForegroundColor Green
  Write-Host "Events parsed: $($events.Count) (kept=$($kept.Count))" -ForegroundColor DarkGray
  Write-Host "--- Tail (voip) ---" -ForegroundColor Cyan
  $events | Select-Object -Last $Tail | ForEach-Object { $_.line }
}
