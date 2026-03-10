param(
    [string]$DumpDir = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\call_dumps",
    [int]$MaxFiles = 20
)

$ErrorActionPreference = 'Stop'

function Get-Json([string]$Path) {
    Get-Content -Path $Path -Raw | ConvertFrom-Json
}

$plainFiles = Get-ChildItem -Path $DumpDir -Filter 'call_offer_enc_plain_*.json' |
    Sort-Object LastWriteTime -Descending |
    Select-Object -First $MaxFiles

if (-not $plainFiles) {
    Write-Host "No call_offer_enc_plain_*.json files found in $DumpDir"
    exit 0
}

$rows = @()
$f10Set = New-Object System.Collections.Generic.HashSet[string]
$f35Set = New-Object System.Collections.Generic.HashSet[string]

foreach ($pf in $plainFiles) {
    $callId = ($pf.BaseName -replace '^call_offer_enc_plain_\d+_', '')
    $j = Get-Json $pf.FullName
    $parsed = $j.parsed_fields

    $f10 = $null
    $f35 = $null
    $f10len = $null
    $f35len = $null

    try { $f10 = $parsed.'proto.f10.f1'.hex } catch {}
    try { $f10len = $parsed.'proto.f10.f1'.len } catch {}
    try { $f35 = $parsed.'proto.f35'.hex } catch {}
    try { $f35len = $parsed.'proto.f35'.len } catch {}

    if ($f10) { [void]$f10Set.Add($f10) }
    if ($f35) { [void]$f35Set.Add($f35) }

    $rows += [pscustomobject]@{
        CallID   = $callId
        PlainLen = $j.plain_len
        F10F1Len = $f10len
        F10F1    = $f10
        F35Len   = $f35len
        F35      = $f35
        File     = $pf.Name
    }
}

$rows | Format-Table CallID, PlainLen, F10F1Len, F35Len, File -AutoSize
Write-Host ""
Write-Host ("unique_f10f1 = {0}" -f $f10Set.Count)
Write-Host ("unique_f35   = {0}" -f $f35Set.Count)
Write-Host ""

$rows |
    Select-Object CallID, F10F1 |
    Format-List
