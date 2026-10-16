package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TCAccountModel = (*customTCAccountModel)(nil)

type (
	// TCAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTCAccountModel.
	TCAccountModel interface {
		tCAccountModel
		withSession(session sqlx.Session) TCAccountModel
		AddBalance(ctx context.Context, uid int64, amount int64, modifyTime time.Time) error
		SubBalance(ctx context.Context, uid int64, amount int64, modifyTime time.Time) error
		FindOneForUpdate(ctx context.Context, uid int64) (*TCAccount, error)
	}

	customTCAccountModel struct {
		*defaultTCAccountModel
	}
)

// NewTCAccountModel returns a model for the database table.
func NewTCAccountModel(conn sqlx.SqlConn) TCAccountModel {
	return &customTCAccountModel{
		defaultTCAccountModel: newTCAccountModel(conn),
	}
}

func (m *customTCAccountModel) withSession(session sqlx.Session) TCAccountModel {
	return NewTCAccountModel(sqlx.NewSqlConnFromSession(session))
}

// AddBalance 加余额, 同时写入余额变更时间: 秒为 UNIX 秒, 毫秒只存秒内的毫秒部分(0-999)
// modifyTime 由调用方生成, 保证 DB 与缓存版本号使用同一时间
func (m *customTCAccountModel) AddBalance(ctx context.Context, uid int64, amount int64, modifyTime time.Time) error {
	query := fmt.Sprintf("update %s set `balance` = `balance` + ?, `balance_modify_time_s` = ?, `balance_modify_time_ms` = ? where `uid` = ?", m.table)
	ret, err := m.conn.ExecCtx(ctx, query, amount, modifyTime.Unix(), modifyTime.UnixMilli()%1000, uid)
	if err != nil {
		return err
	}
	return checkOneRowAffected(ret)
}

// SubBalance 扣余额, 同时写入余额变更时间: 秒为 UNIX 秒, 毫秒只存秒内的毫秒部分(0-999)
// modifyTime 由调用方生成, 保证 DB 与缓存版本号使用同一时间
func (m *customTCAccountModel) SubBalance(ctx context.Context, uid int64, amount int64, modifyTime time.Time) error {
	query := fmt.Sprintf("update %s set `balance` = `balance` - ?, `balance_modify_time_s` = ?, `balance_modify_time_ms` = ? where `uid` = ? and `balance` >= ?", m.table)
	ret, err := m.conn.ExecCtx(ctx, query, amount, modifyTime.Unix(), modifyTime.UnixMilli()%1000, uid, amount)
	if err != nil {
		return err
	}
	return checkOneRowAffected(ret)
}

func (m *customTCAccountModel) FindOneForUpdate(ctx context.Context, uid int64) (*TCAccount, error) {
	query := fmt.Sprintf("select %s from %s where `uid` = ? for update", tCAccountRows, m.table)
	var resp TCAccount
	err := m.conn.QueryRowCtx(ctx, &resp, query, uid)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
