param(
  [string]$TurnAllocateJson = "",
  [string]$VoipEventsNdjson = "",
  [string]$OutDir = ".dist/wa_desktop_correlation",
  [int]$WindowSeconds = 12,
  [int]$MaxAllocates = 200,
  [int]$MaxVoipLinesPerAllocate = 30,
  [switch]$Quiet
)

$ErrorActionPreference = "Stop"

function Find-LatestFile([string]$globPath) {
  $items = Get-ChildItem -Path $globPath -File -ErrorAction SilentlyContinue | Sort-Object LastWriteTime -Descending
  if ($items -and $items.Count -gt 0) { return $items[0].FullName }
  return ""
}

if ([string]::IsNullOrWhiteSpace($TurnAllocateJson)) {
  $TurnAllocateJson = Find-LatestFile ".dist/pcaps/wa_desktop_*_turn_allocate*.json"
}
if ([string]::IsNullOrWhiteSpace($VoipEventsNdjson)) {
  $VoipEventsNdjson = Find-LatestFile ".dist/wa_desktop_artifacts/wa_desktop_*/wa_desktop_voip_events.ndjson"
}

if ([string]::IsNullOrWhiteSpace($TurnAllocateJson) -or -not (Test-Path $TurnAllocateJson)) {
  throw "TurnAllocateJson not found. Provide -TurnAllocateJson or generate with helpers/process-wa-desktop-turn-capture.ps1."
}
if ([string]::IsNullOrWhiteSpace($VoipEventsNdjson) -or -not (Test-Path $VoipEventsNdjson)) {
  throw "VoipEventsNdjson not found. Provide -VoipEventsNdjson or generate with helpers/summarize-wa-desktop-voip-logs.ps1."
}

New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

$base = [System.IO.Path]::GetFileName($TurnAllocateJson)

$turn = Get-Content -Raw -Path $TurnAllocateJson | ConvertFrom-Json
$allocs = @($turn.items)
if ($allocs.Count -gt $MaxAllocates) {
  $allocs = $allocs | Select-Object -First $MaxAllocates
}

# Load NDJSON voip events
$voip = New-Object System.Collections.Generic.List[object]
foreach ($line in (Get-Content -Path $VoipEventsNdjson -ErrorAction SilentlyContinue)) {
  $t = $line.Trim()
  if ($t -eq "") { continue }
  try {
    $o = $t | ConvertFrom-Json
    $dt = $null
    try { $dt = [DateTime]::Parse($o.ts) } catch { }
    if ($dt -eq $null) { continue }
    $voip.Add([PSCustomObject]@{ ticks = $dt.TimeOfDay.Ticks; ts = $dt; level=$o.level; call_tag=$o.call_tag; file=$o.file; message=$o.message }) | Out-Null
  } catch { }
}

$voipArr = @($voip | Sort-Object ticks)

function Parse-PktTimeOfDayTicks([string]$pktTs) {
  if ([string]::IsNullOrWhiteSpace($pktTs)) { return $null }
  try {
    # pktTs is HH:mm:ss.fffffff (or .fff)
    $parts = $pktTs.Split('.')
    $hms = $parts[0]
    $frac = if ($parts.Count -gt 1) { $parts[1] } else { "" }
    $dt = [DateTime]::ParseExact($hms, 'HH:mm:ss', [System.Globalization.CultureInfo]::InvariantCulture)
    $ts = New-TimeSpan -Hours $dt.Hour -Minutes $dt.Minute -Seconds $dt.Second
    if ($frac -ne "") {
      $frac = ($frac + '0000000').Substring(0,7)
      $ticks = [int64]$frac
      $ts = $ts.Add([TimeSpan]::FromTicks($ticks))
    }
    return $ts.Ticks
  } catch {
    return $null
  }
}

function DeltaSecondsTimeOfDay([int64]$aTicks, [int64]$bTicks) {
  $delta = [Math]::Abs($aTicks - $bTicks)
  $day = [TimeSpan]::FromDays(1).Ticks
  if ($delta -gt ($day / 2)) { $delta = $day - $delta }
  return ([TimeSpan]::FromTicks($delta)).TotalSeconds
}

$outStamp = Get-Date -Format "yyyyMMdd_HHmmss"
$outPath = Join-Path $OutDir ("wa_desktop_turn_voip_correlation_" + $outStamp + ".md")

$sb = New-Object System.Text.StringBuilder
[void]$sb.AppendLine("# WhatsApp Desktop TURN <-> VoIP correlation")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("- TurnAllocateJson: $TurnAllocateJson")
[void]$sb.AppendLine("- VoipEventsNdjson: $VoipEventsNdjson")
[void]$sb.AppendLine("- WindowSeconds: $WindowSeconds")
[void]$sb.AppendLine("")

foreach ($a in $allocs) {
  $pktTicks = Parse-PktTimeOfDayTicks $a.pkt_ts
  $gid = $a.pkt_group_id
  $pnum = $a.pkt_number

  $ep = ""
  foreach ($attr in $a.attrs) {
    if ($attr.type -eq '0x0016' -and $attr.decoded -and $attr.decoded.endpoint) { $ep = $attr.decoded.endpoint }
  }

  [void]$sb.AppendLine("## Allocate txid=$($a.txid) endpoint=$ep")
  [void]$sb.AppendLine("- pkt_ts=$($a.pkt_ts) pkt_group_id=$gid pkt_number=$pnum")
  if ($pktTicks -ne $null) {
    $tod = [TimeSpan]::FromTicks([int64]$pktTicks)
    [void]$sb.AppendLine("- approx_time_of_day=$($tod.ToString())")
  }
  [void]$sb.AppendLine("- attrs: 0x4000=$((($a.attrs | Where-Object {$_.type -eq '0x4000'})[0]).len) 0x4024=$((($a.attrs | Where-Object {$_.type -eq '0x4024'})[0]).len) 0x0016=8 mi=$($a.mi)")

  if ($pktTicks -ne $null -and $voipArr.Count -gt 0) {
    $win = [Math]::Max(1,$WindowSeconds)
    $near = New-Object System.Collections.Generic.List[object]
    foreach ($e in $voipArr) {
      $dsec = DeltaSecondsTimeOfDay ([int64]$pktTicks) ([int64]$e.ticks)
      if ($dsec -le $win) {
        $near.Add([PSCustomObject]@{ e=$e; dsec=$dsec }) | Out-Null
        if ($near.Count -ge $MaxVoipLinesPerAllocate) { break }
      }
    }
    [void]$sb.AppendLine("- voip_events_nearby=$($near.Count)")
    if ($near.Count -gt 0) {
      [void]$sb.AppendLine("")
      foreach ($x in $near) {
        $e = $x.e
        $ct = if ([string]::IsNullOrWhiteSpace($e.call_tag)) { "-" } else { $e.call_tag }
        [void]$sb.AppendLine(("  - {0} (+{1:0.000}s) [{2}] call={3} {4}" -f $e.ts.ToString('o'), $x.dsec, $e.level, $ct, $e.message))
      }
    }
  } else {
    [void]$sb.AppendLine("- voip_events_nearby=skipped (missing pkt_ts or no voip data)")
  }

  [void]$sb.AppendLine("")
}

[System.IO.File]::WriteAllText($outPath, $sb.ToString(), [Text.Encoding]::UTF8)

if (-not $Quiet) {
  Write-Host "Saved: $outPath" -ForegroundColor Green
}
