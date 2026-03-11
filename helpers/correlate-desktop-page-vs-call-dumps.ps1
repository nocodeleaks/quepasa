param(
  [string]$PageLog = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\wa_cdp_page_console.log",
  [string]$DumpDir = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\call_dumps",
  [int]$WindowSeconds = 20
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Parse-DumpFileName {
  param([string]$Name)
  if ($Name -notmatch '^(?<kind>.+?)_(?<ts>\d{14})_(?<callid>[A-F0-9]+)\.json$') {
    return $null
  }
  $dt = [datetime]::ParseExact($matches.ts, 'yyyyMMddHHmmss', [System.Globalization.CultureInfo]::InvariantCulture)
  [pscustomobject]@{
    Kind   = $matches.kind
    CallID = $matches.callid
    Time   = $dt
    File   = $Name
  }
}

function Parse-PageEvent {
  param(
    [string]$Line,
    [int]$LineNo
  )

  if ($Line -notmatch '\[WA-MON\] WebSocket\.send .*"len":(?<len>\d+).*"prefix3_hex":"(?<prefix>[0-9a-f]+)".*"ts":(?<ts>\d+),"seq":(?<seq>\d+)') {
    return $null
  }

  $ms = [int64]$matches.ts
  $utc = [datetimeoffset]::FromUnixTimeMilliseconds($ms).UtcDateTime
  [pscustomobject]@{
    LineNo   = $LineNo
    Seq      = [int]$matches.seq
    Len      = [int]$matches.len
    Prefix   = $matches.prefix
    TimeUtc  = $utc
    RawLine  = $Line
  }
}

if (-not (Test-Path -LiteralPath $PageLog)) {
  throw "Page log not found: $PageLog"
}

if (-not (Test-Path -LiteralPath $DumpDir)) {
  throw "Dump dir not found: $DumpDir"
}

$dumpEvents = Get-ChildItem -LiteralPath $DumpDir -File |
  ForEach-Object { Parse-DumpFileName -Name $_.Name } |
  Where-Object { $_ -ne $null } |
  Sort-Object Time

if (-not $dumpEvents) {
  Write-Host "no call dump events found"
  exit 0
}

$pageEvents = New-Object System.Collections.Generic.List[object]
$lineNo = 0
Get-Content -LiteralPath $PageLog | ForEach-Object {
  $lineNo++
  $evt = Parse-PageEvent -Line $_ -LineNo $lineNo
  if ($null -ne $evt) {
    $pageEvents.Add($evt)
  }
}

if ($pageEvents.Count -eq 0) {
  Write-Host "no page websocket send events found"
  exit 0
}

$window = [timespan]::FromSeconds($WindowSeconds)

$callIds = $dumpEvents | Select-Object -ExpandProperty CallID -Unique
foreach ($callId in $callIds) {
  $callEvents = $dumpEvents | Where-Object { $_.CallID -eq $callId } | Sort-Object Time
  $first = $callEvents[0].Time
  $from = $first.AddSeconds(-$WindowSeconds)
  $to = $first.AddSeconds($WindowSeconds)
  $pageWindow = $pageEvents | Where-Object { $_.TimeUtc -ge $from -and $_.TimeUtc -le $to } | Sort-Object TimeUtc, LineNo

  ""
  "=== CallID $callId ==="
  "Dump events:"
  foreach ($e in $callEvents) {
    "  {0:u}  {1}  {2}" -f $e.Time, $e.Kind, $e.File
  }

  if (-not $pageWindow) {
    "Page events in +/-$WindowSeconds" + "s: none"
    continue
  }

  "Page sends in +/-$WindowSeconds" + "s:"
  foreach ($p in $pageWindow) {
    "  {0:u}  line={1} seq={2} len={3} prefix={4}" -f $p.TimeUtc, $p.LineNo, $p.Seq, $p.Len, $p.Prefix
  }

  "Page family summary:"
  $pageWindow |
    Group-Object Len, Prefix |
    Sort-Object Count -Descending |
    ForEach-Object {
      "  {0,3}x len={1} prefix={2}" -f $_.Count, $_.Group[0].Len, $_.Group[0].Prefix
    }
}
