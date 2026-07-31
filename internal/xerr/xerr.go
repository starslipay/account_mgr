package xerr

// 错误码  10000 0000 ~~99999 9999
// 模块id  30000
// 错误码 = 模块id + 业务错误码
var (
	ModuleId        = int64(30000)
	ModuleErrorBase = ModuleId * 10000
)

var (
	// 系统错误 0000-0999
	ErrCodeServerInternal = ModuleErrorBase + 1

	// 业务错误码 1000-1999
	ErrCodeParam                      = ModuleErrorBase + 1000
	ErrCodeBalanceNotEnough           = ModuleErrorBase + 1001
	ErrCodeDB                         = ModuleErrorBase + 1002
	ErrCodeRepeatButInfoNotConsistent = ModuleErrorBase + 1003 // 重入，但信息不一致
	ErrCodeBillNotFound               = ModuleErrorBase + 1004 // 单据不存在
	ErrCodeBillStateNotOK             = ModuleErrorBase + 1005 // 单据状态不是OK
	ErrCodeSupplyModeC2BBillNotFound  = ModuleErrorBase + 1006 // 补单模式下C2B单据不存在
	ErrCodeC2BBillStateInvalid        = ModuleErrorBase + 1007 // C2B单据状态无效
	ErrCodeC2BBillStateAlreadyClose   = ModuleErrorBase + 1008 // C2B单据状态已关闭
)
