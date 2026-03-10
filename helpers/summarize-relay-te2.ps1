param(
    [string]$DumpDir = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\call_dumps"
)

$ErrorActionPreference = "Stop"

function Convert-BytesToHex {
    param([byte[]]$Bytes)
    if (-not $Bytes) { return "" }
    return (($Bytes | ForEach-Object { $_.ToString("x2") }) -join "")
}

function Convert-HexToBytes {
    param([string]$Hex)
    if ([string]::IsNullOrWhiteSpace($Hex)) { return @() }
    $clean = ($Hex -replace "[^0-9a-fA-F]", "")
    if (($clean.Length % 2) -ne 0) {
        throw "hex with odd length: $Hex"
    }
    $bytes = New-Object byte[] ($clean.Length / 2)
    for ($i = 0; $i -lt $bytes.Length; $i++) {
        $bytes[$i] = [Convert]::ToByte($clean.Substring($i * 2, 2), 16)
    }
    return $bytes
}

function Convert-BytesToIpv6Prefix {
    param([byte[]]$Bytes)
    if (-not $Bytes -or $Bytes.Length -ne 8) { return $null }
    $parts = @()
    for ($i = 0; $i -lt 8; $i += 2) {
        $hi = [int]$Bytes[$i]
        $lo = [int]$Bytes[$i + 1]
        $parts += ("{0:x4}" -f (($hi * 256) + $lo))
    }
    return ($parts -join ":")
}

function Get-JsonProperty {
    param(
        $Object,
        [string]$Name
    )
    if ($null -eq $Object) { return $null }
    $prop = $Object.PSObject.Properties[$Name]
    if ($null -eq $prop) { return $null }
    return $prop.Value
}

$files = Get-ChildItem -Path $DumpDir -Filter "call_offer_received_*.json" -File | Sort-Object LastWriteTime
if (-not $files) {
    Write-Host "no call_offer_received_*.json found in $DumpDir"
    exit 0
}

$rows = New-Object System.Collections.Generic.List[object]

foreach ($file in $files) {
    $json = Get-Content -Path $file.FullName -Raw | ConvertFrom-Json
    $callID = Get-JsonProperty $json "call_id"
    if (-not $callID) { $callID = Get-JsonProperty $json "callID" }

    $dataNode = Get-JsonProperty $json "data"
    if ($null -eq $dataNode) { continue }
    $contentNodes = Get-JsonProperty $dataNode "Content"
    if ($null -eq $contentNodes) { continue }

    $relay = $null
    foreach ($node in $contentNodes) {
        if ((Get-JsonProperty $node "Tag") -eq "relay") {
            $relay = $node
            break
        }
    }
    if ($null -eq $relay) { continue }

    $te2List = @()
    $relayContent = Get-JsonProperty $relay "Content"
    foreach ($entry in $relayContent) {
        if ((Get-JsonProperty $entry "Tag") -eq "te2") {
            $te2List += $entry
        }
    }
    if (-not $te2List.Count) { continue }

    foreach ($entry in $te2List) {
        $entryAttrs = Get-JsonProperty $entry "Attrs"
        $entryContent = Get-JsonProperty $entry "Content"
        $payloadHex = $null
        $payloadB64 = Get-JsonProperty $entryContent "base64"
        if ($payloadB64) {
            try {
                $payloadHex = Convert-BytesToHex ([Convert]::FromBase64String($payloadB64))
            } catch {
                $payloadHex = $null
            }
        }
        if (-not $payloadHex) { continue }

        $payload = Convert-HexToBytes $payloadHex
        if ($payload.Length -ne 18) { continue }

        $prefixBytes = $payload[0..7]
        $markerBytes = $payload[8..11]
        $relayTailBytes = $payload[12..15]
        $suffixBytes = $payload[16..17]

        $rows.Add([pscustomobject]@{
            CallID       = $callID
            RelayName    = (Get-JsonProperty $entryAttrs "relay_name")
            RelayID      = (Get-JsonProperty $entryAttrs "relay_id")
            Protocol     = (Get-JsonProperty $entryAttrs "protocol")
            PayloadHex   = $payloadHex
            IPv6Prefix   = (Convert-BytesToIpv6Prefix $prefixBytes)
            MarkerHex    = (Convert-BytesToHex $markerBytes)
            RelayTailHex = (Convert-BytesToHex $relayTailBytes)
            SuffixHex    = (Convert-BytesToHex $suffixBytes)
        }) | Out-Null
    }
}

if (-not $rows.Count) {
    Write-Host "no te2 payload len=18 found"
    exit 0
}

$rows | Format-Table -AutoSize

$uniquePrefixes = ($rows | Select-Object -ExpandProperty IPv6Prefix -Unique)
$uniqueMarkers = ($rows | Select-Object -ExpandProperty MarkerHex -Unique)
$uniqueSuffixes = ($rows | Select-Object -ExpandProperty SuffixHex -Unique)

Write-Host ""
Write-Host ("unique_ipv6_prefixes = {0}" -f $uniquePrefixes.Count)
Write-Host ("unique_markers      = {0}" -f $uniqueMarkers.Count)
Write-Host ("unique_suffixes     = {0}" -f $uniqueSuffixes.Count)

if ($uniqueMarkers.Count -gt 0) {
    Write-Host ("marker_values       = {0}" -f ($uniqueMarkers -join ", "))
}
if ($uniqueSuffixes.Count -gt 0) {
    Write-Host ("suffix_values       = {0}" -f ($uniqueSuffixes -join ", "))
}
