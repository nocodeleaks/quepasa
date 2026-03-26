param(
    [string]$LogPath = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\wa_cdp_page_console.log",
    [int]$Top = 40
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if (-not (Test-Path -LiteralPath $LogPath)) {
    throw "log not found: $LogPath"
}

function Get-HexField {
    param(
        [string]$Line,
        [string]$FieldName
    )

    $rx = '"' + [regex]::Escape($FieldName) + '"\s*:\s*"([0-9a-fA-F]+)"'
    $m = [regex]::Match($Line, $rx)
    if ($m.Success) {
        return $m.Groups[1].Value.ToLowerInvariant()
    }
    return $null
}

function Get-IntField {
    param(
        [string]$Line,
        [string]$FieldName
    )

    $rx = '"' + [regex]::Escape($FieldName) + '"\s*:\s*([0-9]+)'
    $m = [regex]::Match($Line, $rx)
    if ($m.Success) {
        return [int]$m.Groups[1].Value
    }
    return $null
}

function Get-PrefixHex {
    param(
        [string]$Hex,
        [int]$Bytes = 4
    )

    if ([string]::IsNullOrWhiteSpace($Hex)) {
        return ""
    }
    $chars = [Math]::Min($Hex.Length, $Bytes * 2)
    return $Hex.Substring(0, $chars)
}

function Get-NestedDataLen {
    param(
        [string]$Line
    )

    $m = [regex]::Match($Line, '"data"\s*:\s*\{\s*"len"\s*:\s*([0-9]+)')
    if ($m.Success) {
        return [int]$m.Groups[1].Value
    }
    return $null
}

function Get-NestedDataHex {
    param(
        [string]$Line
    )

    $m = [regex]::Match($Line, '"data"\s*:\s*\{.*?"head_hex"\s*:\s*"([0-9a-fA-F]+)"')
    if ($m.Success) {
        return $m.Groups[1].Value.ToLowerInvariant()
    }
    return $null
}

$decryptGroups = @{}
$encryptGroups = @{}
$interesting = New-Object System.Collections.Generic.List[object]

Get-Content -LiteralPath $LogPath | ForEach-Object {
    $line = $_

    if ($line -like '*crypto.result.decrypt*') {
        $plainHex = Get-HexField -Line $line -FieldName 'full_hex'
        if (-not $plainHex) {
            $plainHex = Get-HexField -Line $line -FieldName 'head_hex'
        }
        $plainLen = Get-IntField -Line $line -FieldName 'len'
        if ($plainHex -and $plainLen -ne $null) {
            $prefix = Get-PrefixHex -Hex $plainHex -Bytes 4
            $key = "$plainLen|$prefix"
            if (-not $decryptGroups.ContainsKey($key)) {
                $decryptGroups[$key] = [pscustomobject]@{
                    Kind    = 'decrypt'
                    Len     = $plainLen
                    Prefix  = $prefix
                    Count   = 0
                    Example = $plainHex
                }
            }
            $decryptGroups[$key].Count++
            if ($prefix.StartsWith('00f8')) {
                $interesting.Add([pscustomobject]@{
                    Kind    = 'decrypt'
                    Len     = $plainLen
                    Prefix  = $prefix
                    PlainHex = $plainHex
                })
            }
        }
    }

    if ($line -like '*crypto.direct.encrypt*') {
        $dataHex = Get-NestedDataHex -Line $line
        $dataLen = Get-NestedDataLen -Line $line
        if ($dataHex -and $dataLen -ne $null) {
            $prefix = Get-PrefixHex -Hex $dataHex -Bytes 4
            $key = "$dataLen|$prefix"
            if (-not $encryptGroups.ContainsKey($key)) {
                $encryptGroups[$key] = [pscustomobject]@{
                    Kind    = 'encrypt'
                    Len     = $dataLen
                    Prefix  = $prefix
                    Count   = 0
                    Example = $dataHex
                }
            }
            $encryptGroups[$key].Count++
            if ($prefix.StartsWith('00f8')) {
                $interesting.Add([pscustomobject]@{
                    Kind    = 'encrypt'
                    Len     = $dataLen
                    Prefix  = $prefix
                    PlainHex = $dataHex
                })
            }
        }
    }
}

$decryptTop = $decryptGroups.Values | Sort-Object -Property @{ Expression = 'Count'; Descending = $true }, @{ Expression = 'Len'; Descending = $false }, @{ Expression = 'Prefix'; Descending = $false } | Select-Object -First $Top
$encryptTop = $encryptGroups.Values | Sort-Object -Property @{ Expression = 'Count'; Descending = $true }, @{ Expression = 'Len'; Descending = $false }, @{ Expression = 'Prefix'; Descending = $false } | Select-Object -First $Top

"decrypt_top"
$decryptTop | Format-Table Count, Len, Prefix, Example -AutoSize

""
"encrypt_top"
$encryptTop | Format-Table Count, Len, Prefix, Example -AutoSize

""
"interesting_00f8"
$interesting | Sort-Object Kind, Len, Prefix, PlainHex -Unique | Format-Table Kind, Len, Prefix, PlainHex -AutoSize
