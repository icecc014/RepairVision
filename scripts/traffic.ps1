<#
RepairVision 访问量观测（V9.9 抗压配套）
用法：
  powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\traffic.ps1
  powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\traffic.ps1 -Tail 5000
说明：读取 web 容器（nginx）访问日志，汇总请求量、来源 IP、Top 路径与 429/5xx，
     用于演示前后对比；不改动任何服务。
#>
param([int]$Tail = 2000)

$dockerDir = 'C:\Users\28753\AppData\Local\Programs\DockerDesktop\resources\bin'
if (Test-Path $dockerDir) { $env:Path += ';' + $dockerDir }

$lines = docker logs repairvision-web-1 --tail $Tail 2>&1
$re = '^(?<ip>\S+) - - \[(?<t>[^\]]+)\] "(?<m>[A-Z]+) (?<p>\S+)[^"]*" (?<s>\d{3}) (?<b>\d+)'
$rows = foreach ($l in $lines) {
  if ($l -match $re) {
    [pscustomobject]@{
      IP     = $Matches.ip
      T      = [datetime]::ParseExact($Matches.t, 'dd/MMM/yyyy:HH:mm:ss zzz', [Globalization.CultureInfo]::InvariantCulture)
      Method = $Matches.m
      Path   = ($Matches.p -split '\?')[0]
      Status = [int]$Matches.s
      Bytes  = [int]$Matches.b
    }
  }
}

if (-not $rows) { Write-Output '未解析到访问日志（确认 web 容器在运行）'; exit 0 }
$rows = @($rows)

Write-Output ('=== RepairVision 访问观测（最近 {0} 行日志）===' -f $rows.Count)
Write-Output ('时间范围: {0:yyyy-MM-dd HH:mm:ss} ~ {1:yyyy-MM-dd HH:mm:ss}' -f ($rows.T | Measure-Object -Minimum).Minimum, ($rows.T | Measure-Object -Maximum).Maximum)
Write-Output ('总请求数: {0}    独立来源 IP: {1}    下行流量: {2:N1} KB' -f $rows.Count, ($rows.IP | Sort-Object -Unique).Count, (($rows.Bytes | Measure-Object -Sum).Sum / 1KB))

Write-Output ''
Write-Output '--- 状态码分布 ---'
$rows | Group-Object Status | Sort-Object Name | ForEach-Object { '  {0}  x{1}' -f $_.Name, $_.Count }

Write-Output ''
Write-Output '--- 每分钟请求峰值（Top 3）---'
$rows | Group-Object { $_.T.ToString('HH:mm') } | Sort-Object Count -Descending | Select-Object -First 3 | ForEach-Object { '  {0}  {1} 次' -f $_.Name, $_.Count }

Write-Output ''
Write-Output '--- 访问最多的来源 IP（Top 5）---'
$rows | Group-Object IP | Sort-Object Count -Descending | Select-Object -First 5 | ForEach-Object { '  {0,-18} {1} 次' -f $_.Name, $_.Count }

Write-Output ''
Write-Output '--- 请求最多的路径（Top 10）---'
$rows | Group-Object Path | Sort-Object Count -Descending | Select-Object -First 10 | ForEach-Object { '  {0,-42} {1} 次' -f $_.Name, $_.Count }

$rl = @($rows | Where-Object { $_.Status -eq 429 }).Count
$bad = @($rows | Where-Object { $_.Status -ge 500 }).Count
Write-Output ''
Write-Output ('限流命中(429): {0} 次    服务端错误(5xx): {1} 次' -f $rl, $bad)