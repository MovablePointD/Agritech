@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

echo ==============================================
echo   AI Chat Service Launcher (DeepSeek) v4.0.0
echo ==============================================
echo.

REM --- Check conda ---
where conda >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] conda not found. Please install Anaconda/Miniconda first.
    echo Download from: https://www.anaconda.com/download
    pause
    exit /b 1
)

REM --- Activate conda environment ---
echo [1/4] Checking Python environment...
call conda activate python310 2>nul
if %errorlevel% neq 0 (
    echo [INFO] Environment 'python310' not found, creating...
    call conda create -n python310 python=3.10 -y
    if %errorlevel% neq 0 (
        echo [ERROR] Failed to create conda environment.
        pause
        exit /b 1
    )
    call conda activate python310
)

REM --- Verify python ---
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Failed to activate Python environment.
    pause
    exit /b 1
)

REM --- Install dependencies ---
echo [2/4] Installing dependencies...
pip install -r requirements.txt -q
if %errorlevel% neq 0 (
    echo [ERROR] Failed to install dependencies.
    pause
    exit /b 1
)

REM --- Check API Key ---
echo [3/4] Checking configuration...
if not exist ".env" (
    echo [WARNING] .env file not found, creating default...
    (
        echo # DeepSeek API Key ^(Required^)
        echo # Get from https://platform.deepseek.com/
        echo DEEPSEEK_API_KEY=your_deepseek_api_key_here
        echo.
        echo # Service Config
        echo FLASK_HOST=0.0.0.0
        echo FLASK_PORT=5000
        echo FLASK_DEBUG=true
    ) > .env
)

findstr /C:"your_deepseek_api_key_here" .env >nul 2>&1
if %errorlevel% equ 0 (
    echo.
    echo ==============================================
    echo   [WARNING] DeepSeek API Key not configured!
    echo   Please edit .env and set your API key.
    echo   Get one from https://platform.deepseek.com/
    echo ==============================================
    echo.
    choice /C YN /M "Open .env in Notepad now"
    if !errorlevel! equ 1 start notepad.exe .env
    echo.
    echo Press any key to continue starting service...
    pause >nul
)

REM --- Start service ---
echo.
echo [4/4] Starting AI Chat Service...
echo ==============================================
echo.

python app.py

pause
