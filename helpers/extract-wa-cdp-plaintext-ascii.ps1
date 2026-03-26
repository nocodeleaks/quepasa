param(
    [string]$LogPath = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\wa_cdp_page_console.log",
    [string]$Kind = "decrypt",
    [string]$Prefix = "00f80a09",
    [int]$MinAsciiLen = 4,
    [int]$Top = 80
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

function Convert-HexToBytes {
    param([string]$Hex)
    $bytes = New-Object byte[] ($Hex.Length / 2)
    for ($i = 0; $i -lt $bytes.Length; $i++) {
        $bytes[$i] = [Convert]::ToByte($Hex.Substring($i * 2, 2), 16)
    }
    return $bytes
}

function Get-AsciiRuns {
    param(
        [byte[]]$Bytes,
        [int]$MinLen = 4
    )

    $runs = New-Object System.Collections.Generic.List[string]
    $sb = New-Object System.Text.StringBuilder
    foreach ($b in $Bytes) {
        if ($b -ge 32 -and $b -le 126) {
            [void]$sb.Append([char]$b)
        } else {
            if ($sb.Length -ge $MinLen) {
                $runs.Add($sb.ToString())
            }
            [void]$sb.Clear()
        }
    }
    if ($sb.Length -ge $MinLen) {
        $runs.Add($sb.ToString())
    }
    return $runs
}

$rows = New-Object System.Collections.Generic.List[object]

Get-Content -LiteralPath $LogPath | ForEach-Object {
    $line = $_
    if ($Kind -eq 'decrypt' -and $line -notlike '*crypto.result.decrypt*') { return }
    if ($Kind -eq 'encrypt' -and $line -notlike '*crypto.direct.encrypt*') { return }

    $hex = Get-HexField -Line $line -FieldName 'full_hex'
    if (-not $hex) {
        if ($Kind -eq 'decrypt') {
            $hex = Get-HexField -Line $line -FieldName 'head_hex'
        } else {
            $m = [regex]::Match($line, '"data"\s*:\s*\{.*?"head_hex"\s*:\s*"([0-9a-fA-F]+)"')
            if ($m.Success) {
                $hex = $m.Groups[1].Value.ToLowerInvariant()
            }
        }
    }
    if (-not $hex) { return }
    if (-not $hex.StartsWith($Prefix.ToLowerInvariant())) { return }

    $bytes = Convert-HexToBytes -Hex $hex
    $ascii = Get-AsciiRuns -Bytes $bytes -MinLen $MinAsciiLen
    $rows.Add([pscustomobject]@{
        Hex   = $hex
        ASCII = ($ascii -join ' | ')
    })
}

$rows |
    Sort-Object Hex -Unique |
    Select-Object -First $Top |
    Format-Table ASCII, Hex -Wrap -AutoSize
