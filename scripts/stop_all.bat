@echo off
title rxtcloud - Stop All Services

echo ========================================
echo   Stopping all services...
echo ========================================
echo.

echo [1/5] Stopping common-service    (8080)...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8080" ^| findstr "LISTENING"') do (
    taskkill /F /PID %%a >nul 2>&1
)

echo [2/5] Stopping knowledge-service (8081)...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8081" ^| findstr "LISTENING"') do (
    taskkill /F /PID %%a >nul 2>&1
)

echo [3/5] Stopping message-service   (8082)...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8082" ^| findstr "LISTENING"') do (
    taskkill /F /PID %%a >nul 2>&1
)

echo [4/5] Stopping post-service      (8083)...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8083" ^| findstr "LISTENING"') do (
    taskkill /F /PID %%a >nul 2>&1
)

echo [5/5] Stopping product-service   (8084)...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8084" ^| findstr "LISTENING"') do (
    taskkill /F /PID %%a >nul 2>&1
)

echo.
echo ========================================
echo   All services stopped!
echo   Press any key to close this window...
echo ========================================
pause >nul
