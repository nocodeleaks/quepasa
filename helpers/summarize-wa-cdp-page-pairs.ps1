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

Write-Output ("events={0}" -f $events.Count)
Write-Output ""
Write-Output "families:"
$events |
  Group-Object Dir, Len, Prefix |
  Sort-Object Count -Descending |
  ForEach-Object {
    "{0,3}  {1}" -f $_.Count, $_.Name
  }

Write-Output ""
Write-Output "adjacent send->recv pairs (gap<=4 lines):"

$pairs = @()
for ($j = 0; $j -lt ($events.Count - 1); $j++) {
  $a = $events[$j]
  $b = $events[$j + 1]
  if ($a.Dir -eq 'send' -and $b.Dir -eq 'recv' -and (($b.Line - $a.Line) -le 4)) {
    $pairs += [pscustomobject]@{
      SendLine   = $a.Line
      RecvLine   = $b.Line
      Gap        = $b.Line - $a.Line
      SendLen    = $a.Len
      SendPrefix = $a.Prefix
      RecvLen    = $b.Len
      RecvPrefix = $b.Prefix
      SendHex    = $a.Hex
      RecvHex    = $b.Hex
    }
  }
}

if ($pairs.Count -eq 0) {
  Write-Output "no adjacent pairs found"
  exit 0
}

$pairs | ForEach-Object {
  "send({0},{1}) -> recv({2},{3}) gap={4} [lines {5}->{6}]" -f $_.SendLen, $_.SendPrefix, $_.RecvLen, $_.RecvPrefix, $_.Gap, $_.SendLine, $_.RecvLine
}

Write-Output ""
Write-Output "pair families:"
$pairs |
  Group-Object SendLen, SendPrefix, RecvLen, RecvPrefix |
  Sort-Object Count -Descending |
  ForEach-Object {
    "{0,3}  {1}" -f $_.Count, $_.Name
  }

Write-Output ""
Write-Output "sample pairs:"
$pairs |
  Select-Object -First 3 |
  ForEach-Object {
    [pscustomobject]@{
      SendLen = $_.SendLen
      SendPrefix = $_.SendPrefix
      SendHex = $_.SendHex
      RecvLen = $_.RecvLen
      RecvPrefix = $_.RecvPrefix
      RecvHex = $_.RecvHex
    }
  } | Format-List
