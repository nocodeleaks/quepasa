param(
    [string]$LogPath = "Z:\Desenvolvimento\nocodeleaks-quepasa\.dist\wa_cdp_page_console.log",
    [int[]]$Lens = @(65, 69, 70, 102, 408, 430, 659, 2697),
    [int]$Context = 4
)

if (-not (Test-Path $LogPath)) {
    Write-Host "log not found: $LogPath"
    exit 1
}

$lines = Get-Content -Path $LogPath
$pattern = ($Lens | ForEach-Object { '"len":' + $_ + ',' }) -join '|'
$hits = Select-String -Path $LogPath -Pattern $pattern

if (-not $hits) {
    Write-Host "no matching rare frames found"
    exit 0
}

foreach ($hit in $hits) {
    $line = $hit.Line
    $lineNo = $hit.LineNumber
    $start = [Math]::Max(1, $lineNo - $Context)
    $end = [Math]::Min($lines.Count, $lineNo + $Context)

    Write-Host ("--- line {0} ---" -f $lineNo)
    for ($i = $start; $i -le $end; $i++) {
        Write-Host $lines[$i - 1]
    }
    Write-Host ""
}
