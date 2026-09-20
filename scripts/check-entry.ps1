<#
  RepairVision V9 统一入口验收脚本
  用途：本地或内网穿透后的公网地址，一键核对“统一入口 + 统一登录 + 各端可达”是否达标。
  用法：
    powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\check-entry.ps1
    powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\check-entry.ps1 -BaseUrl https://xxxx.cpolar.cn
    powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\check-entry.ps1 -BaseUrl http://10.1.97.149:8080 -CheckWs
  说明：使用 HttpClient 且关闭自动重定向与系统代理，保证 302 与 WebSocket 检测结果可信。
#>
param(
  [string]$BaseUrl = 'http://localhost:8080',
  [switch]$CheckWs
)

$ErrorActionPreference = 'Continue'
Add-Type -AssemblyName System.Net.Http | Out-Null
$BaseUrl = $BaseUrl.TrimEnd('/')
$script:pass = 0
$script:fail = 0

function Write-Pass([string]$name, [string]$detail) {
  $script:pass++
  Write-Host ("  [通过] {0,-38} {1}" -f $name, $detail) -ForegroundColor Green
}
function Write-Fail([string]$name, [string]$detail) {
  $script:fail++
  Write-Host ("  [失败] {0,-38} {1}" -f $name, $detail) -ForegroundColor Red
}
function Write-Info([string]$name, [string]$detail) {
  Write-Host ("  [信息] {0,-38} {1}" -f $name, $detail) -ForegroundColor DarkGray
}

# 关闭自动重定向与代理：既能拿到 302 本身，也不受系统代理影响
$handler = New-Object System.Net.Http.HttpClientHandler
$handler.AllowAutoRedirect = $false
$handler.UseProxy = $false
$client = New-Object System.Net.Http.HttpClient($handler)
$client.Timeout = [TimeSpan]::FromSeconds(20)

function Invoke-Probe([string]$url, [string]$method = 'GET', [string]$body = $null) {
  try {
    if ($method -eq 'POST') {
      $content = New-Object System.Net.Http.StringContent($body, [System.Text.Encoding]::UTF8, 'application/json')
      $resp = $client.PostAsync($url, $content).GetAwaiter().GetResult()
    } else {
      $resp = $client.GetAsync($url).GetAwaiter().GetResult()
    }
    $loc = ''
    if ($resp.Headers.Location) { $loc = $resp.Headers.Location.ToString() }
    $text = $resp.Content.ReadAsStringAsync().GetAwaiter().GetResult()
    return [pscustomobject]@{ Code = [int]$resp.StatusCode; Body = $text; Location = $loc }
    $resp.Dispose()
  } catch {
    return [pscustomobject]@{ Code = 0; Body = $_.Exception.Message; Location = '' }
  }
}

Write-Host ''
Write-Host ("RepairVision V9 统一入口验收： " + $BaseUrl) -ForegroundColor Cyan
Write-Host ('-' * 74)

# 1) 统一入口：根路径与 /login 都应 302 到统一登录页
foreach ($probe in @(@{ Path = '/'; Name = '根路径 /' }, @{ Path = '/login'; Name = '/login' })) {
  $r = Invoke-Probe ($BaseUrl + $probe.Path)
  if ($r.Code -eq 302 -and $r.Location -match '/login/?$') {
    Write-Pass $probe.Name ('302 -> ' + $r.Location)
  } elseif ($r.Code -eq 200) {
    Write-Pass $probe.Name '200（直接返回统一登录页）'
  } else {
    Write-Fail $probe.Name ('HTTP ' + $r.Code + ' Location=' + $r.Location)
  }
}

# 2) 统一登录页与静态资源
$login = Invoke-Probe ($BaseUrl + '/login/')
if ($login.Code -eq 200 -and $login.Body -match '统一登录') {
  Write-Pass '统一登录页 /login/' '200，内容含「统一登录」'
} else {
  Write-Fail '统一登录页 /login/' ('HTTP ' + $login.Code)
}
foreach ($asset in @('/login/styles.css', '/login/app.js')) {
  $r = Invoke-Probe ($BaseUrl + $asset)
  if ($r.Code -eq 200) { Write-Pass ('登录页资源 ' + $asset) '200' } else { Write-Fail ('登录页资源 ' + $asset) ('HTTP ' + $r.Code) }
}

# 3) 三端入口可达
foreach ($entry in @(@{ Path = '/admin/'; Name = 'PC 管理端 /admin/' }, @{ Path = '/m/'; Name = '移动端 /m/' })) {
  $r = Invoke-Probe ($BaseUrl + $entry.Path)
  if ($r.Code -eq 200) { Write-Pass $entry.Name '200' } else { Write-Fail $entry.Name ('HTTP ' + $r.Code) }
}
foreach ($short in @(@{ Path = '/dorm'; To = '/m/dorm' }, @{ Path = '/worker'; To = '/m/worker' }, @{ Path = '/report'; To = '/m/report' })) {
  $r = Invoke-Probe ($BaseUrl + $short.Path)
  if ($r.Code -eq 302 -and $r.Location -match [regex]::Escape($short.To)) {
    Write-Pass ('短入口 ' + $short.Path) ('302 -> ' + $r.Location)
  } else {
    Write-Fail ('短入口 ' + $short.Path) ('HTTP ' + $r.Code + ' Location=' + $r.Location)
  }
}

# 4) 公共报修接口（免登录）
$pub = Invoke-Probe ($BaseUrl + '/api/public/buildings')
if ($pub.Code -eq 200 -and $pub.Body -match '"code"\s*:\s*0') {
  Write-Pass '公共楼栋接口（免登录）' '200，code=0'
} else {
  Write-Fail '公共楼栋接口（免登录）' ('HTTP ' + $pub.Code)
}

# 5) 统一登录接口 + 角色映射
$cases = @(
  @{ User = 'admin';  Role = 1; Expect = '/admin/' },
  @{ User = 'water1'; Role = 2; Expect = '/m/worker' },
  @{ User = 'dorm1';  Role = 3; Expect = '/m/dorm' }
)
foreach ($case in $cases) {
  $body = '{"username":"' + $case.User + '","password":"admin123"}'
  $r = Invoke-Probe ($BaseUrl + '/api/login') 'POST' $body
  if ($r.Code -eq 200) {
    try {
      $json = $r.Body | ConvertFrom-Json
      if ($json.code -eq 0 -and $json.data.user.role -eq $case.Role -and $json.data.token) {
        Write-Pass ('登录 ' + $case.User) ('role=' + $json.data.user.role + ' -> ' + $case.Expect + '，token ' + $json.data.token.Length + ' 字符')
      } else {
        Write-Fail ('登录 ' + $case.User) ('返回 code=' + $json.code + ' role=' + $json.data.user.role)
      }
    } catch {
      Write-Fail ('登录 ' + $case.User) '响应解析失败'
    }
  } else {
    Write-Fail ('登录 ' + $case.User) ('HTTP ' + $r.Code + ' ' + $r.Body)
  }
}

# 6) 可选：WebSocket 通道（/ws/orders 需 JWT；未带 token 返回 401 即说明 nginx -> order-api:8890 转发链路正常）
if ($CheckWs) {
  $wsUrl = $BaseUrl + '/ws/orders'
  if (-not (Get-Command curl.exe -ErrorAction SilentlyContinue)) {
    Write-Info 'WebSocket /ws/orders' '未找到 curl.exe，跳过该项'
  } else {
    $code = (& curl.exe -s -o NUL -w '%{http_code}' --max-time 10 `
      -H 'Connection: Upgrade' -H 'Upgrade: websocket' `
      -H 'Sec-WebSocket-Version: 13' -H 'Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==' $wsUrl)
    if ($code -eq '101') {
      Write-Pass 'WebSocket /ws/orders' '101 Switching Protocols（握手成功）'
    } elseif ($code -eq '401' -or $code -eq '403') {
      Write-Pass 'WebSocket /ws/orders' ($code + ' 鉴权生效：转发到 ws 服务正常，携带 token 才能建连')
    } else {
      Write-Fail 'WebSocket /ws/orders' ('HTTP ' + $code + '（期望 101 或 401/403；404/502 说明 nginx /ws/ 配置或后端异常）')
    }
  }
} else {
  Write-Info 'WebSocket /ws/orders' '已跳过（加 -CheckWs 可测试）'
}
$client.Dispose()
Write-Host ('-' * 74)
if ($script:fail -eq 0) {
  Write-Host ("验收结果：通过 {0} 项，失败 0 项 —— 统一入口可用" -f $script:pass) -ForegroundColor Green
  exit 0
} else {
  Write-Host ("验收结果：通过 {0} 项，失败 {1} 项" -f $script:pass, $script:fail) -ForegroundColor Red
  exit 1
}