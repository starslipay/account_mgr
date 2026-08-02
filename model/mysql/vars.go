package mysql

import (
	"database/sql"
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var ErrNotFound = sqlx.ErrNotFound

// ErrInvalidRowsAffected 表示DB修改操作影响的行数不是预期的1行
var ErrInvalidRowsAffected = errors.New("invalid rows affected, expect 1 row")

// checkOneRowAffected 校验DB修改操作影响的行数必须正好为1行，否则返回错误
func checkOneRowAffected(ret sql.Result) error {
	rows, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrInvalidRowsAffected
	}
	return nil
}
