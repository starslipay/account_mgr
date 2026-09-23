package xerr

// 错误码  10000 0000 ~~99999 9999
// 模块id  30000
// 错误码 = 模块id + 业务错误码
var (
	ModuleId        = int64(455905)
	ModuleErrorBase = ModuleId * 100
)

var (
	// 系统错误 0000-0999
	ErrCodeServerInternal = ModuleErrorBase + 1

	// 业务错误码 1000-1999
	ErrCodeParam                      = ModuleErrorBase + 100
	ErrCodeBalanceNotEnough           = ModuleErrorBase + 101
	ErrCodeDB                         = ModuleErrorBase + 102
	ErrCodeRepeatButInfoNotConsistent = ModuleErrorBase + 103 // 重入,但信息不一致
	ErrCodeC2CBillNotFound            = ModuleErrorBase + 104 // C2C单据不存在
	ErrCodeBillStateNotOK             = ModuleErrorBase + 105 // 单据状态不是OK
	ErrCodeSupplyModeC2BBillNotFound  = ModuleErrorBase + 106 // 补单模式下C2B单据不存在
	ErrCodeC2BBillStateInvalid        = ModuleErrorBase + 107 // C2B单据状态无效
	ErrCodeC2BBillStateAlreadyClose   = ModuleErrorBase + 108 // C2B单据状态已关闭
	ErrCodeC2BBillNotFound            = ModuleErrorBase + 109 // C2B单据不存在
	ErrCodeC2BBillConflict            = ModuleErrorBase + 110 // C2B单据已存在,插入冲突
)
