@echo off
title Seed Demo Data

set ROOT=%~dp0..
echo ========================================
echo   导入演示种子数据
echo   数据库: gorxt_db
echo ========================================
echo.
cd /d "%ROOT%\common"
go run ./cmd/seed
echo.
pause
