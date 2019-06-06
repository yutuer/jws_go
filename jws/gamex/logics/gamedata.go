package logics

import (
	"github.com/golang/protobuf/proto"

	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/secure"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

type GameData struct{}

func NewGameData(r Router) *GameData {
	g := &GameData{}
	r.HandleFunc("GD/GetPlayerDataAtt", g.GetPlayerDataAtt)
	r.HandleFunc("GD/GetItemsDataAtt", g.GetItemsDataAtt)
	return g
}

type RequestGetPlayerData struct {
	Req
	Level int64 `codec:"level"`
}

type ResponseGetPlayerData struct {
	Resp
	Result  string `codec:"result"`
	Result2 string `codec:"result2"`
}

func (gd *GameData) GetPlayerDataAtt(r servers.Request) *servers.Response {
	var request RequestGetPlayerData
	decode(r.RawBytes, &request)

	logs.Trace("data is %v", request)

	const (
		_                  = iota
		CODE_BaseData_Err  // 失败:基本信息编码错误
		CODE_LevelData_Err // 失败:关卡信息编码错误
	)

	resp := &ResponseGetPlayerData{
		Resp: Resp{
			PassthroughID: request.PassthroughID,
			MsgOK:         "ok",
		},
	}

	levelId := request.Level
	basicAtt := gamedata.GetPlayerBasicAtt()
	lvlatt := gamedata.GetPlayerLevelAttr(uint32(levelId))

	basic, err := proto.Marshal(&basicAtt)
	if err != nil {
		logs.Error("GetPlayerDataAtt basicAtt Err %s", err.Error())
		resp.SetCode(CODE_ERR, CODE_BaseData_Err)
	}
	lvlbuf, err := proto.Marshal(lvlatt)
	if err != nil {
		logs.Error("GetPlayerDataAtt basicAtt Err %s", err.Error())
		resp.SetCode(CODE_ERR, CODE_LevelData_Err)
	}

	if err == nil {
		resp.Result = secure.Encode64ForNet(basic)
		resp.Result2 = secure.Encode64ForNet(lvlbuf)
	}

	return &servers.Response{
		Code:     "GD/GetPlayerDataAttResponse",
		RawBytes: encode(resp),
	}
}

type RequestGetItemsDataAttr struct {
	Req
	IdList []string `codec:"ids"`
}

type ResponseGetitemsDataAttr struct {
	Resp
	Result string `codec:"result"`
}

func (gd *GameData) GetItemsDataAtt(r servers.Request) *servers.Response {
	var req RequestGetItemsDataAttr
	decode(r.RawBytes, &req)
	logs.Trace("data is %v", req)

	const (
		_                 = iota
		CODE_ItemData_Err // 失败:Item编码错误
	)

	resp := &ResponseGetitemsDataAttr{
		Resp: Resp{
			PassthroughID: req.PassthroughID,
			MsgOK:         "ok",
		},
	}

	ids := req.IdList
	items := gamedata.GetItemDataByList(ids)
	itemarr, err := proto.Marshal(items)
	if err != nil {
		logs.Error("GetItemsDataAtt error: %s", err.Error())
		resp.SetCode(CODE_ERR, CODE_ItemData_Err)
	}
	resp.Result = secure.Encode64ForNet(itemarr)

	//logs.Info("data result is: %v, %s, %v, arr:%v", itemarr, result, ids, items)
	return &servers.Response{
		Code:     "GD/GetItemsDataAttResponse",
		RawBytes: encode(resp),
	}
}
