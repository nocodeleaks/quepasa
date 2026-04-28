@echo off
REM QuePasa Debug Build Script
REM Always generates the same executable: debug.exe

setlocal

REM Workaround for Go 1.25 on Windows: disable DWARF5 to avoid generating a PE
REM with misaligned .debug_* sections (can fail to execute with ERROR_BAD_EXE_FORMAT).
set "GOEXPERIMENT=nodwarf5"

REM Silence noisy sqlite3 amalgamation warning on GCC (doesn't impact runtime).
set "CGO_CFLAGS=-Wno-return-local-addr"

echo Building QuePasa debug executable...
cd /d "%~dp0src"

if exist "apps\vuejs\package.json" (
    where npm >nul 2>nul
    if %errorlevel% equ 0 (
        echo Building Vue frontend bundle...
        pushd "apps\vuejs"
        if exist package-lock.json (
            call npm ci
        ) else (
            call npm install
        )
        if %errorlevel% neq 0 (
            echo.
            echo Vue frontend build failed!
            popd
            exit /b 1
        )

        call npm run build
        if %errorlevel% neq 0 (
            echo.
            echo Vue frontend build failed!
            popd
            exit /b 1
        )
        popd
    ) else (
        echo npm not found - skipping Vue frontend build.
    )
)

go build -o ..\.dist\debug.exe .
if %errorlevel% equ 0 (
    echo.
    echo Build successful! Executable: .dist\debug.exe
    echo Run with: .\.dist\debug.exe
    echo Debug with: dlv exec .\.dist\debug.exe
) else (
    echo.
    echo Build failed!
)
echo.

endlocal