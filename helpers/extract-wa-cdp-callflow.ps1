param(
    [string]$LogPath = ".dist\wa_cdp_page_console.log",
    [string]$OutPath = ".dist\wa_cdp_callflow.json"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$srcDir = Join-Path $root "src"
$decoder = "..\helpers\extract-wa-cdp-callflow.go"

Push-Location $srcDir
try {
    go run $decoder "..\$LogPath" | Set-Content "..\$OutPath"
    Write-Output $OutPath
} finally {
    Pop-Location
}
