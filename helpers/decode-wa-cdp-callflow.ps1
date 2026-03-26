param(
    [string]$LogPath = ".dist\wa_cdp_page_console.log"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$srcDir = Join-Path $root "src"
$decoder = "..\helpers\decode-wa-cdp-log.go"

$pattern = '<ack class="call"|<call |<call>|<receipt |<transport |<relaylatency|<accept |<preaccept|<terminate|<mute_v2'

Push-Location $srcDir
try {
    go run $decoder "..\$LogPath" | Select-String -Pattern $pattern
} finally {
    Pop-Location
}
