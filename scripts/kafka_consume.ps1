# kafka-consume.ps1
# 查看 Kafka topic 当前消息的 Windows 脚本
# 通过 docker exec 进入 kafka 容器, 调用 kafka-console-consumer.sh
#
# 用法:
#   .\kafka-consume.ps1                              # 交互模式
#   .\kafka-consume.ps1 t_update_balance             # 最新模式(默认)
#   .\kafka-consume.ps1 t_update_balance -Mode all   # 全部模式(从头消费)
#   .\kafka-consume.ps1 -Topic t_update_balance -Mode latest -Container kafka1
#
# 注意: 本脚本在 kafka 容器内部执行命令, 必须使用集群内部监听地址 kafka1:29092
#       (EXTERNAL 的 127.0.0.1:9092 只在宿主机可用, 容器内使用会拿到错误的 broker 元数据)
#
# 模式说明:
#   latest - 只消费连上之后新到的消息(Ctrl+C 停止)
#   all    - 从头消费所有已持久化的消息, 打完后回主菜单
#
# 参数说明:
#   -TimeoutSec  all 模式下等待新消息的秒数(默认 5 秒)

param(
    [Parameter(Position=0)]
    [string]$Topic = "",
    [ValidateSet("latest", "all")]
    [string]$Mode = "latest",
    [string]$Container = "kafka1",
    [string]$Bootstrap = "kafka1:29092",
    [int]$TimeoutSec = 5
)

$ErrorActionPreference = "Stop"

function Test-KafkaContainer {
    $null = docker ps --filter "name=^/$Container$" --format "{{.Names}}" 2>$null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "错误: docker 不可用或容器 $Container 不存在" -ForegroundColor Red
        Write-Host "可用容器: docker ps" -ForegroundColor Yellow
        exit 1
    }
}

function Invoke-ConsoleConsumer {
    param(
        [string]$TopicName,
        [string]$ConsumeMode
    )

    $bootstrapServer = $Bootstrap
    $argsList = @(
        "exec", $Container, "bash", "-c"
    )

    $cmd = "/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server $bootstrapServer --topic $TopicName"

    if ($ConsumeMode -eq "all") {
        $cmd += " --from-beginning"
    }

    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "Topic:    $TopicName" -ForegroundColor Green
    Write-Host "Mode:     $ConsumeMode" -ForegroundColor Green
    Write-Host "Container:$Container" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Cyan

    if ($ConsumeMode -eq "all") {
        Write-Host "正在从头消费所有消息, 请稍候..." -ForegroundColor Yellow
        # all 模式: 用 timeout 命令在指定秒数后自动退出, 避免一直挂起
        # timeout 3000ms 不是 3 秒, 是 30 分钟; 正确语法是 5s / 10s
        $cmdWithTimeout = "timeout ${TimeoutSec}s $cmd 2>/dev/null || true"
        docker @argsList $cmdWithTimeout
        Write-Host "`n消费完成, 共等待 ${TimeoutSec}s 无新消息" -ForegroundColor Green
    } else {
        Write-Host "正在监听新消息, 按 Ctrl+C 停止..." -ForegroundColor Yellow
        docker @argsList $cmd
    }
}

function Show-MainMenu {
    Write-Host "`n--- Kafka 消息消费工具 ---" -ForegroundColor Cyan
    Write-Host "1. 查看所有消息 (from-beginning)"
    Write-Host "2. 查看最新消消息 (新到的)"
    Write-Host "3. 列出所有 topic"
    Write-Host "4. 修改容器名 (当前: $Container)"
    Write-Host "5. 退出"
}

# 命令行带参数直接运行
if ($Topic -ne "") {
    Test-KafkaContainer
    Invoke-ConsoleConsumer -TopicName $Topic -ConsumeMode $Mode
    exit 0
}

# 交互模式
Test-KafkaContainer
while ($true) {
    Show-MainMenu
    $choice = Read-Host "请选择"

    switch ($choice) {
        "1" {
            $t = Read-Host "输入 topic 名称"
            if ($t -ne "") { Invoke-ConsoleConsumer -TopicName $t -ConsumeMode "all" }
        }
        "2" {
            $t = Read-Host "输入 topic 名称"
            if ($t -ne "") { Invoke-ConsoleConsumer -TopicName $t -ConsumeMode "latest" }
        }
        "3" {
            Write-Host "`n--- 所有 topic ---" -ForegroundColor Cyan
            docker exec $Container bash -c "/opt/kafka/bin/kafka-topics.sh --bootstrap-server $Bootstrap --list"
        }
        "4" {
            $newContainer = Read-Host "输入新的容器名"
            if ($newContainer -ne "") { $Container = $newContainer }
        }
        "5" { exit 0 }
        default { Write-Host "无效选择" -ForegroundColor Red }
    }
}
