param(
    [string]$DumpDir = ""
)

if ([string]::IsNullOrWhiteSpace($DumpDir)) {
    $DumpDir = Join-Path (Split-Path -Parent $PSCommandPath) "..\\.dist\\call_dumps"
}

function To-Hex([byte[]]$bytes) {
    if (-not $bytes -or $bytes.Length -eq 0) { return "" }
    return ([BitConverter]::ToString($bytes)).Replace('-', '').ToLower()
}

function Prefix-Hex([byte[]]$bytes, [int]$nBytes) {
    if (-not $bytes -or $bytes.Length -eq 0) { return "" }
    $n = [Math]::Min($nBytes, $bytes.Length)
    return To-Hex($bytes[0..($n - 1)])
}

function Suffix-Hex([byte[]]$bytes, [int]$nBytes) {
    if (-not $bytes -or $bytes.Length -eq 0) { return "" }
    $n = [Math]::Min($nBytes, $bytes.Length)
    return To-Hex($bytes[($bytes.Length - $n)..($bytes.Length - 1)])
}

function Walk-Nodes($node, [ref]$rows, [string]$callID) {
    if ($null -eq $node) { return }

    if ($node -is [System.Array]) {
        foreach ($item in $node) {
            Walk-Nodes $item ([ref]$rows.Value) $callID
        }
        return
    }

    if (($node.Tag -eq "token") -or ($node.Tag -eq "auth_token")) {
        $raw = [Convert]::FromBase64String([string]$node.Content.base64)
        $headerLen = 0
        if ($node.Tag -eq "token" -and $raw.Length -ge 3 -and $raw[0] -eq 0x09 -and $raw[1] -eq 0x0f -and $raw[2] -eq 0x01) {
            $headerLen = 3
        }
        if ($node.Tag -eq "auth_token" -and $raw.Length -ge 2 -and $raw[0] -eq 0x09 -and $raw[1] -eq 0x03) {
            $headerLen = 2
        }
        $trim = @()
        if ($headerLen -gt 0) {
            $trim = $raw[$headerLen..($raw.Length - 1)]
        }

        $headLen = 0
        if ($trim.Length -gt 0) {
            $headLen = $trim.Length % 16
        }
        $block16Count = 0
        $block16FirstHex = ""
        $block16LastHex = ""
        if ($trim.Length -gt $headLen -and (($trim.Length - $headLen) % 16 -eq 0)) {
            $block16Count = ($trim.Length - $headLen) / 16
            if ($block16Count -gt 0) {
                $block16FirstHex = To-Hex($trim[$headLen..($headLen + 15)])
                $block16LastHex = To-Hex($trim[($trim.Length - 16)..($trim.Length - 1)])
            }
        }

        $rows.Value += [pscustomobject]@{
            CallID          = $callID
            Tag             = $node.Tag
            ID              = [string]$node.Attrs.id
            RawLen          = $raw.Length
            HeaderLen       = $headerLen
            TrimLen         = $trim.Length
            TrimModulo16    = if ($trim.Length -gt 0) { $trim.Length % 16 } else { $null }
            EnvelopeHeadLen = $headLen
            EnvelopeHeadHex = if ($headLen -gt 0) { To-Hex($trim[0..($headLen - 1)]) } else { "" }
            Head3Hex        = if ($trim.Length -ge 3) { To-Hex($trim[0..2]) } else { "" }
            Head4Hex        = if ($trim.Length -ge 4) { To-Hex($trim[0..3]) } else { "" }
            Left32Hex       = if ($trim.Length -eq 68) { To-Hex($trim[4..35]) } else { "" }
            Right32Hex      = if ($trim.Length -eq 68) { To-Hex($trim[36..67]) } else { "" }
            Block16Count    = $block16Count
            Block16FirstHex = $block16FirstHex
            Block16LastHex  = $block16LastHex
            RawPrefixHex    = Prefix-Hex $raw 12
            RawSuffixHex    = Suffix-Hex $raw 12
        }
    }

    if ($null -ne $node.Content) {
        Walk-Nodes $node.Content ([ref]$rows.Value) $callID
    }
}

$files = Get-ChildItem -Path $DumpDir -Filter "call_offer_received_*.json" -File | Sort-Object LastWriteTime -Descending
if (-not $files) {
    Write-Output "no call_offer_received dumps found"
    exit 0
}

$rows = @()
foreach ($file in $files) {
    $json = Get-Content -Path $file.FullName -Raw | ConvertFrom-Json
    $callID = [string]$json.call_id
    Walk-Nodes $json.data ([ref]$rows) $callID
}

if (-not $rows) {
    Write-Output "no token/auth_token nodes found"
    exit 0
}

$rows |
    Sort-Object CallID, Tag, ID |
    Format-Table CallID, Tag, ID, RawLen, HeaderLen, TrimLen, TrimModulo16, EnvelopeHeadLen, EnvelopeHeadHex, Block16Count, Block16FirstHex, Block16LastHex -AutoSize
