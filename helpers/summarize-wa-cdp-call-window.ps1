param(
  [string]$LogPath = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\wa_cdp_page_console.log",
  [int]$StartLine = 1519,
  [int]$EndLine = 1557
)

if (-not (Test-Path -LiteralPath $LogPath)) {
  Write-Error "log not found: $LogPath"
  exit 1
}

$lines = Get-Content -Path $LogPath
$max = $lines.Count - 1
if ($StartLine -lt 0) { $StartLine = 0 }
if ($EndLine -gt $max) { $EndLine = $max }

for ($i = $StartLine; $i -le $EndLine; $i++) {
  $line = $lines[$i]
  if (
    $line -match 'WebSocket\.send' -or
    $line -match 'webSocketFrameSent' -or
    $line -match 'webSocketFrameReceived' -or
    $line -match 'callOutcome' -or
    $line -match 'false_' -or
    $line -match 'Worker\.postMessage' -or
    $line -match 'ps_tokens' -or
    $line -match 'wam_meta'
  ) {
    "{0}: {1}" -f $i, $line
  }
}
