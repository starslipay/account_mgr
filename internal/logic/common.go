package logic

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/starslipay/account_mgr/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// RedisKeyBalancePrefix 余额缓存 key 前缀: balance|{uid} -> balance=xx&balance_modify_time_s=xx&balance_modify_time_ms=xx&cur_type=xx
const RedisKeyBalancePrefix = "balance|"

// setBalanceScript 乐观锁 CAS 写入余额缓存, 版本校验与写入在 Redis 内原子完成
// (go-zero 的 redis.Redis 未暴露 WATCH, Lua 脚本是 Redis 官方的原子 CAS 实现,
// 相比 WATCH/MULTI/EXEC 没有并发冲突重试窗口, 且只需一次往返)
//
// 版本号 = 余额变更时间: 秒为 UNIX 秒, 毫秒为秒内的毫秒部分(0-999); 先比秒, 同一秒再比毫秒部分, 越大越新
// KEYS[1]: balance|{uid}
// ARGV[1]: 期望版本-秒(获取时读到的); ARGV[2]: 期望版本-毫秒部分
// ARGV[3]: 新的 value 串; ARGV[4]: 新版本-秒; ARGV[5]: 新版本-毫秒部分
// 返回: 1-写入成功; 0-已有更新或相同版本数据, 跳过; -1-版本号与获取时不一致(并发冲突); -2-key 被并发删除
var setBalanceScript = redis.NewScript(`
-- key 必须通过 KEYS 数组传入(而不是拼在 ARGV 里), 这是 Redis 集群/分片模式下脚本能正确定位到对应节点的前提
local key = KEYS[1]
-- 期望版本号-秒: 调用方"先获取"时从缓存读到的 balance_modify_time_s, 它是乐观锁的比对基准; tonumber 把字符串参数转成数字, 后面才能做数值比较
local expectS = tonumber(ARGV[1])
-- 期望版本号-毫秒部分(0-999): 与秒组成完整版本号, 用于"秒相同"时比更细的新旧, 同样来自获取阶段读到的值
local expectMs = tonumber(ARGV[2])
-- 本次要写入的完整 value 串(balance=xx&balance_modify_time_s=xx&balance_modify_time_ms=xx), 由 Go 侧 url.Values.Encode() 拼好传进来
local newVal = ARGV[3]
-- 新版本号-秒: 本次余额变更的时间戳, 与写入 MySQL t_c_account 的时间字段是同一份时间, 保证 DB 与缓存对齐
local newS = tonumber(ARGV[4])
-- 新版本号-毫秒部分(0-999): 同上, 作为同一秒内的更细粒度版本号
local newMs = tonumber(ARGV[5])

-- 读取当前值。整个"读取-校验-写入"都在这一次脚本执行里完成: Redis 单线程执行 Lua 期间不会插入其他命令,
-- 所以这里读到的值一直到下面 SET 之前都不会变, 这就是不用 WATCH/MULTI 也能实现原子 CAS 的原因
local cur = redis.call('GET', key)
-- key 不存在, 需要区分是"首次写入"还是"获取之后被并发删掉/过期", 两种处理方式不同
if not cur then
    -- 期望版本为 0(且毫秒也为 0), 说明调用方获取时 key 本来就不存在, 属于首次写入场景, 允许创建
    if expectS == 0 and expectMs == 0 then
        -- 直接创建该 key
        redis.call('SET', key, newVal)
        -- 返回 1 表示写入成功
        return 1
    end
    -- 期望版本非 0 但 key 已经没了: 说明"获取"之后 key 被其他请求删除或自然过期, 属于并发变更, 返回 -2 让调用方感知冲突并报错
    return -2
end

-- 从当前 value 串解析出"秒"版本: string.match 用 %d+ 抓取 balance_modify_time_s= 后面的连续数字; tonumber 转数字;
-- or 0 是兜底——字段缺失或格式异常时按 0 处理(等价于"版本未知"), 避免脚本直接报错
local curS = tonumber(string.match(cur, 'balance_modify_time_s=(%d+)')) or 0
-- 同理解析出"毫秒"版本, 与 curS 一起构成当前数据的版本号
local curMs = tonumber(string.match(cur, 'balance_modify_time_ms=(%d+)')) or 0

-- 乐观锁核心校验: 当前版本号 != 获取时的期望版本号, 说明这条数据在"获取"与"set"之间已经被其他并发请求改过了。
-- 此时直接放弃写入并返回 -1, 即需求中的"set 时版本号与获取时不一致直接报错", 由调用方记录冲突日志
if curS ~= expectS or curMs ~= expectMs then
    return -1
end

-- 新旧判断(防止乱序请求用旧数据覆盖新数据, 例如异步入账消息晚于后来的一次余额变更到达):
-- 缓存中已有的版本比本次新, 或版本完全相同(同一次变更被重复投递), 都无需覆盖, 返回 0 跳过写入
-- 比较规则就是需求里的"先比秒, 秒相同再比毫秒": curS == newS and curMs >= newMs
if curS > newS or (curS == newS and curMs >= newMs) then
    return 0
end

-- 版本校验通过且数据确实更新, 原子写入新值(Lua 执行期间无其他命令插入, 这里不会再被并发篡改)
redis.call('SET', key, newVal)
-- 返回 1 表示 CAS 写入成功
return 1
`)

// updateBalanceCache 更新 redis 中 balance|{uid} 的余额信息, 乐观锁 CAS 设计:
//  1. 先获取, 解析版本号(余额变更时间秒/毫秒)并判断新旧, 缓存不旧于新数据则跳过
//  2. 构造新值, CAS set; 若 set 时版本号与获取时不一致(期间被并发修改), 直接报错
//
// modifyTime 必须与写入数据库 t_c_account 的余额变更时间一致
// 缓存属尽力而为: 失败仅记录日志, 不影响余额变更主流程
func updateBalanceCache(ctx context.Context, svcCtx *svc.ServiceContext, uid int64, curType int64, balance int64, modifyTime time.Time) {
	key := RedisKeyBalancePrefix + strconv.FormatInt(uid, 10)
	// 新版本号: 秒为 UNIX 秒; 毫秒只取秒内的毫秒部分(0-999), 而不是按毫秒换算的完整时间戳
	newS, newMs := modifyTime.Unix(), modifyTime.UnixMilli()%1000

	// 1. 先获取, 解析当前版本号; key 不存在视为版本 0(首次写入)
	var expectS, expectMs int64
	cached, err := svcCtx.RedisClient.GetCtx(ctx, key)
	if err != nil && !errors.Is(err, redis.Nil) {
		logx.WithContext(ctx).Errorf("get balance cache failed, uid: %d, err: %v", uid, err)
		return
	}
	if err == nil {
		expectS, expectMs = parseBalanceVersion(cached)
		// 新旧判断: 先比秒, 同一秒再比毫秒; 缓存不旧于新数据则无需更新
		if expectS > newS || (expectS == newS && expectMs >= newMs) {
			logx.WithContext(ctx).Infof("balance cache has newer version, skip, uid: %d, balance: %d, version: %d/%d, new: %d/%d",
				uid, balance, expectS, expectMs, newS, newMs)
			return
		}
	}

	// 2. 构造新值
	newVal := url.Values{}
	newVal.Set("balance", strconv.FormatInt(balance, 10))
	newVal.Set("balance_modify_time_s", strconv.FormatInt(newS, 10))
	newVal.Set("balance_modify_time_ms", strconv.FormatInt(newMs, 10))
	// 货币类型, 与 t_c_account.cur_type 同源
	newVal.Set("cur_type", strconv.FormatInt(curType, 10))

	// 3. CAS set: 版本号与获取时不一致则报错
	resp, err := svcCtx.RedisClient.ScriptRunCtx(ctx, setBalanceScript, []string{key},
		strconv.FormatInt(expectS, 10), strconv.FormatInt(expectMs, 10),
		newVal.Encode(),
		strconv.FormatInt(newS, 10), strconv.FormatInt(newMs, 10))
	if err != nil {
		logx.WithContext(ctx).Errorf("cas set balance cache failed, uid: %d, balance: %d, err: %v", uid, balance, err)
		return
	}
	if n, ok := resp.(int64); ok {
		switch n {
		case -1, -2:
			// 乐观锁冲突: 获取后版本被并发修改(或 key 被并发删除), 直接报错
			logx.WithContext(ctx).Errorf("balance cache cas conflict, version changed after get, uid: %d, balance: %d, expect version: %d/%d",
				uid, balance, expectS, expectMs)
		case 0:
			logx.WithContext(ctx).Infof("balance cache has newer version, skip, uid: %d, balance: %d", uid, balance)
		}
	}
}

// parseBalanceVersion 从 url 键值对串中解析版本号(balance_modify_time_s/balance_modify_time_ms), 解析失败视为 0
func parseBalanceVersion(val string) (s, ms int64) {
	v, err := url.ParseQuery(val)
	if err != nil {
		return 0, 0
	}
	s, _ = strconv.ParseInt(v.Get("balance_modify_time_s"), 10, 64)
	ms, _ = strconv.ParseInt(v.Get("balance_modify_time_ms"), 10, 64)
	return s, ms
}
