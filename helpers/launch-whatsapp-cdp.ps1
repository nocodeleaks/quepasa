param(
  [int]$Port = 9222
)

$repo = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$pkg = Get-AppxPackage *WhatsApp* | Select-Object -First 1
if (-not $pkg) { throw 'WhatsApp AppX package not found' }
$exe = Join-Path $pkg.InstallLocation 'WhatsApp.Root.exe'
if (-not (Test-Path $exe)) { throw "WhatsApp.Root.exe not found at $exe" }

Get-Process WhatsApp.Root -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
$env:WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS = "--remote-debugging-port=$Port"
Start-Process -FilePath $exe
Start-Sleep -Seconds 4
try {
  $json = Invoke-WebRequest -UseBasicParsing ("http://127.0.0.1:{0}/json/list" -f $Port)
  $json.Content
} catch {
  Write-Output $_.Exception.Message
}
