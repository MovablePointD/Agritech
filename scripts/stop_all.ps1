# =============================================
#  rxtcloud - Stop All Services (PowerShell)
#  Terminate processes by port number
# =============================================

$services = @(
    @{Name="common-service";    Port=8080},
    @{Name="knowledge-service"; Port=8081},
    @{Name="message-service";   Port=8082},
    @{Name="post-service";      Port=8083},
    @{Name="product-service";   Port=8084}
)

Write-Host "========================================" -ForegroundColor Green
Write-Host "  Stopping all services..." -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

foreach ($svc in $services) {
    $i = $services.IndexOf($svc) + 1
    Write-Host "[$i/5] Stopping $($svc.Name) (:$($svc.Port))..." -ForegroundColor Yellow
    
    $conn = netstat -ano | Select-String ":$($svc.Port)" | Select-String "LISTENING"
    foreach ($line in $conn) {
        $parts = $line -split '\s+'
        $pid = $parts[$parts.Length - 1]
        if ($pid -match '^\d+$') {
            Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
            Write-Host "       Process $pid terminated" -ForegroundColor DarkGray
        }
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  All services stopped!" -ForegroundColor Green
Write-Host "  Press any key to exit..." -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
$null = $Host.UI.RawUI.ReadKey('NoEcho,IncludeKeyDown')
