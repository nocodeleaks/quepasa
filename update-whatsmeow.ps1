# PowerShell script to update go.mau.fi/whatsmeow to latest version in all modules
# Usage: .\update-whatsmeow.ps1

Write-Host '=== Updating go.mau.fi/whatsmeow to @latest in all modules ===' -ForegroundColor Cyan
Write-Host ''

function Update-QpVersion {
    $defaultsFile = Join-Path $PSScriptRoot 'src\models\qp_defaults.go'
    
    if (Test-Path $defaultsFile) {
        Write-Host 'Updating QpVersion in qp_defaults.go...' -ForegroundColor Cyan
        
        $now = Get-Date
        $year = $now.ToString('yy')
        $date = $now.ToString('MMdd')
        $time = $now.ToString('HHmm')
        $newVersion = "3.$year.$date.$time"
        
        $content = Get-Content $defaultsFile -Raw
        $pattern = 'const QpVersion = "\d+\.\d+\.\d+\.\d+"'
        $replacement = "const QpVersion = `"$newVersion`""
        $newContent = $content -replace $pattern, $replacement
        
        Set-Content -Path $defaultsFile -Value $newContent -NoNewline
        
        Write-Host "  Updated QpVersion to: $newVersion" -ForegroundColor Green
        Write-Host ''
    } else {
        Write-Host "  Warning: qp_defaults.go not found" -ForegroundColor Yellow
        Write-Host ''
    }
}

Update-QpVersion

$modules = @(
    'src',
    'src\api',
    'src\docs',
    'src\environment',
    'src\form',
    'src\library',
    'src\media',
    'src\metrics',
    'src\models',
    'src\rabbitmq',
    'src\sipproxy',
    'src\webserver',
    'src\whatsapp',
    'src\whatsmeow'
)

$rootDir = $PSScriptRoot
$successCount = 0
$failCount = 0

foreach ($module in $modules) {
    $modulePath = Join-Path $rootDir $module
    
    if (Test-Path (Join-Path $modulePath 'go.mod')) {
        Write-Host "Processing module: $module" -ForegroundColor Yellow
        Push-Location $modulePath
        
        go get go.mau.fi/whatsmeow@latest
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  Successfully updated in $module" -ForegroundColor Green
            go mod tidy
            $successCount++
        } else {
            Write-Host "  Failed to update in $module" -ForegroundColor Red
            $failCount++
        }
        
        Pop-Location
        Write-Host ''
    } else {
        Write-Host "Skipping $module (no go.mod found)" -ForegroundColor Gray
        Write-Host ''
    }
}

Write-Host '=== Update Summary ===' -ForegroundColor Cyan
Write-Host "Successfully updated: $successCount modules" -ForegroundColor Green
Write-Host "Failed: $failCount modules" -ForegroundColor Red
Write-Host ''

Write-Host '=== Building the project ===' -ForegroundColor Cyan
Push-Location (Join-Path $rootDir 'src')

go build -o '..\\.dist\\win-quepasa-service.exe'
if ($LASTEXITCODE -eq 0) {
    Write-Host '  Build successful!' -ForegroundColor Green
} else {
    Write-Host '  Build failed!' -ForegroundColor Red
}

Pop-Location
Write-Host ''
Write-Host 'Done!' -ForegroundColor Cyan