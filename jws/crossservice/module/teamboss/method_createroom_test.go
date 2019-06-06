package teamboss

import (
	"testing"

	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"

	"github.com/stretchr/testify/assert"
)

func TestMethodCreateRoom_Do(t *testing.T) {
	tb := module.LoadModulesList[0].NewModule(1)
	tb.Start()
	defer tb.Stop()

	// 正常流程
	param := genCreateRoomParam(level, "Team Liquid", "ACIDTest001")
	trans := new(module.Transaction)

	c := tb.GetMethod("createroom")
	errCode, ret := c.Do(*trans, param)

	respInfo := ret.(RetCreateRoom).Info.Info
	t.Logf("Response info: %v", respInfo)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetCreateRoom).Info.Code)
	assert.Equal(t, 0, respInfo.RoomStatus)
	assert.Equal(t, "ACIDTest001", respInfo.LeadID)
	assert.EqualValues(t, level, respInfo.Level)

	// 异常流程
}
