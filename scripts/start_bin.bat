@echo off
title rxtcloud - Start from Binaries
set ROOT=%~dp0..
set BINDIR=%ROOT%\bin

if not exist "%BINDIR%" (
    echo bin/ directory not found. Run build_all.bat first.
    pause
    exit /b 1
)

echo ========================================
echo   Starting services from bin/
echo ========================================
echo.

start "common-service"    "%BINDIR%\common-service.exe"
timeout /t 2 /nobreak >nul
start "knowledge-service" "%BINDIR%\knowledge-service.exe"
timeout /t 2 /nobreak >nul
start "message-service"   "%BINDIR%\message-service.exe"
timeout /t 2 /nobreak >nul
start "post-service"      "%BINDIR%\post-service.exe"
timeout /t 2 /nobreak >nul
start "product-service"   "%BINDIR%\product-service.exe"

echo.
echo ========================================
echo   All services started!
echo   Press any key to close...
echo ========================================
pause >nul
