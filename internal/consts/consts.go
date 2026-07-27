package consts

const InoutTypeIn = 1
const InoutTypeOut = 2

const BizTypeBank2C = 1
const BizTypeC2C = 2
const BizTypeC2Bank = 3

const C2CBillStateOK = 1
const SaveBillStateOK = 1

const MsgTypeC2CTransfer = 1

const (
	MsgStateInit = 0 // 初始状态
	MsgStateSent = 1 // 已发送
	MsgStateDone = 2 // 已完成
)

const (
	PendingC2cTransferInit = 1 // 待入账
	PendingC2cTransferDone = 2 // 已入账
)
