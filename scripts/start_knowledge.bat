@echo off
title knowledge-service :8081

set ROOT=%~dp0..
echo ========================================
echo   knowledge-service
echo   Port: 8081
echo ========================================
echo.
cd /d "%ROOT%\knowledge-service"
go run main.go
pause
