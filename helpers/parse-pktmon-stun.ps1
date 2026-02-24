param(
  [Parameter(Mandatory = $true)]
  [string]$InputTxt,
  [int]$MaxPackets = 50,
  [switch]$ShowUnknownAttrs,
  [switch]$ShowRawUsername,
  [switch]$DebugParse,
  [string]$OutJson = "",
  [string]$OnlyMsgTypeHex = "",
  [switch]$Quiet
)

$ErrorActionPreference = "Stop"

# Sentinel: some runners cannot pass an empty string argument reliably.
# Treat OnlyMsgTypeHex="all" as "no filter".
if ($OnlyMsgTypeHex -and $OnlyMsgTypeHex.Trim().ToLower() -eq 'all') {
  $OnlyMsgTypeHex = ""
}

function HexToBytes([string[]]$words) {
  $bytes = New-Object System.Collections.Generic.List[byte]
  foreach ($w in $words) {
    $t = $w.Trim()
    if ($t -match '^[0-9A-Fa-f]{4}$') {
      $bytes.Add([Convert]::ToByte($t.Substring(0, 2), 16))
      $bytes.Add([Convert]::ToByte($t.Substring(2, 2), 16))
    } elseif ($t -match '^[0-9A-Fa-f]{2}$') {
      $bytes.Add([Convert]::ToByte($t, 16))
    }
  }
  return ,$bytes.ToArray()
}

function BytesToHex([byte[]]$b, [int]$max = 64) {
  if (-not $b) { return "" }
  $take = [Math]::Min($b.Length, $max)
  ($b[0..($take-1)] | ForEach-Object { $_.ToString('x2') }) -join ''
}

function BytesToAsciiSafe([byte[]]$b) {
  if (-not $b) { return "" }
  $printable = $true
  foreach ($x in $b) {
    if ($x -lt 0x20 -or $x -gt 0x7E) { $printable = $false; break }
  }
  if (-not $printable) { return "" }
  return [Text.Encoding]::ASCII.GetString($b)
}

function ReadU16BE([byte[]]$b, [int]$o) {
  return (([int]$b[$o]) -shl 8) -bor ([int]$b[$o+1])
}

function ReadU32BE([byte[]]$b, [int]$o) {
  return ((([int]$b[$o]) -shl 24) -bor (([int]$b[$o+1]) -shl 16) -bor (([int]$b[$o+2]) -shl 8) -bor ([int]$b[$o+3]))
}

function SliceBytes([byte[]]$b, [int]$start, [int]$length) {
  if (-not $b) { return $null }
  if ($length -le 0) { return @() }
  if ($start -lt 0) { return $null }
  if ($start -ge $b.Length) { return $null }
  $end = $start + $length - 1
  if ($end -ge $b.Length) { $end = $b.Length - 1 }
  if ($end -lt $start) { return @() }
  return $b[$start..$end]
}

function ParseIPv4UdpPayload([byte[]]$pkt) {
  # Ethernet header is present in pktmon dumps: 14 bytes.
  if ($pkt.Length -lt 14 + 20 + 8) { return $null }
  $ethType = ReadU16BE $pkt 12
  if ($ethType -ne 0x0800) { return $null }
  $ipOff = 14
  $verIhl = $pkt[$ipOff]
  $ver = ($verIhl -shr 4)
  if ($ver -ne 4) { return $null }
  $ihl = ($verIhl -band 0x0F) * 4
  if ($ihl -lt 20) { return $null }
  $proto = $pkt[$ipOff + 9]
  if ($proto -ne 17) { return $null }
  $udpOff = $ipOff + $ihl
  if ($pkt.Length -lt $udpOff + 8) { return $null }
  $srcPort = ReadU16BE $pkt $udpOff
  $dstPort = ReadU16BE $pkt ($udpOff + 2)
  $udpLen = ReadU16BE $pkt ($udpOff + 4)
  $payloadOff = $udpOff + 8
  $payloadLen = $udpLen - 8
  if ($payloadLen -le 0) {
    return [PSCustomObject]@{ SrcPort = $srcPort; DstPort = $dstPort; Payload = @() }
  }
  if ($pkt.Length -lt $payloadOff + $payloadLen) {
    $payloadLen = [Math]::Max(0, $pkt.Length - $payloadOff)
  }
  $payload = SliceBytes $pkt $payloadOff $payloadLen
  return [PSCustomObject]@{ SrcPort = $srcPort; DstPort = $dstPort; Payload = $payload }
}

function ParseStunMessage([byte[]]$payload) {
  if ($payload.Length -lt 20) { return $null }
  $msgType = ReadU16BE $payload 0
  $msgLen = ReadU16BE $payload 2
  $cookie = ReadU32BE $payload 4
  if ($cookie -ne 0x2112A442) { return $null }
  if ($payload.Length -lt 20 + $msgLen) { return $null }
  $txidBytes = SliceBytes $payload 8 12
  if (-not $txidBytes -or $txidBytes.Length -ne 12) { return $null }
  $txidHex = BytesToHex $txidBytes 24
  $attrs = @()
  $pos = 20
  $end = 20 + $msgLen
  while ($pos + 4 -le $end) {
    $at = ReadU16BE $payload $pos
    $al = ReadU16BE $payload ($pos + 2)
    $pos += 4
    if ($pos + $al -gt $end) { break }
    $val = if ($al -gt 0) { SliceBytes $payload $pos $al } else { @() }
    $pos += $al
    # 32-bit padding
    $pad = (4 - ($al % 4)) % 4
    $pos += $pad
    $attrs += [PSCustomObject]@{ Type = $at; Len = $al; Value = $val }
  }
  return [PSCustomObject]@{ Type = $msgType; Len = $msgLen; TxID = $txidHex; Attrs = $attrs }
}

function FindStunMessagesInPacket([byte[]]$pkt) {
  if (-not $pkt -or $pkt.Length -lt 20) { return @() }
  $cookie = @(0x21, 0x12, 0xA4, 0x42)
  $out = New-Object System.Collections.Generic.List[object]
  # Scan for STUN header starts: [type(2)][len(2)][cookie(4)]
  for ($start = 0; $start -le $pkt.Length - 20; $start++) {
    if ($pkt[$start + 4] -ne $cookie[0]) { continue }
    if ($pkt[$start + 5] -ne $cookie[1]) { continue }
    if ($pkt[$start + 6] -ne $cookie[2]) { continue }
    if ($pkt[$start + 7] -ne $cookie[3]) { continue }

    $msgType = ReadU16BE $pkt $start
    $msgLen = ReadU16BE $pkt ($start + 2)
    if ($DebugParse) {
      $b0 = $pkt[$start]
      $b1 = $pkt[$start+1]
      $b2 = $pkt[$start+2]
      $b3 = $pkt[$start+3]
      $b4 = $pkt[$start+4]
      $b5 = $pkt[$start+5]
      $b6 = $pkt[$start+6]
      $b7 = $pkt[$start+7]
      Write-Host ("[DEBUG] STUN candidate start={0} bytes={1:x2} {2:x2} {3:x2} {4:x2} {5:x2} {6:x2} {7:x2} {8:x2} type=0x{9:x4} len={10}" -f $start, $b0,$b1,$b2,$b3,$b4,$b5,$b6,$b7,$msgType,$msgLen) -ForegroundColor DarkGray
    }
    # STUN type must have the two MSBs = 0
    if (($msgType -band 0xC000) -ne 0) {
      if ($DebugParse) { Write-Host "[DEBUG] reject: type MSBs" -ForegroundColor DarkGray }
      continue
    }
    if (($msgLen % 4) -ne 0) { continue }
    if ($start + 20 + $msgLen -gt $pkt.Length) { continue }

    $slice = SliceBytes $pkt $start (20 + $msgLen)
    $stun = ParseStunMessage $slice
    if ($stun) {
      $out.Add($stun) | Out-Null
      $start = $start + 20 + $msgLen - 1
    } elseif ($DebugParse) {
      Write-Host "[DEBUG] reject: ParseStunMessage returned null" -ForegroundColor DarkGray
    }
  }
  return ,$out.ToArray()
}

if (-not (Test-Path $InputTxt)) { throw "Input file not found: $InputTxt" }

$pktBytesWords = New-Object System.Collections.Generic.List[string]
$inPacket = $false
$wantPacket = $false
$packetsParsed = 0
$foundStun = 0
$seen = @{}
$ipSummary = ""

$pktTs = ""
$pktGroupId = ""
$pktNumber = ""

$export = New-Object System.Collections.Generic.List[object]

function ToHexType([int]$t) {
  return ("0x{0:x4}" -f $t)
}

function AttrValueToHex([byte[]]$b) {
  if (-not $b) { return "" }
  return (BytesToHex $b 1048576)
}

function U32ToIPv4([int]$u) {
  $b0 = ($u -shr 24) -band 0xFF
  $b1 = ($u -shr 16) -band 0xFF
  $b2 = ($u -shr 8) -band 0xFF
  $b3 = $u -band 0xFF
  return ("{0}.{1}.{2}.{3}" -f $b0, $b1, $b2, $b3)
}

function TryDecodeXorAddrAttr([int]$attrType, [int]$attrLen, [byte[]]$attrValue) {
  # WhatsApp Desktop TURN Allocate packets include attr 0x0016 with len=8.
  # This matches XOR-ADDRESS encoding (family + xorPort + xorIPv4) using the STUN magic cookie.
  if ($attrType -ne 0x0016) { return $null }
  if (-not $attrValue -or $attrLen -ne 8 -or $attrValue.Length -lt 8) { return $null }

  $family = ReadU16BE $attrValue 0
  if ($family -ne 0x0001) { return $null } # IPv4 only

  $xorPort = ReadU16BE $attrValue 2
  $port = $xorPort -bxor 0x2112

  $xorAddr = ReadU32BE $attrValue 4
  $addr = $xorAddr -bxor 0x2112A442

  $ip = U32ToIPv4 $addr
  return [PSCustomObject]@{ kind = "xor-addr"; family = "ipv4"; endpoint = ("{0}:{1}" -f $ip, $port) }
}

$onlyTypeInt = $null
if ($OnlyMsgTypeHex -and $OnlyMsgTypeHex.Trim() -ne "") {
  $s = $OnlyMsgTypeHex.Trim().ToLower()
  if ($s.StartsWith("0x")) { $s = $s.Substring(2) }
  $onlyTypeInt = [Convert]::ToInt32($s, 16)
}

$onlyTypeLabel = ""
if ($onlyTypeInt -ne $null) {
  $onlyTypeLabel = ToHexType $onlyTypeInt
}

function Flush-Packet {
  if (-not $inPacket) { return }
  if (-not $wantPacket) { $pktBytesWords.Clear(); return }
  if ($script:packetsParsed -ge $MaxPackets) { return }

  $words = $pktBytesWords.ToArray()
  $pktBytesWords.Clear()

  if ($words.Count -lt 8) { return }
  $pkt = HexToBytes $words
  if ($DebugParse -and $script:packetsParsed -eq 0) {
    $hexAll = ($pkt | ForEach-Object { $_.ToString('x2') }) -join ''
    $hasCookie = ($hexAll -match '2112a442')
    Write-Host ("[DEBUG] bytes={0} words={1} wantPacket={2} hasCookie={3} ip='{4}'" -f $pkt.Length, $words.Count, $wantPacket, $hasCookie, $ipSummary) -ForegroundColor DarkGray
  }
  $stuns = FindStunMessagesInPacket $pkt
  if (-not $stuns -or $stuns.Count -eq 0) { return }

  foreach ($stun in $stuns) {
    if ($script:packetsParsed -ge $MaxPackets) { break }

    $hasMI = $false
    $username = $null
    $realm = $null
    $nonce = $null
    $errorCode = $null

    foreach ($a in $stun.Attrs) {
      switch ($a.Type) {
        0x0006 { $username = $a.Value }
        0x0014 { $realm = $a.Value }
        0x0015 { $nonce = $a.Value }
        0x0008 { $hasMI = $true }
        0x0009 {
          if ($a.Len -ge 4) {
            $cl = $a.Value[2]
            $num = $a.Value[3]
            $errorCode = ($cl * 100) + $num
          }
        }
      }
    }

    $dedupeKey = "$($stun.TxID):$($stun.Type):$hasMI"
    if ($script:seen.ContainsKey($dedupeKey)) { continue }
    $script:seen[$dedupeKey] = $true

    $script:foundStun++
    $script:packetsParsed++

    $unameAscii = BytesToAsciiSafe $username
    $unameHex = BytesToHex $username 128
    $realmAscii = BytesToAsciiSafe $realm
    $nonceHex = BytesToHex $nonce 64

    $errShow = "-"
    if ($null -ne $errorCode) { $errShow = "$errorCode" }
    if (-not $Quiet) {
      Write-Host ("STUN msgType=0x{0:x4} len={1} txid={2} mi={3} err={4} ip='{5}'" -f $stun.Type, $stun.Len, $stun.TxID, $hasMI, $errShow, $ipSummary)
      if ($username) {
        if ($unameAscii -ne "" -and $ShowRawUsername) {
          Write-Host ("  USERNAME ascii='{0}' len={1}" -f $unameAscii, $username.Length)
        } else {
          Write-Host ("  USERNAME hex={0} len={1}" -f $unameHex, $username.Length)
        }
      }
      if ($realm) {
        if ($realmAscii -ne "") {
          Write-Host ("  REALM ascii='{0}' len={1}" -f $realmAscii, $realm.Length)
        } else {
          Write-Host ("  REALM hex={0} len={1}" -f (BytesToHex $realm 128), $realm.Length)
        }
      }
      if ($nonce) {
        Write-Host ("  NONCE hex={0} len={1}" -f $nonceHex, $nonce.Length)
      }

      if ($ShowUnknownAttrs) {
        foreach ($a in $stun.Attrs) {
          if ($a.Type -in 0x0006, 0x0014, 0x0015, 0x0008, 0x0009) { continue }
          Write-Host ("  ATTR type=0x{0:x4} len={1} val_hex_prefix={2}" -f $a.Type, $a.Len, (BytesToHex $a.Value 32))
        }
      }
    }

    if ($OutJson -and $OutJson.Trim() -ne "") {
      if ($onlyTypeInt -ne $null -and $stun.Type -ne $onlyTypeInt) {
        continue
      }
      $attrsJson = @()
      foreach ($a in $stun.Attrs) {
        $decoded = TryDecodeXorAddrAttr $a.Type $a.Len $a.Value
        $attrsJson += [PSCustomObject]@{
          type = ToHexType $a.Type
          len  = $a.Len
          hex  = AttrValueToHex $a.Value
          decoded = $decoded
        }
      }
      $export.Add([PSCustomObject]@{
        msg_type = ToHexType $stun.Type
        len      = $stun.Len
        txid     = $stun.TxID
        mi       = $hasMI
        err      = $errShow
        pkt_ts   = $script:pktTs
        pkt_group_id = $script:pktGroupId
        pkt_number = $script:pktNumber
        ip       = $ipSummary
        attrs    = $attrsJson
      }) | Out-Null
    }
  }
}

$reader = New-Object System.IO.StreamReader($InputTxt)
try {
  while (-not $reader.EndOfStream) {
    $line = $reader.ReadLine()

    # Start of a new packet record
    if ($line -match '^(?<ts>\d{2}:\d{2}:\d{2}\.\d+)\s+PktGroupId\s+(?<gid>\d+),\s+PktNumber\s+(?<pnum>\d+),') {
      if ($inPacket) { Flush-Packet }
      $inPacket = $true
      $wantPacket = $false
      $ipSummary = ""

      $script:pktTs = $Matches['ts']
      $script:pktGroupId = $Matches['gid']
      $script:pktNumber = $Matches['pnum']
      continue
    }

    if (-not $inPacket) { continue }

    # IP summary line tells us ports/proto quickly
    if ($line -match '^\s*(IP6?)\s+(.+)$') {
      $ipSummary = $Matches[0].Trim()
      # pktmon may output protocol names as 'UDP' or 'udp' depending on version/locale.
      if ($ipSummary -match '(?i)\.3478' -and $ipSummary -match '(?i)\budp\b') { $wantPacket = $true }
      elseif ($ipSummary -match '(?i)\.443' -and $ipSummary -match '(?i)\budp\b') { $wantPacket = $true }
      else { $wantPacket = $false }
      continue
    }

    # Packet hex line
    if ($wantPacket -and ($line -match '^\s*0x[0-9a-fA-F]{4}:\s+(.*)$')) {
      $rest = $Matches[1]
      $parts = $rest -split '\s+'
      foreach ($p in $parts) {
        if ($p -match '^[0-9A-Fa-f]{4}$') { $pktBytesWords.Add($p) | Out-Null }
      }
      continue
    }

    if ($packetsParsed -ge $MaxPackets) { break }
  }

  if ($inPacket) { Flush-Packet }
}
finally {
  $reader.Dispose()
}

Write-Host "Done. STUN messages found: $foundStun" -ForegroundColor Green

if ($OutJson -and $OutJson.Trim() -ne "") {
  $outPath = $OutJson
  if (-not [System.IO.Path]::IsPathRooted($outPath)) {
    $outPath = Join-Path (Get-Location) $outPath
  }
  $payload = [PSCustomObject]@{
    input = $InputTxt
    only_msg_type = $onlyTypeLabel
    exported = $export.Count
    items = $export
  }
  $json = $payload | ConvertTo-Json -Depth 6
  [System.IO.File]::WriteAllText($outPath, $json, [Text.Encoding]::UTF8)
  Write-Host "Saved JSON: $outPath (items=$($export.Count))" -ForegroundColor Green
}
