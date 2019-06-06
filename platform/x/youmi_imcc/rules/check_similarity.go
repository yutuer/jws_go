package rules

import (
	"fmt"

	"math"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/youmi_imcc/base"
)

var whiteList map[string]struct{}

func similarityInitial() {
	whiteList = make(map[string]struct{}, 1000)
}
func checkSimilarity(infos *base.MsgInfo) (ret []*base.MuteInfo) {
	ret = make([]*base.MuteInfo, 0)
	tmp := make([]string, 0)
	for k, v := range infos.MsgInfoMap {
		msgValue := base.ParseMsgValue(v.Value)
		if _, ok := whiteList[k]; ok || len(msgValue) < 5 {
			continue
		}
		for _, item := range msgValue {
			tmp = append(tmp, item.MsgContent)
		}
		if isSimilarity(tmp) {
			ret = append(ret, base.GenMuteInfo(k, "世界聊天内容相似度太高",
				base.Cfg.SimilarityGagTime, fmt.Sprintf("%v", tmp), false))
			// for test
			logs.Info("similarity mute user: %v, content: %v", k, tmp)
		} else {
			whiteList[k] = struct{}{}
		}
	}
	return ret
}

// 前面的每一条都与最后一条相似度达{base.Cfg.SimilarityVal}以上，即认为相似
func isSimilarity(s []string) bool {
	if len(s) < 1 {
		return false
	}
	lastIndex := len(s) - 1
	for i := 0; i < len(s)-1; i++ {
		if getSimilarity(s[i], s[lastIndex]) < base.Cfg.SimilarityVal {
			return false
		}
	}
	return true
}

func getSimilarity(s1, s2 string) float64 {
	return levenshteinDistance(s1, s2)
}

func levenshteinDistance(s1, s2 string) float64 {
	row := len(s1)
	column := len(s2)
	rows := make([][]float64, row+1)
	for i := 0; i <= row; i++ {
		rows[i] = make([]float64, column+1)
	}
	for i := 1; i <= row; i++ {
		rows[i][0] = float64(i)
	}
	for j := 1; j <= column; j++ {
		rows[0][j] = float64(j)
	}
	for j := 1; j <= column; j++ {
		for i := 1; i <= row; i++ {
			var substitutionCost float64
			if s1[i-1] != s2[j-1] {
				substitutionCost = 1
			}
			rows[i][j] = math.Min(rows[i-1][j]+1, math.Min(rows[i][j-1]+1, rows[i-1][j-1]+substitutionCost))
		}
	}
	return 1 - rows[row][column]/math.Max(float64(row), float64(column))
}
