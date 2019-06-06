package alibaba_uc

const (
	RetCode_Success         = 2000000 //成功
	RetCode_InnerError      = 5000000 // 服务器内部错误
	RetCode_CallerError     = 5000010 // Caller不正确
	RetCode_MD5SignError    = 5000011 //  MD5计算的签名不正确
	RetCode_AESError        = 5000012 // 数据节点AES加/解密失败
	RetCode_ArgError        = 5000020 // 业务参数无效
	RetCode_GameIDError     = 5000030 // gameID错误
	RetCode_ACIDError       = 5000031 // accountID错误
	RetCode_GiftIDError     = 5000032 // 礼包编号错误
	RetCode_GiftNumError    = 5000033 // 礼包数量错误
	RetCode_RoleInfoError   = 5000034 // 角色信息错误
	RetCode_ServerInfoError = 5000035 // 区服信息错误
	RetCode_AlreadyGot      = 5000036 // 已经领取
)

var errMap map[int]string

func initErrorCode() {
	errMap = make(map[int]string, 6)
	errMap[RetCode_Success] = "成功"
	errMap[RetCode_InnerError] = "服务器内部错误"
	errMap[RetCode_CallerError] = "Caller不正确"
	errMap[RetCode_MD5SignError] = "MD5计算的签名不正确"
	errMap[RetCode_AESError] = "数据节点AES加/解密失败"
	errMap[RetCode_ArgError] = "业务参数无效"
	errMap[RetCode_GameIDError] = "gameID错误"
	errMap[RetCode_ACIDError] = "accountID错误"
	errMap[RetCode_GiftIDError] = "礼包编号错误"
	errMap[RetCode_GiftNumError] = "礼包数量错误"
	errMap[RetCode_RoleInfoError] = "角色信息错误"
	errMap[RetCode_ServerInfoError] = "区服信息错误"
	errMap[RetCode_AlreadyGot] = "已经领取"
}

func GetErrorCode() map[int]string {
	return errMap
}
