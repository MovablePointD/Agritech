@echo off
title rxtcloud - Self-Sign Certificate (local only)

echo ========================================
echo   Self-Signed Certificate Generator
echo   NOTE: Only works on this machine!
echo ========================================
echo.

set CERTNAME=rxtcloud-dev
set PFXPASS=rxtcloud123

echo [1/2] Creating self-signed certificate...
powershell -Command ^
  "$cert = New-SelfSignedCertificate -Type CodeSigningCert -Subject 'CN=rxtcloud-dev' -KeyUsage DigitalSignature -CertStoreLocation 'Cert:\CurrentUser\My'; ^
   $pwd = ConvertTo-SecureString -String '%PFXPASS%' -Force -AsPlainText; ^
   Export-PfxCertificate -Cert $cert -FilePath '%CD%\rxtcloud-dev.pfx' -Password $pwd; ^
   Write-Host 'Certificate thumbprint:' $cert.Thumbprint"

if %errorlevel% neq 0 (
    echo   Certificate creation FAILED
    echo   Try running PowerShell as Administrator
    goto end
)

echo.
echo [2/2] Signing binaries with self-signed cert...
echo   (Requires Windows SDK signtool.exe)
echo.

where signtool >nul 2>&1
if %errorlevel% neq 0 (
    echo   signtool.exe not found!
    echo   Install Windows SDK or run this in Developer Command Prompt
    goto end
)

set ROOT=%~dp0..
set OUTDIR=%ROOT%\bin

if exist "%OUTDIR%\common-service.exe" (
    signtool sign /fd SHA256 /f "%CD%\rxtcloud-dev.pfx" /p %PFXPASS% /tr http://timestamp.digicert.com /td SHA256 "%OUTDIR%\common-service.exe"
    echo   common-service.exe signed
)
if exist "%OUTDIR%\knowledge-service.exe" (
    signtool sign /fd SHA256 /f "%CD%\rxtcloud-dev.pfx" /p %PFXPASS% /tr http://timestamp.digicert.com /td SHA256 "%OUTDIR%\knowledge-service.exe"
    echo   knowledge-service.exe signed
)
if exist "%OUTDIR%\message-service.exe" (
    signtool sign /fd SHA256 /f "%CD%\rxtcloud-dev.pfx" /p %PFXPASS% /tr http://timestamp.digicert.com /td SHA256 "%OUTDIR%\message-service.exe"
    echo   message-service.exe signed
)
if exist "%OUTDIR%\post-service.exe" (
    signtool sign /fd SHA256 /f "%CD%\rxtcloud-dev.pfx" /p %PFXPASS% /tr http://timestamp.digicert.com /td SHA256 "%OUTDIR%\post-service.exe"
    echo   post-service.exe signed
)
if exist "%OUTDIR%\product-service.exe" (
    signtool sign /fd SHA256 /f "%CD%\rxtcloud-dev.pfx" /p %PFXPASS% /tr http://timestamp.digicert.com /td SHA256 "%OUTDIR%\product-service.exe"
    echo   product-service.exe signed
)

echo.
echo ========================================
echo   Done! Now trust this cert once:
echo     certmgr.msc  ->  Trusted Publishers
echo     -> Import rxtcloud-dev.pfx
echo   Then restart services.
echo ========================================

:end
pause
