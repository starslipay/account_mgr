# ============================================================
# Redis 快捷查询脚本 (PowerShell)
# 容器: redis-master   地址: redis-master:6379   无密码
# 用法: .\redis-query.ps1            交互菜单
#       .\redis-query.ps1 "balance|*"  带初始 pattern
# ============================================================

$Container = 'redis-master'
$script:Pattern = if ($args.Count -gt 0) { $args[0] } else { '*' }

function Show-Menu {
    Clear-Host
    Write-Host "============================================" -ForegroundColor Cyan
    Write-Host "  Redis 查询工具   当前匹配: $script:Pattern"
    Write-Host "============================================" -ForegroundColor Cyan
    Write-Host "  1. 查看 key 总数 (dbsize)"
    Write-Host "  2. 按 pattern 列出 key (scan 分页, 生产安全)"
    Write-Host "  3. 按 pattern 列出 key (keys, 全量, 仅调试用)"
    Write-Host "  4. 查看某个 key 的类型和值"
    Write-Host "  5. 修改匹配 pattern"
    Write-Host "  6. 进入 redis-cli 交互模式 (自由敲命令)"
    Write-Host "  0. 退出"
    Write-Host "============================================" -ForegroundColor Cyan
}

# 在容器内执行 redis-cli 命令
function Invoke-Redis {
    param([string[]]$RedisArgs)
    docker exec $Container redis-cli @RedisArgs
}

while ($true) {
    Show-Menu
    $choice = Read-Host "请选择"

    switch ($choice) {
        '1' {
            # dbsize 返回当前库 key 总数
            Invoke-Redis @('dbsize')
        }
        '2' {
            # SCAN 游标分页遍历, 不阻塞 Redis, 生产环境安全
            $cnt = Read-Host "每页数量(直接回车默认100)"
            if ([string]::IsNullOrWhiteSpace($cnt)) { $cnt = '100' }
            Write-Host "`n--- scan 匹配 `"$script:Pattern`" ---" -ForegroundColor Yellow
            # 返回值第一行是游标, 0 表示遍历完一轮; count 只是每轮扫描的桶数提示值
            $result = Invoke-Redis @('scan', '0', 'match', $script:Pattern, 'count', $cnt)
            $cursor = $result[0]
            $result[1..($result.Count - 1)] | ForEach-Object { Write-Host $_ }
            Write-Host "`n游标: $cursor  (非 0 时可继续翻页)" -ForegroundColor Yellow
            if ($cursor -ne '0') {
                $more = Read-Host "输入 y 用该游标看下一页, 其他键返回菜单"
                while ($more -eq 'y' -and $cursor -ne '0') {
                    $result = Invoke-Redis @('scan', $cursor, 'match', $script:Pattern, 'count', $cnt)
                    $cursor = $result[0]
                    $result[1..($result.Count - 1)] | ForEach-Object { Write-Host $_ }
                    Write-Host "`n游标: $cursor" -ForegroundColor Yellow
                    if ($cursor -ne '0') { $more = Read-Host "输入 y 继续下一页, 其他键返回菜单" } else { break }
                }
            }
        }
        '3' {
            # KEYS 一次性全量匹配, key 量大时会阻塞 Redis, 仅调试用
            Write-Host "`n--- keys 匹配 `"$script:Pattern`" ---" -ForegroundColor Yellow
            Invoke-Redis @('keys', $script:Pattern)
        }
        '4' {
            $k = Read-Host "请输入 key"
            if (-not [string]::IsNullOrWhiteSpace($k)) {
                Write-Host "`n--- 类型 ---" -ForegroundColor Yellow
                $type = (Invoke-Redis @('type', $k)) | Out-String
                Write-Host $type.Trim()
                # 根据类型选对应的读取命令: string->get, hash->hgetall, list->lrange, set->smembers, zset->zrange
                switch ($type.Trim()) {
                    'string' { Invoke-Redis @('get', $k) }
                    'hash'   { Invoke-Redis @('hgetall', $k) }
                    'list'   { Invoke-Redis @('lrange', $k, '0', '-1') }
                    'set'    { Invoke-Redis @('smembers', $k) }
                    'zset'   { Invoke-Redis @('zrange', $k, '0', '-1', 'withscores') }
                }
                # TTL: -1 永不过期, -2 key 不存在
                Write-Host "`n--- TTL ---" -ForegroundColor Yellow
                Invoke-Redis @('ttl', $k)
            }
        }
        '5' {
            $script:Pattern = Read-Host "请输入 pattern (如 userinfo|* 或 balance|123)"
        }
        '6' {
            Write-Host "`n进入交互模式, 常用命令:" -ForegroundColor Yellow
            Write-Host "  dbsize                 key 总数"
            Write-Host "  keys balance|*         列出所有余额缓存 key"
            Write-Host '  get "balance|123"      看某个 key 的值'
            Write-Host '  ttl "balance|123"      看过期时间(-1永不过期)'
            Write-Host "  exit                   退出交互模式"
            # -it 分配终端, 退出后回到菜单
            docker exec -it $Container redis-cli
        }
        '0' { break }
        default { }
    }
    if ($choice -eq '0') { break }
    if ($choice -ne '6') {
        Write-Host ""
        Read-Host "回车返回菜单" | Out-Null
    }
}
