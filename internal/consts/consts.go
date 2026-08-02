package consts

const (
	InoutTypeIn  = 1
	InoutTypeOut = 2
)

const (
	BizTypeBank2C = 1
	BizTypeC2C    = 2
	BizTypeC2Bank = 3
	BizTypeC2B    = 4
)

const C2CBillStateOK = 1

const (
	C2BBillStateSuccess = 1
	C2BBillStateClose   = 99
)

const SaveBillStateOK = 1

const (
	MsgStateInit = 0 // 初始状态
	MsgStateSent = 1 // 已发送
	MsgStateDone = 2 // 已完成
)

const (
	PendingC2cTransferInit = 1 // 待入账
	PendingC2cTransferDone = 2 // 已入账
)

const (
	PendingC2bTransferInit = 1 // 待入账
	PendingC2bTransferDone = 2 // 已入账
)
