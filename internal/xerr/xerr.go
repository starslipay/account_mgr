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
	ErrCodeDB             = ModuleErrorBase + 0
	ErrCodeServerInternal = ModuleErrorBase + 1

	// 业务错误码 1000-1999
	ErrCodeParam            = ModuleErrorBase + 1000
	ErrCodeBalanceNotEnough = ModuleErrorBase + 1001
)
