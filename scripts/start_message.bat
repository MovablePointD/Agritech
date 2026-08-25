@echo off
title message-service :8082

set ROOT=%~dp0..
echo ========================================
echo   message-service
echo   Port: 8082
echo ========================================
echo.
cd /d "%ROOT%\message-service"
go run main.go
pause
