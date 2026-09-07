$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
if (-not (Test-Path '.env')) {
  Copy-Item '.env.example' '.env'
  Write-Host '已从 .env.example 生成 .env，请确认 MYSQL_ROOT_PASSWORD 等敏感项后再启动。' -ForegroundColor Yellow
}
docker compose up -d --build