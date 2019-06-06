package rules

import (
	"testing"

	"unicode/utf8"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/youmi_imcc/base"
)

func TestCheckSensitiveWord(t *testing.T) {
	infos := []base.PlayerMsgInfo{
		{
			MsgID:      "Test",
			SenderAcID: "1234545",
			ChatType:   2,
			MsgContent: "fsdfasdfasdfasgfdgfdgfsdafsfdssdfsd職先貨",
			CreateTime: "2017-02-23 10:33:42",
		},
		{
			MsgID:      "Test",
			SenderAcID: "1234545",
			ChatType:   2,
			MsgContent: "職先货＋fdssdfsd+鑽",
			CreateTime: "2017-02-23 10:35:42",
		},
		{
			MsgID:      "Test",
			SenderAcID: "1234545",
			ChatType:   2,
			MsgContent: "職先   货＋fdssdfsd+鑽",
			CreateTime: "2017-02-23 10:35:42",
		},
	}
	for i, v := range infos {
		if result := checkSensitiveWord(v); result != nil {
			t.Errorf("第%d条测试失败，应返回nil，实际返回:%v", i+1, result)
		}
	}
}

func TestDeleteTargetRune(t *testing.T) {
	infos := []string{
		"我爱钛核，钛核是我家",
		"工作使我快乐",
		"做最棒的游戏",
		"是我的理想",
	}
	for i, v := range infos {
		r, size := utf8.DecodeRuneInString(v[3:])
		logs.Trace("要删除:%q,大小为:%d", r, size)
		result := deleteTargetRune(v, 3, size)
		logs.Trace("检查后的字符串为：%v", result)
		if result != v[:3]+v[3+size:] {
			t.Errorf("第%d条测试失败，应返回%v，实际返回:%v", i+1, v[:3]+v[3+size:], result)
		}
	}
}
func TestJudgeAndDeleteSpacesIfSpaceNumSurpassRequireSpaceNum(t *testing.T) {
	infos := []string{
		"我爱  钛核，钛 核是　我　家",
		"工作 使我  快乐",
		"做最  棒 的游戏",
		"是我的   理想",
	}
	infosTarget := []string{
		"我爱钛核，钛核是我家",
		"工作使我快乐",
		"做最棒的游戏",
		"是我的理想",
	}
	for i, v := range infos {
		result := judgeAndDeleteSpacesIfSpaceNumSurpassRequireSpaceNum(v)
		if infosTarget[i] != result {
			t.Errorf("第%d条测试失败，应返回%v，实际返回:%v", i+1, infosTarget[i], result)
		}
	}
}
