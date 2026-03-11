param(
    [string]$DumpDir = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\call_dumps"
)

function Get-HexPrefix {
    param(
        [byte[]]$Bytes,
        [int]$HexChars = 40
    )
    if (-not $Bytes -or $Bytes.Length -eq 0) {
        return ""
    }
    $hex = [System.BitConverter]::ToString($Bytes).Replace('-', '').ToLowerInvariant()
    return $hex.Substring(0, [Math]::Min($HexChars, $hex.Length))
}

function Get-HexSuffix {
    param(
        [byte[]]$Bytes,
        [int]$HexChars = 16
    )
    if (-not $Bytes -or $Bytes.Length -eq 0) {
        return ""
    }
    $hex = [System.BitConverter]::ToString($Bytes).Replace('-', '').ToLowerInvariant()
    return $hex.Substring([Math]::Max(0, $hex.Length - $HexChars))
}

$files = Get-ChildItem -Path $DumpDir -Filter 'call_offer_received_*.json' -ErrorAction SilentlyContinue |
    Sort-Object LastWriteTime -Descending

if (-not $files) {
    Write-Host "no call_offer_received dumps found"
    exit 0
}

$rows = @()

foreach ($file in $files) {
    $json = Get-Content -Path $file.FullName -Raw | ConvertFrom-Json
    if (-not $json.data -or -not $json.data.Content) {
        continue
    }

    $relayNode = $json.data.Content | Where-Object { $_.Tag -eq 'relay' } | Select-Object -First 1
    if (-not $relayNode -or -not $relayNode.Content) {
        continue
    }

    $tokens = @{}
    $auths = @{}

    foreach ($node in $relayNode.Content) {
        if ($node.Tag -eq 'token') {
            $id = [string]$node.Attrs.id
            $raw = [Convert]::FromBase64String([string]$node.Content.base64)
            $tokens[$id] = [pscustomobject]@{
                Len    = $raw.Length
                Prefix = Get-HexPrefix -Bytes $raw
                Suffix = Get-HexSuffix -Bytes $raw
            }
        }

        if ($node.Tag -eq 'auth_token') {
            $id = [string]$node.Attrs.id
            $raw = [Convert]::FromBase64String([string]$node.Content.base64)
            $auths[$id] = [pscustomobject]@{
                Len    = $raw.Length
                Prefix = Get-HexPrefix -Bytes $raw
                Suffix = Get-HexSuffix -Bytes $raw
            }
        }
    }

    $relayEntries = $relayNode.Content |
        Where-Object { $_.Tag -eq 'te2' -and $_.Content -and $_.Content.len -eq 18 } |
        Group-Object { [string]$_.Attrs.relay_id } |
        ForEach-Object { $_.Group | Select-Object -First 1 }

    foreach ($entry in $relayEntries) {
        $relayName = [string]$entry.Attrs.relay_name
        $relayID = [string]$entry.Attrs.relay_id
        $tokenID = [string]$entry.Attrs.token_id
        $authID = [string]$entry.Attrs.auth_token_id

        $rows += [pscustomobject]@{
            CallID      = $json.call_id
            RelayName   = $relayName
            RelayID     = $relayID
            TokenID     = $tokenID
            TokenLen    = if ($tokens.ContainsKey($tokenID)) { $tokens[$tokenID].Len } else { $null }
            TokenPrefix = if ($tokens.ContainsKey($tokenID)) { $tokens[$tokenID].Prefix } else { "" }
            TokenSuffix = if ($tokens.ContainsKey($tokenID)) { $tokens[$tokenID].Suffix } else { "" }
            AuthID      = $authID
            AuthLen     = if ($auths.ContainsKey($authID)) { $auths[$authID].Len } else { $null }
            AuthPrefix  = if ($auths.ContainsKey($authID)) { $auths[$authID].Prefix } else { "" }
            AuthSuffix  = if ($auths.ContainsKey($authID)) { $auths[$authID].Suffix } else { "" }
            File        = $file.Name
        }
    }
}

if (-not $rows) {
    Write-Host "no relay auth/token mappings found"
    exit 0
}

$rows |
    Sort-Object CallID, RelayID |
    Format-Table CallID, RelayName, RelayID, TokenID, TokenLen, TokenPrefix, AuthID, AuthLen, AuthPrefix -AutoSize

Write-Host ""
Write-Host ("unique_token_prefixes = {0}" -f (($rows.TokenPrefix | Sort-Object -Unique).Count))
Write-Host ("unique_auth_prefixes  = {0}" -f (($rows.AuthPrefix | Sort-Object -Unique).Count))
Write-Host ""

$sharedAuth = $rows |
    Group-Object CallID, AuthID, AuthPrefix |
    Where-Object { $_.Count -gt 1 } |
    ForEach-Object {
        $first = $_.Group | Select-Object -First 1
        [pscustomobject]@{
            CallID     = $first.CallID
            AuthID     = $first.AuthID
            AuthPrefix = $first.AuthPrefix
            Relays     = (($_.Group | ForEach-Object { $_.RelayName }) -join ',')
        }
    }

if ($sharedAuth) {
    Write-Host "shared_auth_tokens:"
    $sharedAuth | Format-Table -AutoSize
}
