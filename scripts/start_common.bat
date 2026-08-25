@echo off
title common-service :8080

set ROOT=%~dp0..
echo ========================================
echo   common-service
echo   Port: 8080
echo   Modules: User/Address/Expert/Rural/Upload
echo ========================================
echo.
cd /d "%ROOT%\common"
go run main.go
pause
