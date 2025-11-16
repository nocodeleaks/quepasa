@echo off
REM QuePasa Debug Build Script
REM Always generates the same executable: debug.exe

echo Building QuePasa debug executable...
cd /d "%~dp0src"
go build -o ../dist/debug.exe .
if %errorlevel% equ 0 (
    echo.
    echo Build successful! Executable: dist\debug.exe
    echo Run with: .\dist\debug.exe
    echo Debug with: dlv exec .\dist\debug.exe
) else (
    echo.
    echo Build failed!
)
echo.