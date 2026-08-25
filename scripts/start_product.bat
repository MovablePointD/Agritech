@echo off
title product-service :8084

set ROOT=%~dp0..
echo ========================================
echo   product-service
echo   Port: 8084
echo ========================================
echo.
cd /d "%ROOT%\product-service"
go run main.go
pause
