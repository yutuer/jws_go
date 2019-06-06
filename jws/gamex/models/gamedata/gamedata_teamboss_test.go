package gamedata

import (
	"fmt"
	"testing"
)

func TestCalTBOpenCost(t *testing.T) {
	MoveToRoot()
	loadTBBoxData("conf/data/tbossboxdata.data")
	loadTBBossConfig("conf/data/tbossconfig.data")
	fmt.Println(CalTBOpenCost(301))
}
