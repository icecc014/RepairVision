$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
if (-not (Test-Path '.env')) {
  Copy-Item '.env.example' '.env'
  Write-Host '已从 .env.example 生成 .env，请确认 MYSQL_ROOT_PASSWORD 等敏感项后再启动。' -ForegroundColor Yellow
}
docker compose up -d --build

# 启动后打印入口地址（含局域网 IP，方便手机访问）
$lanIp = (Get-NetIPAddress -AddressFamily IPv4 -ErrorAction SilentlyContinue |
  Where-Object { $_.IPAddress -notlike '127.*' -and $_.IPAddress -notlike '169.254.*' -and $_.PrefixOrigin -ne 'WellKnown' } |
  Select-Object -First 1 -ExpandProperty IPAddress)
$adminPort = if ($env:WEB_PORT) { $env:WEB_PORT } else { '8080' }
$publicPort = if ($env:PUBLIC_PORT) { $env:PUBLIC_PORT } else { '8081' }
Write-Host ''
Write-Host '本机访问入口：' -ForegroundColor Cyan
Write-Host ('  统一入口 http://localhost:{0}/        （按角色自动跳转）' -f $adminPort)
Write-Host ('  管理端   http://localhost:{0}/admin/' -f $adminPort)
Write-Host ('  宿管端   http://localhost:{0}/dorm' -f $adminPort)
Write-Host ('  工人端   http://localhost:{0}/worker' -f $adminPort)
Write-Host ('  公共报修 http://localhost:{0}/report    （免登录，学生/教师用）' -f $publicPort)
if ($lanIp) {
  Write-Host ''
  Write-Host '手机访问（同一 WiFi，需放行端口）：' -ForegroundColor Cyan
  Write-Host ('  宿管端   http://{0}:{1}/dorm' -f $lanIp, $publicPort)
  Write-Host ('  工人端   http://{0}:{1}/worker' -f $lanIp, $publicPort)
  Write-Host ('  管理端   http://{0}:{1}/admin/' -f $lanIp, $adminPort)
}
