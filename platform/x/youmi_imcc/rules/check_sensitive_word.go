package rules

import (
	"strings"

	"encoding/json"
	"io/ioutil"

	"fmt"

	"unicode/utf8"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/youmi_imcc/base"
)

const (
	JSONPATH        = "conf/sw_conf.json"
	REQUIRESPACENUM = 3
)

var (
	swc       SensitiveWordConfig
	muteTimes map[string]int
)

func sensitiveWordInitial() {
	err := loadConf()
	if err != nil {
		logs.Error("load sensitive word json config err by %v", err)
		panic(fmt.Sprintf("load sensitive word json config err by %v", err))
	}
	muteTimes = make(map[string]int, 0)
}

type SensitiveWordConfig struct {
	MaxExist      int      `json:"max_exist"`
	SensitiveWord []string `json:"sensitive_word"`
	MuteTime      []int    `json:"mute_time"`
}

func loadConf() error {
	byteData, err := ioutil.ReadFile(JSONPATH)
	if err != nil {
		return err
	}
	err = json.Unmarshal(byteData, &swc)
	if err != nil {
		return err
	}
	logs.Debug("load conf success: %v", swc)
	return nil
}

func saveConf() {

}

func checkSensitiveWord(newMsg base.PlayerMsgInfo) *base.MuteInfo {
	if newMsg.ChatType == worldChatType {
		newMsg.MsgContent = judgeAndDeleteSpacesIfSpaceNumSurpassRequireSpaceNum(newMsg.MsgContent)
		containTimes := 0
		for _, word := range swc.SensitiveWord {
			if strings.Contains(newMsg.MsgContent, word) {
				containTimes++
			}
		}
		if containTimes >= swc.MaxExist {
			addMuteTime(newMsg.SenderAcID)
			times := muteTimes[newMsg.SenderAcID]
			if times >= len(swc.MuteTime) {
				times = len(swc.MuteTime) - 1
			}
			return base.GenMuteInfo(newMsg.SenderAcID, "世界频道聊天信息中含有大量敏感词",
				int64(swc.MuteTime[times]), newMsg.MsgContent, true)
		}
	}
	return nil
}

func addMuteTime(acID string) {
	if v, ok := muteTimes[acID]; ok {
		muteTimes[acID] = v + 1
	} else {
		muteTimes[acID] = 1
	}
}

func judgeAndDeleteSpacesIfSpaceNumSurpassRequireSpaceNum(info string) (result string) {
	copyInfo := info
	spaceCount := 0
	//检查copyInfo中的每一个字符，每遇到一个空格则摘除并计数++
	for i := 0; i < len(copyInfo); {
		r, size := utf8.DecodeRuneInString(copyInfo[i:])
		if r == ' ' || r == '　' {
			copyInfo = deleteTargetRune(copyInfo, i, size)
			spaceCount++
		} else {
			i += size
		}
	}
	if spaceCount >= REQUIRESPACENUM {
		return copyInfo
	} else {
		return info
	}
}

func deleteTargetRune(info string, target, size int) string {
	return info[:target] + info[target+size:]
}
