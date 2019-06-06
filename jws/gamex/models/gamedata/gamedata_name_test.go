package gamedata

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"os"
	"path/filepath"

	"time"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func MoveToRoot() {
	GOPATH := os.Getenv("GOPATH")
	fmt.Println(GOPATH)
	testPath := "/src/vcs.taiyouxi.net/jws/gamex"
	workPath, _ := filepath.Abs(GOPATH + testPath)
	os.Chdir(workPath)
}

func TestGetTBTeamTypeID(t *testing.T) {
	MoveToRoot()
	loadTBBossDungeon("conf/data/tbossdungeon.data")
	loadTBBossHeroType("conf/data/tbossherotype.data")
	day := uint32(time.Now().Weekday())
	fmt.Println(day)
	date := GetTBTeamTypeID()
	fmt.Println(GetTBBossID(date))
	fmt.Println(GetTBSceneID(date))
}

func TestRandNames(t *testing.T) {
	MoveToRoot()
	logs.Close()
	//LoadNameEN("conf/data/nameen.data")
	//LoadNameVN("conf/data/namevi.data")
	LoadNameKO("conf/data/nameko.data")
	ret := randRobotNamesByLimit(KoRand, 100, false)

	fmt.Println(ret)
}

func TestSpace(t *testing.T) {
	//LoadGameData("")

	const (
		//允许存在多个不连续空格
		sym_reg = "[`~!@#\\$%\\^&\\*\\(\\)_\\-\\+=\\|\\\\{}\\[\\]\\:;\"'/\\?,<>？，。·！￥……（）+｛｝【】、|《》]|\\s{2,}"
	)
	var err error
	var gdSymReg *regexp.Regexp
	gdSymReg, err = regexp.Compile(sym_reg)
	if err != nil {
		//logs.Error("reg load error", err)
		return
	}

	teststr := "aaa   aaa"

	if gdSymReg.Find([]byte(strings.ToLower(teststr))) != nil {
		fmt.Println(false)
	} else {
		fmt.Println(true)
	}

	spacename := "  aaaa aa aaa  "
	spacename = strings.TrimSpace(spacename) //删除首尾空格
	fmt.Println(spacename)
	b := "aaaa aa aaa"
	fmt.Println(b)
	if !strings.EqualFold(spacename, b) {
		t.FailNow()
	}
}

func TestRandRobotNamesForVN(t *testing.T) {
	//MoveToRoot()
	//defer logs.Close()
	//LoadNameVN("conf/data/namevi.data")
	//
	//startTime := time.Now().UnixNano()
	//result := randRobotNamesByLimit(10000)
	//endTime := time.Now().UnixNano()
	//println((endTime - startTime) / 1e6)
	//if len(result) != 10000 {
	//	t.Error("result =", len(result))
	//	t.FailNow()
	//}
	//nameSet := make(map[string]struct{})
	//for _, name := range result {
	//	nameSet[name] = struct{}{}
	//}
	//if len(nameSet) != 10000 {
	//	t.Error("map=", len(nameSet))
	//	t.FailNow()
	//}
	//fmt.Println(result)
}

func BenchmarkRandRobotNamesForVN(b *testing.B) {
	//MoveToRoot()
	//defer logs.Close()
	//LoadNameVN("conf/data/namevi.data")
	//
	//successCount := 0
	//failCount := 0
	//b.N = 100
	//for i := 0; i < b.N; i++ {
	//	ret := randRobotNamesByLimit(10000)
	//	if len(ret) == 10000 {
	//		successCount++
	//	} else {
	//		failCount++
	//	}
	//}
	//fmt.Println("bench result: ", successCount, failCount)
}

func TestKoreaNameLen(t *testing.T) {
	fmt.Println(len("귀장지"))
}
