package logic

import (
	"context"
	"encoding/json"
	"time"

	"github.com/starslipay/account_mgr/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// UpdateBalanceMessage 余额缓存更新消息, 投递到 Kafka topic t_update_balance
// 消费端收到后回调 cache_mgr.UpdateBalance RPC 执行实际的 Redis CAS 更新
type UpdateBalanceMessage struct {
	Uid                 int64 `json:"uid"`
	CurType             int64 `json:"cur_type"`
	Balance             int64 `json:"balance"`
	BalanceModifyTimeS  int64 `json:"balance_modify_time_s"`  // 余额变更时间(秒)
	BalanceModifyTimeMs int64 `json:"balance_modify_time_ms"` // 余额变更时间的毫秒部分(0-999)
}

// sendUpdateBalanceMessage 发送余额缓存更新消息到 Kafka(t_update_balance), 由消费端异步回调 cache_mgr 更新缓存
// 替代原先直接写缓存的调用, 解耦主流程与缓存写入
// modifyTime 必须与写入数据库 t_c_account 的余额变更时间一致
// 发送失败仅记录日志, 不影响主流程(缓存最终一致性)
func sendUpdateBalanceMessage(ctx context.Context, svcCtx *svc.ServiceContext, uid int64, curType int64, balance int64, modifyTime time.Time) {
	msg := &UpdateBalanceMessage{
		Uid:                 uid,
		CurType:             curType,
		Balance:             balance,
		BalanceModifyTimeS:  modifyTime.Unix(),
		BalanceModifyTimeMs: modifyTime.UnixMilli() % 1000,
	}
	jsonStr, err := json.Marshal(msg)
	if err != nil {
		logx.WithContext(ctx).Errorf("marshal update balance message failed, uid: %d, err: %v", uid, err)
		return
	}
	if err := svcCtx.UpdateBalanceProducer.Push(ctx, string(jsonStr)); err != nil {
		logx.WithContext(ctx).Errorf("send update balance message failed, uid: %d, err: %v", uid, err)
	}
}
