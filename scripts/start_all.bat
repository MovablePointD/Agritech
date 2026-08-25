@echo off
title rxtcloud - Start All Services

echo ========================================
echo   rxtcloud Service Launcher
echo ========================================
echo.
echo Services:
echo   [1] common-service    :8080  (User/Addr/Expert/Rural)
echo   [2] knowledge-service :8081  (Knowledge)
echo   [3] message-service   :8082  (Message/Notify)
echo   [4] post-service      :8083  (Post)
echo   [5] product-service   :8084  (Product/Order)
echo.
echo Starting all services...

set ROOT=%~dp0..

start "common-service"    cmd /k "cd /d %ROOT%\common && go run main.go"
timeout /t 2 /nobreak >nul
start "knowledge-service" cmd /k "cd /d %ROOT%\knowledge-service && go run main.go"
timeout /t 2 /nobreak >nul
start "message-service"   cmd /k "cd /d %ROOT%\message-service && go run main.go"
timeout /t 2 /nobreak >nul
start "post-service"      cmd /k "cd /d %ROOT%\post-service && go run main.go"
timeout /t 2 /nobreak >nul
start "product-service"   cmd /k "cd /d %ROOT%\product-service && go run main.go"

echo.
echo ========================================
echo   All services started!
echo   Press any key to close this window...
echo ========================================
pause >nul
