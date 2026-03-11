param(
  [string]$LogPath = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\wa_cdp_page_console.log"
)

if (-not (Test-Path $LogPath)) {
  throw "log not found: $LogPath"
}

$rows = @()
Get-Content -Path $LogPath | ForEach-Object {
  $line = $_
  if ($line -match '\[WA-MON\] crypto\.(direct|proto|result)\.(encrypt|decrypt)(?:\.error)?\s+(\{.*\})$') {
    $scope = $Matches[1]
    $op = $Matches[2]
    $json = $Matches[3]
    try {
      $obj = $json | ConvertFrom-Json
      $rows += [pscustomobject]@{
        Scope = $scope
        Op = $op
        Len = if ($obj.data -and $obj.data.len) { $obj.data.len } elseif ($obj.len) { $obj.len } else { $null }
        HeadHex = if ($obj.data -and $obj.data.head_hex) { $obj.data.head_hex } elseif ($obj.head_hex) { $obj.head_hex } else { $null }
        IV = if ($obj.algorithm -and $obj.algorithm.iv -and $obj.algorithm.iv.head_hex) { $obj.algorithm.iv.head_hex } else { $null }
      }
    } catch {}
  }
}

$rows | Format-Table -AutoSize
