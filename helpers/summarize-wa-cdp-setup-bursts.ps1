param(
  [string]$Log = 'Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\wa_cdp_page_console.log',
  [int]$Radius = 6
)
$lines = Get-Content -Path $Log
$re = 'WebSocket\.send .*"len":(431|446|459|514|634|724|2683)'
for ($i = 0; $i -lt $lines.Count; $i++) {
  if ($lines[$i] -match $re) {
    $from = [Math]::Max(0, $i - $Radius)
    $to = [Math]::Min($lines.Count - 1, $i + $Radius)
    "--- setup window $($from+1)..$($to+1) ---"
    for ($j=$from; $j -le $to; $j++) { '{0}: {1}' -f ($j+1), $lines[$j] }
    ''
  }
}
