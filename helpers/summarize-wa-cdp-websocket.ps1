param(
  [string]$LogPath = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\wa_cdp_console.log"
)

if (!(Test-Path $LogPath)) {
  Write-Output "log not found: $LogPath"
  exit 1
}

$lines = Get-Content -Path $LogPath
$rows = @()

foreach ($line in $lines) {
  if ($line -match '\[WA-MON\] WebSocket\.(send|message) (\{.*\})$') {
    $dir = $Matches[1]
    $json = $Matches[2]
    try {
      $obj = $json | ConvertFrom-Json
      $rows += [pscustomobject]@{
        Dir = $dir
        Seq = $obj.seq
        Ts = $obj.ts
        Type = $obj.type
        Len = $obj.len
        LenHdr = $obj.len_hdr
        LenHdrMatches = $obj.len_hdr_matches
        Prefix3Hex = $obj.prefix3_hex
        HeadHex = $obj.head_hex
        HeadAscii = $obj.head_ascii
      }
    } catch {
    }
  }
}

$rows = @($rows | Where-Object { $_.Seq -ne $null } | Sort-Object Seq)

if ($rows.Count -eq 0) {
  Write-Output "no websocket frames with seq found"
  exit 0
}

$rows | Select-Object Seq,Dir,Len,LenHdr,LenHdrMatches,Prefix3Hex,HeadAscii,HeadHex | Format-Table -AutoSize
Write-Output ""
Write-Output ("frame_count = {0}" -f $rows.Count)
Write-Output ("len_hdr_matches_true = {0}" -f (@($rows | Where-Object { $_.LenHdrMatches -eq $true }).Count))
Write-Output ("len_hdr_matches_false = {0}" -f (@($rows | Where-Object { $_.LenHdrMatches -eq $false }).Count))
Write-Output ""
Write-Output "by_dir_len"
$rows | Group-Object Dir,Len | Sort-Object Count -Descending | ForEach-Object {
  "{0} x{1}" -f $_.Name, $_.Count
}
Write-Output ""
Write-Output "by_prefix3"
$rows | Group-Object Dir,Prefix3Hex | Sort-Object Count -Descending | ForEach-Object {
  "{0} x{1}" -f $_.Name, $_.Count
}
