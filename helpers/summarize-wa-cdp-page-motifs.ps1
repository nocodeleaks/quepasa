param(
  [string]$LogPath = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\wa_cdp_page_console.log"
)

if (-not (Test-Path -LiteralPath $LogPath)) {
  Write-Error "log not found: $LogPath"
  exit 1
}

$lines = Get-Content -Path $LogPath
$events = @()

for ($i = 0; $i -lt $lines.Count; $i++) {
  $line = $lines[$i]

  if ($line -match 'WebSocket\.send .*?"len":(\d+).*?"prefix3_hex":"([0-9a-f]*)".*?"full_hex":"([0-9a-f]+)"') {
    $events += [pscustomobject]@{
      Line   = $i
      Dir    = 'send'
      Len    = [int]$matches[1]
      Prefix = $matches[2]
      Hex    = $matches[3]
    }
    continue
  }

  if ($line -match 'webSocketFrameReceived .*?"bin_len": (\d+).*?"bin_head_hex": "([0-9a-f]+)"') {
    $hex = $matches[2]
    $events += [pscustomobject]@{
      Line   = $i
      Dir    = 'recv'
      Len    = [int]$matches[1]
      Prefix = $hex.Substring(0, [Math]::Min(6, $hex.Length))
      Hex    = $hex
    }
  }
}

$events = $events | Sort-Object Line

Write-Output "motifs around send(634,000277):"
$starts = $events | Where-Object { $_.Dir -eq 'send' -and $_.Len -eq 634 -and $_.Prefix -eq '000277' }

if ($starts.Count -eq 0) {
  Write-Output "none"
  exit 0
}

foreach ($start in $starts) {
  $window = $events | Where-Object { $_.Line -ge $start.Line -and $_.Line -le ($start.Line + 12) }
  foreach ($ev in $window) {
    "{0}: {1}({2},{3})" -f $ev.Line, $ev.Dir, $ev.Len, $ev.Prefix
  }
  Write-Output ""
}
