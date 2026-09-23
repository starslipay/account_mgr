package xerr

// 错误码  10000 0000 ~~99999 9999
// 模块id  30000
// 错误码 = 模块id + 业务错误码
var (
	ModuleId = int64(455905)
)

var (
	// 系统错误 0000-0999
	ErrCodeServerInternal = int64(455905001)

	// 业务错误码 1000-1999
	ErrCodeParam                      = int64(455905100)
	ErrCodeBalanceNotEnough           = int64(455905101)
	ErrCodeDB                         = int64(455905102)
	ErrCodeRepeatButInfoNotConsistent = int64(455905103) // 重入,但信息不一致
	ErrCodeC2CBillNotFound            = int64(455905104) // C2C单据不存在
	ErrCodeBillStateNotOK             = int64(455905105) // 单据状态不是OK
	ErrCodeSupplyModeC2BBillNotFound  = int64(455905106) // 补单模式下C2B单据不存在
	ErrCodeC2BBillStateInvalid        = int64(455905107) // C2B单据状态无效
	ErrCodeC2BBillStateAlreadyClose   = int64(455905108) // C2B单据状态已关闭
	ErrCodeC2BBillNotFound            = int64(455905109) // C2B单据不存在
	ErrCodeC2BBillConflict            = int64(455905110) // C2B单据已存在,插入冲突
)
