param(
    [string]$LogPath = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\wa_cdp_console.log",
    [int]$MaxLen = 256
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
            if ($null -ne $obj.seq -and $null -ne $obj.ts -and $obj.len -le $MaxLen) {
                $rows += [pscustomobject]@{
                    Seq       = [int]$obj.seq
                    Ts        = [int64]$obj.ts
                    Dir       = [string]$dir
                    Len       = [int]$obj.len
                    Prefix3   = [string]$obj.prefix3_hex
                    LenHdr    = $obj.len_hdr
                    Match     = $obj.len_hdr_matches
                    HeadAscii = [string]$obj.head_ascii
                    FullHex   = [string]$obj.full_hex
                    HeadHex   = [string]$obj.head_hex
                    TailHex   = [string]$obj.tail_hex
                }
            }
        } catch {
        }
    }
}

if (-not $rows -or $rows.Count -eq 0) {
    Write-Host "no small websocket frames found"
    exit 0
}

$rows = $rows | Sort-Object Seq

Write-Host "small_frames"
$rows | Select-Object Seq,Dir,Len,Prefix3,HeadAscii | Format-Table -AutoSize

Write-Host ""
Write-Host "families"
$rows |
    Group-Object Dir,Len,Prefix3 |
    Sort-Object Count -Descending |
    ForEach-Object {
        $first = $_.Group[0]
        $unique = ($_.Group | Select-Object -ExpandProperty FullHex -Unique | Measure-Object).Count
        [pscustomobject]@{
            Count          = $_.Count
            Dir            = $first.Dir
            Len            = $first.Len
            Prefix3        = $first.Prefix3
            UniquePayloads = $unique
            FirstSeq       = ($_.Group | Select-Object -First 1 -ExpandProperty Seq)
            LastSeq        = ($_.Group | Select-Object -Last 1 -ExpandProperty Seq)
        }
    } | Format-Table -AutoSize

Write-Host ""
Write-Host "sample_payloads"
$rows |
    Group-Object Dir,Len,Prefix3 |
    Sort-Object Count -Descending |
    Select-Object -First 12 |
    ForEach-Object {
        $first = $_.Group[0]
        $sample = $_.Group | Select-Object -First 3
        Write-Host ("[{0}] dir={1} len={2} prefix={3}" -f $_.Count, $first.Dir, $first.Len, $first.Prefix3)
        $sample | ForEach-Object {
            Write-Host ("  seq={0} full={1}" -f $_.Seq, $_.FullHex)
        }
    }

