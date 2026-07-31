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
)
