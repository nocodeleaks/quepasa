param(
    [string]$LogPath = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\wa_cdp_page_console.log",
    [string]$Prefix = "00f80a09",
    [int]$MinAsciiLen = 4,
    [int]$Top = 40
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if (-not (Test-Path -LiteralPath $LogPath)) {
    throw "log not found: $LogPath"
}

function Get-HexField {
    param([string]$Line,[string]$FieldName)
    $rx = '"' + [regex]::Escape($FieldName) + '"\s*:\s*"([0-9a-fA-F]+)"'
    $m = [regex]::Match($Line, $rx)
    if ($m.Success) { return $m.Groups[1].Value.ToLowerInvariant() }
    return $null
}

function HexToBytes {
    param([string]$Hex)
    $bytes = New-Object byte[] ($Hex.Length / 2)
    for ($i = 0; $i -lt $bytes.Length; $i++) {
        $bytes[$i] = [Convert]::ToByte($Hex.Substring($i * 2, 2), 16)
    }
    return $bytes
}

function IsPrintableAscii {
    param([byte[]]$Bytes)
    foreach ($b in $Bytes) {
        if ($b -lt 32 -or $b -gt 126) {
            return $false
        }
    }
    return $true
}

$rows = New-Object System.Collections.Generic.List[object]

Get-Content -LiteralPath $LogPath | ForEach-Object {
    $line = $_
    if ($line -notlike '*crypto.result.decrypt*') { return }
    $hex = Get-HexField -Line $line -FieldName 'full_hex'
    if (-not $hex) { return }
    if (-not $hex.StartsWith($Prefix.ToLowerInvariant())) { return }

    $bytes = HexToBytes -Hex $hex
    for ($i = 0; $i -lt ($bytes.Length - 2); $i++) {
        if ($bytes[$i] -ne 0xfc) { continue }
        $len = [int]$bytes[$i + 1]
        if ($len -lt $MinAsciiLen) { continue }
        $start = $i + 2
        if (($start + $len) -gt $bytes.Length) { continue }
        $seg = $bytes[$start..($start + $len - 1)]
        if (-not (IsPrintableAscii -Bytes $seg)) { continue }
        $ascii = -join ($seg | ForEach-Object { [char]$_ })
        $rows.Add([pscustomobject]@{
            Offset = $i
            Len    = $len
            ASCII  = $ascii
            Hex    = $hex
        })
    }
}

$rows |
    Sort-Object ASCII, Offset, Hex -Unique |
    Select-Object -First $Top |
    Format-Table Offset, Len, ASCII, Hex -Wrap -AutoSize
