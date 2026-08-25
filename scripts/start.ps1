# =============================================
#  rxtcloud - Single Service Quick Start (PowerShell)
#  Usage: .\start.ps1 <service>
#  e.g.: .\start.ps1 common   -> start common-service
#        .\start.ps1 all      -> start all services
#        .\start.ps1          -> interactive menu
# =============================================

param(
    [Parameter(Mandatory=$false)]
    [string]$Service = ""
)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = Split-Path -Parent $ScriptDir

$services = @{
    "common"    = @{Name="common-service";    Dir="common";            Port=8080; Desc="User/Addr/Expert/Rural/Upload"}
    "knowledge" = @{Name="knowledge-service"; Dir="knowledge-service"; Port=8081; Desc="Knowledge"}
    "message"   = @{Name="message-service";   Dir="message-service";   Port=8082; Desc="Message/Notify"}
    "post"      = @{Name="post-service";      Dir="post-service";      Port=8083; Desc="Post"}
    "product"   = @{Name="product-service";   Dir="product-service";   Port=8084; Desc="Product/Order"}
}

Function Start-ServiceWindow($svc) {
    $dir = Join-Path $RootDir $svc.Dir
    $title = $svc.Name
    
    Write-Host "  -> Starting $($svc.Name) (:$($svc.Port)) $($svc.Desc)..." -ForegroundColor Cyan
    Start-Process powershell -ArgumentList @(
        "-NoExit",
        "-Command",
        "`$host.ui.RawUI.WindowTitle='$title - rxtcloud (:$($svc.Port))'; cd '$dir'; Write-Host '[$title] $($svc.Desc) Starting...' -ForegroundColor Green; Write-Host ''; go run main.go; Read-Host 'Press Enter to exit'"
    )
}

Function Show-Menu {
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  rxtcloud Service Quick Start" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "  Available services:" -ForegroundColor White
    foreach ($key in $services.Keys) {
        $s = $services[$key]
        Write-Host "    $key".PadRight(14) -NoNewline -ForegroundColor Yellow
        Write-Host "$($s.Name)".PadRight(22) -NoNewline -ForegroundColor Cyan
        Write-Host ":$($s.Port)".PadRight(8) -NoNewline -ForegroundColor DarkGray
        Write-Host $s.Desc -ForegroundColor White
    }
    Write-Host "    all".PadRight(14) -NoNewline -ForegroundColor Yellow
    Write-Host "all-services" -ForegroundColor Magenta
    Write-Host ""
}

if ($Service -eq "") {
    Show-Menu
    $Service = Read-Host "Enter service name (common/knowledge/message/post/product/all)"
}

if ($Service -eq "all") {
    Write-Host ""
    Write-Host "Starting all services..." -ForegroundColor Yellow
    foreach ($key in $services.Keys) {
        Start-ServiceWindow $services[$key]
        Start-Sleep -Seconds 2
    }
    Write-Host ""
    Write-Host "All services started!" -ForegroundColor Green
}
elseif ($services.ContainsKey($Service)) {
    Start-ServiceWindow $services[$Service]
}
else {
    Write-Host ""
    Write-Host "Error: Unknown service '$Service'" -ForegroundColor Red
    Write-Host "Valid: common, knowledge, message, post, product, all" -ForegroundColor Yellow
    exit 1
}
