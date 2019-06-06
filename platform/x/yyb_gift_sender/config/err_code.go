package config

const (
	RetCode_FinishAndSuccess = 0 //步骤已完成或奖励发放成功
	RetCode_InvalidRequest   = 1 // 请求无效
	RetCode_ErrorArg         = 2 // 请求参数错误
	RetCode_ErrorSig         = 3 // sig参数错误

	RetCode_NoRole      = 101 // 用户尚未在应用创建角色
	RetCode_HadSend     = 102 // 该步骤奖励已发放过
	RetCode_SendFailed  = 103 // 奖励发放失败
	RetCode_NoCondition = 104 //玩家不符合领取条件
)

var ErrMap map[int]string

func InitErrCode() {
	ErrMap = make(map[int]string, 8)
	ErrMap[RetCode_FinishAndSuccess] = "步骤已完成或奖励发放成功"
	ErrMap[RetCode_InvalidRequest] = "请求无效"
	ErrMap[RetCode_ErrorArg] = "请求参数错误"
	ErrMap[RetCode_ErrorSig] = "sig参数错误"
	ErrMap[RetCode_NoRole] = "用户尚未在应用创建角色"
	ErrMap[RetCode_HadSend] = "该步骤奖励已发放过"
	ErrMap[RetCode_SendFailed] = "奖励发放失败"
	ErrMap[RetCode_NoCondition] = "玩家不符合领取条件"
}
