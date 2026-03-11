param(
    [string]$DumpDir = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\call_dumps",
    [string]$CallID = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Decode-CompactEndpointHex {
    param([string]$Hex)
    if ($null -eq $Hex) { $Hex = "" }
    $Hex = $Hex.Trim().ToLowerInvariant()
    if ($Hex.Length -ne 12) { return $null }
    if ($Hex -notmatch '^[0-9a-f]{12}$') { return $null }
    $ip = @(
        [Convert]::ToInt32($Hex.Substring(0,2),16),
        [Convert]::ToInt32($Hex.Substring(2,2),16),
        [Convert]::ToInt32($Hex.Substring(4,2),16),
        [Convert]::ToInt32($Hex.Substring(6,2),16)
    ) -join '.'
    $port = [Convert]::ToInt32($Hex.Substring(8,4),16)
    [pscustomobject]@{
        raw_hex  = $Hex
        ip       = $ip
        port     = $port
        endpoint = "$ip`:$port"
    }
}

function Find-LatestCallID {
    param([string]$Dir)
    $f = Get-ChildItem -Path $Dir -Filter "call_offer_received_*.json" |
        Sort-Object LastWriteTimeUtc -Descending |
        Select-Object -First 1
    if (-not $f) { return "" }
    if ($f.Name -match 'call_offer_received_\d+_([A-F0-9]+)\.json$') {
        return $matches[1]
    }
    return ""
}

function Load-Json {
    param([string]$Path)
    if (-not (Test-Path $Path)) { return $null }
    Get-Content -Raw -Path $Path | ConvertFrom-Json
}

function Find-FilesForCall {
    param([string]$Dir, [string]$Id)
    Get-ChildItem -Path $Dir -Filter "*$Id*.json" | Sort-Object Name
}

function Extract-TransportReceivedCompactItems {
    param($Json)
    $out = @()
    if (-not $Json) { return $out }
    $root = $Json.data
    if (-not $root) { return $out }
    function Walk-Nodes {
        param($Nodes)
        foreach ($node in @($Nodes)) {
            if (-not $node) { continue }
            $tag = ("" + $node.Tag).Trim().ToLowerInvariant()
            if (($tag -eq "te" -or $tag -eq "rte") -and $node.Content) {
                $ep = Decode-CompactEndpointHex ("" + $node.Content)
                if ($ep) {
                    $out += [pscustomobject]@{
                        tag      = $tag
                        priority = ("" + $node.Attrs.priority).Trim()
                        ip       = $ep.ip
                        port     = $ep.port
                        endpoint = $ep.endpoint
                        raw_hex  = $ep.raw_hex
                    }
                }
            }
            if ($node.Content -is [System.Collections.IEnumerable] -and -not ($node.Content -is [string])) {
                Walk-Nodes $node.Content
            }
        }
    }
    Walk-Nodes $root.Content
    return $out
}

function Extract-TransportSentCandidates {
    param($Json)
    $out = @()
    if (-not $Json) { return $out }
    function Get-AttrValue {
        param($Attrs, [string]$Name)
        if ($null -eq $Attrs) { return "" }
        $prop = $Attrs.PSObject.Properties[$Name]
        if ($null -eq $prop) { return "" }
        return ([string]$prop.Value).Trim()
    }
    function Walk-Nodes {
        param($Nodes)
        foreach ($node in @($Nodes)) {
            if (-not $node) { continue }
            $tag = ("" + $node.tag).Trim().ToLowerInvariant()
            if ($tag -eq "candidate") {
                $ip = Get-AttrValue $node.attrs "ip"
                $portRaw = Get-AttrValue $node.attrs "port"
                $port = 0
                [void][int]::TryParse($portRaw, [ref]$port)
                $rawHex = ""
                if ($ip -and $port -gt 0) {
                    $bytes = ($ip.Split('.') | ForEach-Object { [byte][int]$_ })
                    if ($bytes.Count -eq 4) {
                        $rawHex = ('{0:x2}{1:x2}{2:x2}{3:x2}{4:x4}' -f $bytes[0],$bytes[1],$bytes[2],$bytes[3],$port)
                    }
                }
                $out += [pscustomobject]@{
                    id       = Get-AttrValue $node.attrs "id"
                    type     = Get-AttrValue $node.attrs "type"
                    ip       = $ip
                    port     = $port
                    endpoint = if ($ip -and $port) { ('{0}:{1}' -f ([string]$ip), ([int]$port)) } else { "" }
                    raw_hex  = $rawHex
                }
            }
            if ($node.content -is [System.Collections.IEnumerable] -and -not ($node.content -is [string])) {
                Walk-Nodes $node.content
            }
        }
    }
    Walk-Nodes $Json.node.content
    return $out
}

if (-not $CallID) {
    $CallID = Find-LatestCallID -Dir $DumpDir
}
if (-not $CallID) {
    throw "could not determine CallID"
}

$files = Find-FilesForCall -Dir $DumpDir -Id $CallID
if (-not $files) {
    throw "no files found for CallID=$CallID"
}

$offer = Load-Json ($files | Where-Object Name -like "call_offer_received_*" | Select-Object -First 1).FullName
$transportRecv = Load-Json ($files | Where-Object Name -like "call_transport_received_*" | Select-Object -First 1).FullName
$transportSent = Load-Json ($files | Where-Object Name -like "call_transport_sent_*" | Select-Object -First 1).FullName
$relayFiles = $files | Where-Object Name -like "call_relaylatency_*"
$relay = @()
foreach ($rf in $relayFiles) {
    $j = Load-Json $rf.FullName
    if ($j -and $j.endpoints) {
        $relay += @($j.endpoints)
    }
}

$recvCompact = Extract-TransportReceivedCompactItems $transportRecv
$sentCandidates = Extract-TransportSentCandidates $transportSent

Write-Output ("CallID: " + $CallID)
Write-Output ""

if ($offer) {
    Write-Output "offer.rte_block"
    $rb = $offer.rte_block
    if ($rb) {
        [pscustomobject]@{
            ipv6_prefix = $rb.IPv6Prefix
            middle4     = $rb.Middle4Hex
            tail4       = $rb.Tail4Hex
            suffix      = $rb.SuffixHex
        } | Format-List | Out-String | Write-Output
    } else {
        Write-Output "(none)"
    }
}

Write-Output "relaylatency.endpoints"
if ($relay.Count -gt 0) {
    $relay | Select-Object relay_name, endpoint, compact_hex, latency_raw | Format-Table -AutoSize | Out-String | Write-Output
} else {
    Write-Output "(none)"
}

Write-Output "transport.received.compact_items"
if ($recvCompact.Count -gt 0) {
    $recvCompact | Format-Table -AutoSize | Out-String | Write-Output
} else {
    Write-Output "(none)"
}

Write-Output "transport.sent.candidates"
if ($sentCandidates.Count -gt 0) {
    $sentCandidates | Format-Table -AutoSize | Out-String | Write-Output
} else {
    Write-Output "(none)"
}

Write-Output "transport.received vs transport.sent (same IP)"
foreach ($item in $recvCompact) {
    $matches = @($sentCandidates | Where-Object { $_.ip -eq $item.ip })
    if ($matches.Count -gt 0) {
        [pscustomobject]@{
            recv_tag         = $item.tag
            recv_endpoint    = $item.endpoint
            recv_raw_hex     = $item.raw_hex
            matching_types   = ($matches.type -join ",")
            matching_ips     = ($matches.endpoint -join ",")
        } | Format-List | Out-String | Write-Output
    }
}
