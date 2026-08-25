# =============================================
#  rxtcloud - Start All Services (PowerShell)
#  Opens each service in a separate window
# =============================================

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = Split-Path -Parent $ScriptDir

$services = @(
    @{Name="common-service";    Dir="common";            Port=8080},
    @{Name="knowledge-service"; Dir="knowledge-service"; Port=8081},
    @{Name="message-service";   Dir="message-service";   Port=8082},
    @{Name="post-service";      Dir="post-service";      Port=8083},
    @{Name="product-service";   Dir="product-service";   Port=8084}
)

Write-Host "========================================" -ForegroundColor Green
Write-Host "  rxtcloud Service Launcher (PowerShell)" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

foreach ($svc in $services) {
    $info = "{0,-22} :{1}" -f $svc.Name, $svc.Port
    Write-Host "  [$($services.IndexOf($svc)+1)] $info"
}

Write-Host ""
Write-Host "Starting all services..." -ForegroundColor Yellow
Write-Host ""

foreach ($svc in $services) {
    $dir = Join-Path $RootDir $svc.Dir
    $title = $svc.Name
    
    Write-Host "  -> Starting $($svc.Name)..." -ForegroundColor Cyan
    Start-Process powershell -ArgumentList @(
        "-NoExit",
        "-Command",
        "`$host.ui.RawUI.WindowTitle='$title (:$($svc.Port))'; cd '$dir'; Write-Host '[$title] Starting... (port:$($svc.Port))' -ForegroundColor Green; go run main.go; Write-Host '[$title] Exited' -ForegroundColor Red; Read-Host 'Press Enter'"
    )
    Start-Sleep -Seconds 2
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  All services started!" -ForegroundColor Green
Write-Host "  Press any key to exit..." -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
$null = $Host.UI.RawUI.ReadKey('NoEcho,IncludeKeyDown')
