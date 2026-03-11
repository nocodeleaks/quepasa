param(
  [string]$Log = 'Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\wa_cdp_page_console.log',
  [int]$Radius = 6
)
$lines = Get-Content -Path $Log
$patterns = @(
  'WebSocket.send .*"len":37',
  'WebSocket.send .*"len":59',
  'WebSocket.send .*"len":60',
  'idb://model-storage/message/callOutcome',
  'idb://wawc/wam_meta/',
  'idb://wawc/ps_tokens/'
)
$matchIdx = New-Object System.Collections.Generic.List[int]
for ($i = 0; $i -lt $lines.Count; $i++) {
  foreach ($p in $patterns) {
    if ($lines[$i] -match $p) { $matchIdx.Add($i); break }
  }
}
$groups = @()
$start = $null
$prev = $null
foreach ($idx in $matchIdx | Sort-Object -Unique) {
  if ($null -eq $start) { $start = $idx; $prev = $idx; continue }
  if ($idx -le ($prev + $Radius)) { $prev = $idx; continue }
  $groups += ,@($start,$prev)
  $start = $idx; $prev = $idx
}
if ($null -ne $start) { $groups += ,@($start,$prev) }
foreach ($g in $groups) {
  $from = [Math]::Max(0, $g[0]-$Radius)
  $to = [Math]::Min($lines.Count-1, $g[1]+$Radius)
  "--- window $($from+1)..$($to+1) ---"
  for ($j=$from; $j -le $to; $j++) { '{0}: {1}' -f ($j+1), $lines[$j] }
  ''
}
