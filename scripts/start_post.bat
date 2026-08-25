@echo off
title post-service :8083

set ROOT=%~dp0..
echo ========================================
echo   post-service
echo   Port: 8083
echo ========================================
echo.
cd /d "%ROOT%\post-service"
go run main.go
pause
