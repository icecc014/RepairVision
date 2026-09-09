param(
  [string]$BaseUrl = "http://localhost:8080",
  [int]$Count = 60,
  [string]$Password = "admin123"
)

$ErrorActionPreference = "Stop"

# 演示数据：dorm1-dorm4 循环报修，房间号唯一化（每栋最多 5*16=80 个可用房间区间）。
$dorms = @("dorm1", "dorm2", "dorm3", "dorm4")
$faults = @("electric", "water", "other")

function Login($username) {
  $body = @{ username = $username; password = $Password } | ConvertTo-Json
  $resp = Invoke-RestMethod -Uri "$BaseUrl/api/login" -Method Post -ContentType "application/json" -Body $body
  return $resp.data
}

$created = 0
for ($i = 0; $i -lt $Count; $i++) {
  $dorm = $dorms[$i % $dorms.Length]
  $login = Login $dorm
  $headers = @{ Authorization = "Bearer " + $login.token }

  # 以 i 推导楼层与房间序号，避免同房间同类型 2 小时内重复被拦截。
  $seq = ($i % 16) + 1
  $floor = [Math]::Floor($i / 16) % 5 + 1
  $room = "$floor" + $seq.ToString("00")
  $fault = $faults[$i % $faults.Length]

  $body = @{
    room = $room
    floor = [int]$floor
    faultType = $fault
    description = "演示数据自动生成 #$($i + 1)"
  } | ConvertTo-Json

  try {
    $resp = Invoke-RestMethod -Uri "$BaseUrl/api/dorm/orders" -Method Post `
      -ContentType "application/json" -Headers $headers -Body $body
    $created++
    Write-Host ("[$($created)/$Count] " + $dorm + " -> " + $room + " " + $fault + " order=" + $resp.data.orderId)
  } catch {
    Write-Warning ("第 $($i + 1) 单失败: " + $_.Exception.Message)
  }
}

Write-Host ""
Write-Host ("演示数据生成完成，成功 $created 单（访问 PC 统计看板查看趋势）。")