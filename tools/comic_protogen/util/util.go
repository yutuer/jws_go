package util

type ProtoInfo struct {
	Name            string
	Type            string // "Req", "Push"
	Dir             string
	ClientReqParams []ClientReqParam // 客户端请求参数
	HasReqIfPush    bool             // 如果是push消息， 服务器端的player_msg是否需要req字段
}

type ClientReqParam struct {
	ParamType string
	ParamVar  string
}

const (
	MESSAGE_TYPE_REQ      = "Req"
	MESSAGE_TYPE_PUSH     = "Push"
	MESSAGE_TYPE_PUSH_REQ = "PushReq"
)

var Proto2CSTypeMap map[string]string = map[string]string{"string": "string",
	"int32":  "int",
	"bool":   "bool",
	"uint32": "uint",
	"int64":  "long",
}

func endsWith(str string, seq string) bool {
	if len(str) < len(seq) {
		return false
	}
	i := len(str) - 1
	j := len(seq) - 1
	for {
		if str[i] != seq[j] {
			return false
		}
		i--
		j--
		if j == -1 {
			return true
		}
	}
}
