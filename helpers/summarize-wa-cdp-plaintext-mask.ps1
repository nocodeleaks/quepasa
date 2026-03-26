param(
    [string]$LogPath = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\wa_cdp_page_console.log",
    [ValidateSet('decrypt','encrypt')]
    [string]$Kind = "decrypt",
    [string]$Prefix = "00f80719",
    [int]$Len = 21,
    [int]$Show = 20
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

function Get-NestedDataLen {
    param([string]$Line)
    $m = [regex]::Match($Line, '"data"\s*:\s*\{\s*"len"\s*:\s*([0-9]+)')
    if ($m.Success) { return [int]$m.Groups[1].Value }
    return $null
}

function Get-NestedDataHex {
    param([string]$Line)
    $m = [regex]::Match($Line, '"data"\s*:\s*\{.*?"head_hex"\s*:\s*"([0-9a-fA-F]+)"')
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

$samples = New-Object System.Collections.Generic.List[object]

Get-Content -LiteralPath $LogPath | ForEach-Object {
    $line = $_
    $hex = $null
    $actualLen = $null

    if ($Kind -eq 'decrypt' -and $line -like '*crypto.result.decrypt*') {
        $hex = Get-HexField -Line $line -FieldName 'full_hex'
        if (-not $hex) { $hex = Get-HexField -Line $line -FieldName 'head_hex' }
        $actualLen = if ($hex) { [int]($hex.Length / 2) } else { Get-HexField -Line $line -FieldName 'len' }
    }

    if ($Kind -eq 'encrypt' -and $line -like '*crypto.direct.encrypt*') {
        $hex = Get-NestedDataHex -Line $line
        $actualLen = Get-NestedDataLen -Line $line
    }

    if (-not $hex -or $actualLen -ne $Len -or -not $hex.StartsWith($Prefix.ToLowerInvariant())) {
        return
    }

    $samples.Add([pscustomobject]@{
        Hex   = $hex
        Bytes = (HexToBytes -Hex $hex)
    })
}

if ($samples.Count -eq 0) {
    throw "no samples for kind=$Kind len=$Len prefix=$Prefix"
}

$byteCount = $samples[0].Bytes.Length
$mask = New-Object System.Collections.Generic.List[string]

for ($i = 0; $i -lt $byteCount; $i++) {
    $vals = @($samples | ForEach-Object { $_.Bytes[$i] } | Select-Object -Unique)
    if ($vals.Count -eq 1) {
        $mask.Add(('{0:x2}' -f $vals[0]))
    } else {
        $mask.Add('..')
    }
}

$distinct = $samples | ForEach-Object { $_.Hex } | Sort-Object -Unique

[pscustomobject]@{
    kind            = $Kind
    len             = $Len
    prefix          = $Prefix.ToLowerInvariant()
    sample_count    = $samples.Count
    distinct_values = $distinct.Count
    fixed_mask_hex  = ($mask -join '')
} | Format-List

"examples"
$distinct | Select-Object -First $Show | ForEach-Object { $_ }
