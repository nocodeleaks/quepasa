@echo off
REM Delve wrapper to suppress Go version warning
REM Usage: dlv-wrapper.bat [dlv arguments]

dlv.exe %* 2>&1 | findstr /v "WARNING: undefined behavior"