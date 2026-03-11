Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$RepoRoot = Split-Path -Parent $PSScriptRoot
$OutLog = Join-Path $RepoRoot ".dist\wa_cdp_console.log"
$ErrLog = Join-Path $RepoRoot ".dist\wa_cdp_console.err.log"

Get-CimInstance Win32_Process |
    Where-Object { $_.CommandLine -match 'wa_cdp_attach\.py' } |
    ForEach-Object {
        try {
            Stop-Process -Id $_.ProcessId -Force -ErrorAction Stop
        } catch {
        }
    }

Start-Sleep -Milliseconds 500

New-Item -ItemType Directory -Force -Path (Split-Path -Parent $OutLog) | Out-Null
Set-Content -Path $OutLog -Value $null
Set-Content -Path $ErrLog -Value $null

$py = Get-Command python -ErrorAction Stop
Start-Process `
    -FilePath $py.Source `
    -ArgumentList @('-u', 'wa_cdp_attach.py') `
    -WorkingDirectory $RepoRoot `
    -RedirectStandardOutput $OutLog `
    -RedirectStandardError $ErrLog `
    -WindowStyle Hidden

Start-Sleep -Seconds 2

$proc = Get-CimInstance Win32_Process |
    Where-Object { $_.CommandLine -match 'wa_cdp_attach\.py' } |
    Select-Object -First 1

if (-not $proc) {
    throw "wa_cdp_attach.py did not stay running"
}

Write-Host "started wa_cdp_attach.py pid=$($proc.ProcessId)"
Write-Host "stdout: $OutLog"
Write-Host "stderr: $ErrLog"
