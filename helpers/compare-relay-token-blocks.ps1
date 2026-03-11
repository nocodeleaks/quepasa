param(
    [string]$DumpFile = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Get-Hex([byte[]]$Bytes) {
    return -join ($Bytes | ForEach-Object { $_.ToString("x2") })
}

function Get-TrimmedBytes([byte[]]$Bytes, [int]$TrimPrefixLen) {
    if ($Bytes.Length -le $TrimPrefixLen) {
        return [byte[]]@()
    }
    return $Bytes[$TrimPrefixLen..($Bytes.Length - 1)]
}

function Get-RelayNode($Json) {
    if ($null -eq $Json.data -or $null -eq $Json.data.Content) {
        return $null
    }
    foreach ($node in $Json.data.Content) {
        if ($node.Tag -eq "relay") {
            return $node
        }
    }
    return $null
}

function Get-TokenBlocks([byte[]]$Trimmed, [int]$HeadLen) {
    $result = @()
    if ($Trimmed.Length -le $HeadLen) {
        return ,$result
    }
    $body = $Trimmed[$HeadLen..($Trimmed.Length - 1)]
    for ($i = 0; $i -lt $body.Length; $i += 16) {
        $end = [Math]::Min($i + 15, $body.Length - 1)
        $chunk = $body[$i..$end]
        $result += [pscustomobject]@{
            Index = [int]($i / 16)
            Hex   = Get-Hex $chunk
        }
    }
    return ,$result
}

if ([string]::IsNullOrWhiteSpace($DumpFile)) {
    $root = Split-Path -Parent (Split-Path -Parent $PSCommandPath)
    $DumpFile = Join-Path $root ".dist\call_dumps\call_offer_received_20260310101919_ACE7D3E205963F5E2B0AE120BA359554.json"
}

if (-not (Test-Path -LiteralPath $DumpFile)) {
    throw "dump file not found: $DumpFile"
}

$json = Get-Content -LiteralPath $DumpFile -Raw | ConvertFrom-Json
$relay = Get-RelayNode $json
if ($null -eq $relay) {
    throw "relay node not found"
}

$entries = @()
$mapping = @()

foreach ($node in $relay.Content) {
    if ($node.Tag -eq "token") {
        $id = [string]$node.Attrs.id
        $raw = [Convert]::FromBase64String([string]$node.Content.base64)
        $trim = Get-TrimmedBytes $raw 3
        $blocks = Get-TokenBlocks $trim 3
        foreach ($block in $blocks) {
            $entries += [pscustomobject]@{
                Kind = "token"
                ID   = $id
                Index = $block.Index
                Hex  = $block.Hex
            }
        }
    } elseif ($node.Tag -eq "auth_token") {
        $id = [string]$node.Attrs.id
        $raw = [Convert]::FromBase64String([string]$node.Content.base64)
        $trim = Get-TrimmedBytes $raw 2
        $blocks = Get-TokenBlocks $trim 4
        foreach ($block in $blocks) {
            $entries += [pscustomobject]@{
                Kind = "auth"
                ID   = $id
                Index = $block.Index
                Hex  = $block.Hex
            }
        }
    } elseif ($node.Tag -eq "te2" -and [int]$node.Content.len -eq 18) {
        $mapping += [pscustomobject]@{
            Relay    = [string]$node.Attrs.relay_name
            RelayID  = [string]$node.Attrs.relay_id
            Protocol = [string]$node.Attrs.protocol
            TokenID  = [string]$node.Attrs.token_id
            AuthID   = [string]$node.Attrs.auth_token_id
        }
    }
}

$mapping =
    $mapping |
    Sort-Object Relay, RelayID, Protocol, TokenID, AuthID -Unique

$duplicates =
    $entries |
    Group-Object Hex |
    Where-Object { $_.Count -gt 1 } |
    Sort-Object Name

Write-Host "dump_file = $DumpFile"
Write-Host ""
Write-Host "relay_to_token_auth"
$mapping | Format-Table -AutoSize
Write-Host ""
Write-Host ("total_blocks = {0}" -f $entries.Count)
Write-Host ("duplicate_block_values = {0}" -f $duplicates.Count)
Write-Host ""

if ($duplicates.Count -eq 0) {
    Write-Host "NO_DUP_BLOCKS"
} else {
    $rows = foreach ($dup in $duplicates) {
        [pscustomobject]@{
            Hex  = $dup.Name
            Refs = ($dup.Group | ForEach-Object { "{0}:{1}:{2}" -f $_.Kind, $_.ID, $_.Index }) -join ","
        }
    }
    $rows | Format-Table -AutoSize
}
