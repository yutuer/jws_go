package tools

import (
	"strconv"
	"strings"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type StayForJudge struct {
	RowNum int
	Gid    string
	IAPID  string
	IDS
}

var GIDS []string = []string{"207"}
var PlayersPackageData []StayForJudge
var GainnerOrder []int

func GetStayForJudges(xlsx_data [][]string, index_pay_description int) {
	for i := 1; i < len(xlsx_data); i++ {
		eachone := strings.Split(xlsx_data[i][index_pay_description], "|")
		othersData := strings.Split(eachone[3], `:`)
		IAPID := othersData[len(othersData)-1]
		Gid := othersData[0]
		ID, err := strconv.Atoi(eachone[4])
		if err != nil {
			logs.Error("[cyt]ID data error in number %v rows in pay_description :%v", i, err)
			continue
		}
		SubID, err := strconv.Atoi(eachone[5])
		if err != nil {
			logs.Error("[cyt]SubID data error in number %v rows in pay_description :%v", i, err)
			continue
		}
		PlayersPackageData = append(PlayersPackageData, StayForJudge{RowNum: i, Gid: Gid, IAPID: IAPID, IDS: IDS{uint32(ID), uint32(SubID)}})
	}
}

func AffirmAndRecord() {
	for _, v := range PlayersPackageData {
		//发现不是要判断的大区时不比对
		flag := false
		for i := range GIDS {
			if v.Gid == GIDS[i] {
				flag = true
			}
		}
		if !flag {
			continue
		}
		//只有当存在礼包信息时才进行比对
		if packageData, ok := Hot_package[v.IDS]; ok {
			IAPIDS := strings.Split(packageData.GetIAPID(), `,`)
			flag = false
			for i, _ := range IAPIDS {
				if v.IAPID == IAPIDS[i] {
					flag = true
					break
				}
			}
			if !flag {
				//认证玩家出现问题。记录其信息
				GainnerOrder = append(GainnerOrder, v.RowNum)
			}
		}
	}
}
