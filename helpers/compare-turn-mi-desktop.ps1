param(
    [string]$DesktopPath = "z:\Desenvolvimento\nocodeleaks-quepasa\.dist\pcaps\wa_desktop_20260309_172851_stun_0x0003.json",
    [string]$OfferPath = "z:\Desenvolvimento\nocodeleaks-quepasa\.dist\call_dumps\call_offer_received_20260311163737_ACA6BEE27B774535E8A6A4313BE86FB3.json",
    [string]$EncPlainPath = "z:\Desenvolvimento\nocodeleaks-quepasa\.dist\call_dumps\call_offer_enc_plain_20260311163737_ACA6BEE27B774535E8A6A4313BE86FB3.json",
    [string]$ProbePath = "z:\Desenvolvimento\nocodeleaks-quepasa\.dist\call_dumps\call_turn_probe_20260311163738_ACA6BEE27B774535E8A6A4313BE86FB3.json",
    [string]$Endpoint = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Normalize-Base64([string]$s) {
    if ([string]::IsNullOrWhiteSpace($s)) { return "" }
    $v = $s.Trim().Replace('-', '+').Replace('_', '/')
    while (($v.Length % 4) -ne 0) { $v += "=" }
    return $v
}

function HexTo-Bytes([string]$hex) {
    if ([string]::IsNullOrWhiteSpace($hex)) { return [byte[]]@() }
    $clean = ($hex -replace '\s+', '').ToLowerInvariant()
    if (($clean.Length % 2) -ne 0) { throw "invalid hex length: $($clean.Length)" }
    $bytes = New-Object byte[] ($clean.Length / 2)
    for ($i = 0; $i -lt $bytes.Length; $i++) {
        $bytes[$i] = [Convert]::ToByte($clean.Substring($i * 2, 2), 16)
    }
    return $bytes
}

function BytesTo-Hex([byte[]]$bytes) {
    if ($null -eq $bytes) { return "" }
    return ([System.BitConverter]::ToString($bytes)).Replace("-", "").ToLowerInvariant()
}

function Concat-Bytes([byte[][]]$parts) {
    $ms = New-Object System.IO.MemoryStream
    foreach ($part in $parts) {
        if ($null -eq $part) { continue }
        [void]$ms.Write($part, 0, $part.Length)
    }
    return $ms.ToArray()
}

function Sha1-Bytes([byte[]]$data) {
    $sha1 = [System.Security.Cryptography.SHA1]::Create()
    try { return $sha1.ComputeHash($data) } finally { $sha1.Dispose() }
}

function HmacSha1-Bytes([byte[]]$key, [byte[]]$data) {
    $hmac = New-Object System.Security.Cryptography.HMACSHA1
    try {
        $hmac.Key = $key
        return $hmac.ComputeHash($data)
    } finally {
        $hmac.Dispose()
    }
}

function HmacSha256-Bytes([byte[]]$key, [byte[]]$data) {
    $hmac = New-Object System.Security.Cryptography.HMACSHA256
    try {
        $hmac.Key = $key
        return $hmac.ComputeHash($data)
    } finally {
        $hmac.Dispose()
    }
}

function HKDF-Sha256([byte[]]$ikm, [byte[]]$salt, [byte[]]$info, [int]$length) {
    if ($null -eq $ikm) { $ikm = [byte[]]@() }
    if ($null -eq $salt) { $salt = New-Object byte[] 32 }
    if ($null -eq $info) { $info = [byte[]]@() }
    if ($length -le 0) { return [byte[]]@() }

    $prk = HmacSha256-Bytes $salt $ikm
    $okm = New-Object System.Collections.Generic.List[byte]
    $t = [byte[]]@()
    $counter = 1
    while ($okm.Count -lt $length) {
        $input = Concat-Bytes @($t, $info, [byte[]]@([byte]$counter))
        $t = HmacSha256-Bytes $prk $input
        foreach ($b in $t) {
            if ($okm.Count -ge $length) { break }
            $okm.Add($b)
        }
        $counter++
    }
    return $okm.ToArray()
}

function UInt16BE-Bytes([int]$v) {
    return [byte[]]@(
        [byte](($v -shr 8) -band 0xff),
        [byte]($v -band 0xff)
    )
}

function UInt32BE-Bytes([int]$v) {
    return [byte[]]@(
        [byte](($v -shr 24) -band 0xff),
        [byte](($v -shr 16) -band 0xff),
        [byte](($v -shr 8) -band 0xff),
        [byte]($v -band 0xff)
    )
}

function Get-PaddedAttrBytes($attr) {
    $type = [Convert]::ToInt32(($attr.type -replace '^0x', ''), 16)
    $value = HexTo-Bytes $attr.hex
    $pad = (4 - ($value.Length % 4)) % 4
    $padding = New-Object byte[] $pad
    return Concat-Bytes @(
        (UInt16BE-Bytes $type),
        (UInt16BE-Bytes $value.Length),
        $value,
        $padding
    )
}

function Build-DesktopPreimage($item) {
    $attrsBeforeMi = @()
    $miAttr = $null
    foreach ($attr in $item.attrs) {
        if ($attr.type -eq "0x0008") {
            $miAttr = $attr
            break
        }
        $attrsBeforeMi += ,$attr
    }
    if ($null -eq $miAttr) { throw "desktop item has no MI attr" }

    $attrBytes = @()
    $payloadLen = 0
    foreach ($attr in $attrsBeforeMi) {
        $encoded = Get-PaddedAttrBytes $attr
        $attrBytes += ,$encoded
        $payloadLen += $encoded.Length
    }

    $txid = HexTo-Bytes $item.txid
    $header = Concat-Bytes @(
        (UInt16BE-Bytes ([Convert]::ToInt32(($item.msg_type -replace '^0x', ''), 16))),
        (UInt16BE-Bytes $payloadLen),
        (UInt32BE-Bytes 0x2112A442),
        $txid
    )
    return [pscustomobject]@{
        endpoint = $item.attrs | Where-Object { $_.type -eq "0x0016" } | ForEach-Object { $_.decoded.endpoint } | Select-Object -First 1
        target_mi_hex = $miAttr.hex.ToLowerInvariant()
        preimage = (Concat-Bytes (@($header) + $attrBytes))
        attrs = @($attrsBeforeMi)
    }
}

function Add-LabeledSeed([System.Collections.Generic.List[object]]$list, [string]$label, [byte[]]$bytes) {
    if ($null -eq $bytes -or $bytes.Length -eq 0) { return }
    $list.Add([pscustomobject]@{ label = $label; bytes = $bytes })
}

$desktop = Get-Content -Raw $DesktopPath | ConvertFrom-Json
$offer = Get-Content -Raw $OfferPath | ConvertFrom-Json
$encPlain = Get-Content -Raw $EncPlainPath | ConvertFrom-Json
$probe = Get-Content -Raw $ProbePath | ConvertFrom-Json

if ([string]::IsNullOrWhiteSpace($Endpoint)) {
    $Endpoint = [string]$probe.endpoint
}

$desktopItem = $desktop.items | Where-Object {
    ($_.attrs | Where-Object { $_.type -eq "0x0016" } | Select-Object -First 1).decoded.endpoint -eq $Endpoint
} | Select-Object -First 1
if ($null -eq $desktopItem) {
    throw "desktop item not found for endpoint $Endpoint"
}

$desktopMsg = Build-DesktopPreimage $desktopItem
$desktopPreimage = [byte[]]$desktopMsg.preimage
$targetMiHex = [string]$desktopMsg.target_mi_hex
$targetMi = HexTo-Bytes $targetMiHex

$seedList = New-Object 'System.Collections.Generic.List[object]'

$f10 = HexTo-Bytes $encPlain.parsed_fields.'proto.f10.f1'.hex
Add-LabeledSeed $seedList 'enc.plain.proto.f10.f1' $f10

foreach ($auth in @($offer.relay_block.Auth)) {
    $id = [string]$auth.ID
    $authHead4 = HexTo-Bytes $auth.Head4Hex
    $authLeft32 = HexTo-Bytes $auth.Left32Hex
    $authRight32 = HexTo-Bytes $auth.Right32Hex

    Add-LabeledSeed $seedList ("authHead4(id=$id)") $authHead4
    Add-LabeledSeed $seedList ("authLeft32(id=$id)") $authLeft32
    Add-LabeledSeed $seedList ("authRight32(id=$id)") $authRight32
    Add-LabeledSeed $seedList ("authLeft32+Right32(id=$id)") (Concat-Bytes @($authLeft32, $authRight32))
    Add-LabeledSeed $seedList ("authRight32+Left32(id=$id)") (Concat-Bytes @($authRight32, $authLeft32))
    Add-LabeledSeed $seedList ("authHead4+Left32(id=$id)") (Concat-Bytes @($authHead4, $authLeft32))
    Add-LabeledSeed $seedList ("authHead4+Right32(id=$id)") (Concat-Bytes @($authHead4, $authRight32))
    Add-LabeledSeed $seedList ("authHead4+Left32+Right32(id=$id)") (Concat-Bytes @($authHead4, $authLeft32, $authRight32))
    Add-LabeledSeed $seedList ("authLeft32+Head4+Right32(id=$id)") (Concat-Bytes @($authLeft32, $authHead4, $authRight32))
    Add-LabeledSeed $seedList ("f10+authHead4(id=$id)") (Concat-Bytes @($f10, $authHead4))
    Add-LabeledSeed $seedList ("authHead4+f10(id=$id)") (Concat-Bytes @($authHead4, $f10))

    Add-LabeledSeed $seedList ("authHKDF20(left,right,head4,id=$id)") (HKDF-Sha256 $authLeft32 $authRight32 $authHead4 20)
    Add-LabeledSeed $seedList ("authHKDF20(right,left,head4,id=$id)") (HKDF-Sha256 $authRight32 $authLeft32 $authHead4 20)
    Add-LabeledSeed $seedList ("authHKDF20(left,f10,head4,id=$id)") (HKDF-Sha256 $authLeft32 $f10 $authHead4 20)
    Add-LabeledSeed $seedList ("authHKDF20(right,f10,head4,id=$id)") (HKDF-Sha256 $authRight32 $f10 $authHead4 20)
    Add-LabeledSeed $seedList ("authHKDF32(left,right,head4,id=$id)") (HKDF-Sha256 $authLeft32 $authRight32 $authHead4 32)
    Add-LabeledSeed $seedList ("authHKDF32(right,left,head4,id=$id)") (HKDF-Sha256 $authRight32 $authLeft32 $authHead4 32)
    Add-LabeledSeed $seedList ("authHKDF32(left,f10,head4,id=$id)") (HKDF-Sha256 $authLeft32 $f10 $authHead4 32)
    Add-LabeledSeed $seedList ("authHKDF32(right,f10,head4,id=$id)") (HKDF-Sha256 $authRight32 $f10 $authHead4 32)

    foreach ($block in @($auth.Block16Hex)) {
        Add-LabeledSeed $seedList ("authBlock(id=$id)") (HexTo-Bytes $block)
    }
    Add-LabeledSeed $seedList ("f10+authBlock0(id=$id)") (Concat-Bytes @($f10, (HexTo-Bytes $auth.Block16FirstHex)))
    Add-LabeledSeed $seedList ("f10+authBlockLast(id=$id)") (Concat-Bytes @($f10, (HexTo-Bytes $auth.Block16LastHex)))
    Add-LabeledSeed $seedList ("f10+authLeft32(id=$id)") (Concat-Bytes @($f10, $authLeft32))
    Add-LabeledSeed $seedList ("f10+authRight32(id=$id)") (Concat-Bytes @($f10, $authRight32))
}

foreach ($token in @($offer.relay_block.Tokens)) {
    $id = [string]$token.ID
    $tokHead = HexTo-Bytes $token.EnvelopeHeadHex
    $tokBlock0 = HexTo-Bytes $token.Block16FirstHex
    $tokBlockLast = HexTo-Bytes $token.Block16LastHex
    Add-LabeledSeed $seedList ("tokHead(id=$id)") $tokHead
    Add-LabeledSeed $seedList ("tokBlock0(id=$id)") $tokBlock0
    Add-LabeledSeed $seedList ("tokBlockLast(id=$id)") $tokBlockLast
    Add-LabeledSeed $seedList ("tokHead+Block0(id=$id)") (Concat-Bytes @($tokHead, $tokBlock0))
    Add-LabeledSeed $seedList ("tokHead+BlockLast(id=$id)") (Concat-Bytes @($tokHead, $tokBlockLast))
    Add-LabeledSeed $seedList ("f10+tokBlock0(id=$id)") (Concat-Bytes @($f10, $tokBlock0))
    Add-LabeledSeed $seedList ("f10+tokBlockLast(id=$id)") (Concat-Bytes @($f10, $tokBlockLast))
}

$relayKey = [Convert]::FromBase64String((Normalize-Base64 $offer.relay_block.Key))
$hbhKey = [Convert]::FromBase64String((Normalize-Base64 $offer.relay_block.HBHKey))
Add-LabeledSeed $seedList 'relay.key' $relayKey
Add-LabeledSeed $seedList 'relay.hbh_key' $hbhKey

$seen = @{}
$uniqueSeeds = New-Object 'System.Collections.Generic.List[object]'
foreach ($seed in $seedList) {
    $k = $seed.label + "|" + (BytesTo-Hex $seed.bytes)
    if (-not $seen.ContainsKey($k)) {
        $seen[$k] = $true
        $uniqueSeeds.Add($seed)
    }
}

$results = New-Object 'System.Collections.Generic.List[object]'
foreach ($seed in $uniqueSeeds) {
    $seedBytes = [byte[]]$seed.bytes
    $seedSha1 = Sha1-Bytes $seedBytes
    $msgSha1 = Sha1-Bytes $desktopPreimage

    $derived = @(
        [pscustomobject]@{ name = 'direct'; key = $seedBytes },
        [pscustomobject]@{ name = 'hmac1'; key = (HmacSha1-Bytes $seedBytes $desktopPreimage) },
        [pscustomobject]@{ name = 'hmac1_rev'; key = (HmacSha1-Bytes $desktopPreimage $seedBytes) },
        [pscustomobject]@{ name = 'hmac1_seedSHA1'; key = (HmacSha1-Bytes $seedSha1 $desktopPreimage) },
        [pscustomobject]@{ name = 'hmac1_msgSHA1'; key = (HmacSha1-Bytes $seedBytes $msgSha1) },
        [pscustomobject]@{ name = 'sha1(seed||msg)'; key = (Sha1-Bytes (Concat-Bytes @($seedBytes, $desktopPreimage))) },
        [pscustomobject]@{ name = 'sha1(msg||seed)'; key = (Sha1-Bytes (Concat-Bytes @($desktopPreimage, $seedBytes))) }
    )

    foreach ($d in $derived) {
        $mi = HmacSha1-Bytes ([byte[]]$d.key) $desktopPreimage
        $hex = BytesTo-Hex $mi
        $results.Add([pscustomobject]@{
            seed = $seed.label
            derive = $d.name
            mi_hex = $hex
            match = ($hex -eq $targetMiHex)
        })
    }
}

$matches = @($results | Where-Object { $_.match })

Write-Host ""
Write-Host "desktop_endpoint = $Endpoint"
Write-Host "desktop_target_mi = $targetMiHex"
Write-Host "desktop_preimage_len = $($desktopPreimage.Length)"
Write-Host "tested_variants = $($results.Count)"
Write-Host ""

if ($matches.Count -gt 0) {
    Write-Host "exact_matches:"
    $matches | Format-Table -AutoSize
} else {
    Write-Host "exact_matches: none"
    Write-Host ""
    Write-Host "sample_results:"
    $results | Select-Object -First 16 | Format-Table -AutoSize
}
