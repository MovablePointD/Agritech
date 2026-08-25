@echo off
title rxtcloud - Build All Services
set ROOT=%~dp0..
set OUTDIR=%ROOT%\bin

if not exist "%OUTDIR%" mkdir "%OUTDIR%"

echo ========================================
echo   Building all services...
echo ========================================
echo.

echo [1/5] Building common-service...
cd /d "%ROOT%\common"
go build -o "%OUTDIR%\common-service.exe" .
if %errorlevel% neq 0 (echo   FAILED & goto end) else echo   OK

echo [2/5] Building knowledge-service...
cd /d "%ROOT%\knowledge-service"
go build -o "%OUTDIR%\knowledge-service.exe" .
if %errorlevel% neq 0 (echo   FAILED & goto end) else echo   OK

echo [3/5] Building message-service...
cd /d "%ROOT%\message-service"
go build -o "%OUTDIR%\message-service.exe" .
if %errorlevel% neq 0 (echo   FAILED & goto end) else echo   OK

echo [4/5] Building post-service...
cd /d "%ROOT%\post-service"
go build -o "%OUTDIR%\post-service.exe" .
if %errorlevel% neq 0 (echo   FAILED & goto end) else echo   OK

echo [5/5] Building product-service...
cd /d "%ROOT%\product-service"
go build -o "%OUTDIR%\product-service.exe" .
if %errorlevel% neq 0 (echo   FAILED & goto end) else echo   OK

echo.
echo ========================================
echo   Build complete!
echo   Output: %OUTDIR%
echo ========================================

:end
pause
