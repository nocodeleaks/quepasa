param(
    [string]$LogPath = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\wa_cdp_console.log",
    [int]$GapMs = 500
)

if (-not (Test-Path $LogPath)) {
    throw "log not found: $LogPath"
}

$pattern = '\[WA-MON\] WebSocket\.(send|message) (\{.*\})'
$rows = @()

Get-Content -Path $LogPath | ForEach-Object {
    $line = $_
    if ($line -match $pattern) {
        try {
            $dir = $matches[1]
            $obj = $matches[2] | ConvertFrom-Json
            if ($null -ne $obj.seq -and $null -ne $obj.ts) {
                $rows += [pscustomobject]@{
                    Seq      = [int]$obj.seq
                    Ts       = [int64]$obj.ts
                    Dir      = [string]$dir
                    Len      = [int]$obj.len
                    Prefix3  = [string]$obj.prefix3_hex
                    HeadHex  = [string]$obj.head_hex
                }
            }
        } catch {
        }
    }
}

if (-not $rows -or $rows.Count -eq 0) {
    Write-Host "no websocket frames with seq/ts found"
    exit 0
}

$rows = $rows | Sort-Object Ts, Seq

$bursts = @()
$current = New-Object System.Collections.ArrayList
$lastTs = $null

foreach ($row in $rows) {
    if ($null -ne $lastTs -and (($row.Ts - $lastTs) -gt $GapMs) -and $current.Count -gt 0) {
        $bursts += ,@($current.ToArray())
        $current = New-Object System.Collections.ArrayList
    }
    [void]$current.Add($row)
    $lastTs = $row.Ts
}

if ($current.Count -gt 0) {
    $bursts += ,@($current.ToArray())
}

$index = 0
foreach ($burst in $bursts) {
    $index++
    $first = $burst[0]
    $last = $burst[-1]
    $dur = [int64]($last.Ts - $first.Ts)
    $sendCount = ($burst | Where-Object { $_.Dir -eq 'send' }).Count
    $msgCount = ($burst | Where-Object { $_.Dir -eq 'message' }).Count
    $lens = ($burst | Group-Object Dir, Len | Sort-Object Count -Descending | Select-Object -First 6 | ForEach-Object {
        "{0}x{1}@{2}" -f $_.Count, $_.Group[0].Dir, $_.Group[0].Len
    }) -join ", "
    $seqs = "{0}-{1}" -f $first.Seq, $last.Seq
    $heads = ($burst | Select-Object -First 6 | ForEach-Object {
        "{0}:{1}:{2}" -f $_.Seq, $_.Dir, $_.Len
    }) -join " | "

    [pscustomobject]@{
        Burst      = $index
        SeqRange   = $seqs
        StartTs    = $first.Ts
        EndTs      = $last.Ts
        DurationMs = $dur
        Events     = $burst.Count
        Sends      = $sendCount
        Messages   = $msgCount
        TopLens    = $lens
        FirstItems = $heads
    }
}

