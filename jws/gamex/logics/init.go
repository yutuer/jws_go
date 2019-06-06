package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/codec"

	"encoding/json"

	"sort"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logics/notify"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Router interface {
	Handle(pattern string, handler servers.Handler)
	HandleFunc(pattern string, handler func(servers.Request) *servers.Response)
}

//////////////////////////

type reqInterface interface {
	GetPassthroughID() string
	GetTIDs() string
	GetHotVer() int
}

type Req struct {
	PassthroughID string   `codec:"passthrough"`
	DebugTime     int64    `codec:"debugtime,omitempty"`
	_struct       struct{} `codec:",omitempty"`
	HotVer        int      `codec:"hotver"`
	// 教学ID提交
	TIDs string `codec:"tids_"`
}

func (r *Req) GetPassthroughID() string {
	return r.PassthroughID
}

func (r *Req) GetTIDs() string {
	return r.TIDs
}

func (r *Req) GetHotVer() int {
	return r.HotVer
}

type respInterface interface {
	Init(passthroughId, address string, hotVersion int, p *Account)
	SetCode(level, id uint32)
	GetAddress() string
	AccountID() string
	SetMsg(msg string)
	setRng()
}

type Resp struct {
	PassthroughID string   `codec:"passthrough"`
	MsgOK         string   `codec:"msg"`
	Code          uint32   `codec:"code"` //0 is normal, msg should be ok
	RngLastCount  int      `codec:"rnglc"`
	RngLastValue  int      `codec:"rnglv"`
	_struct       struct{} `codec:",omitempty"`

	// As Context
	address              string // 实际上应该在 servers.Request 中提供 TBD by Fanyang
	hotDataClientVersion int    // 客户端发来的热更数据版本
	p                    *Account
}

func (r *Resp) Init(passthroughId, address string, hotVersion int, p *Account) {
	r.PassthroughID = passthroughId
	r.address = address
	r.MsgOK = "ok"
	r.RngLastCount = p.Profile.Rng.Count
	r.RngLastValue = p.Profile.Rng.RandValue
	r.p = p
	r.hotDataClientVersion = hotVersion
}

func (r *Resp) GetAddress() string {
	return r.address
}

func (r *Resp) AccountID() string {
	return r.p.AccountID.String()
}

func (r *Resp) SetMsg(msg string) {
	r.MsgOK = msg
}

func (r *Resp) SetCode(level, id uint32) {
	if r.Code != 0 {
		return
	}
	r.Code = mkCode(level, id)
}

func (r *Resp) setRng() {
	r.RngLastCount = r.p.Profile.Rng.Count
	r.RngLastValue = r.p.Profile.Rng.RandValue
}

func rpcError(r respInterface, code uint32) *servers.Response {
	address := r.GetAddress()
	logs.SentryLogicCritical(
		r.AccountID(),
		"Resp Err %s By %d", address, code)
	r.SetMsg("no")
	r.SetCode(CODE_ERR, code)
	r.setRng()
	return &servers.Response{
		Code:     address,
		RawBytes: encode(r),
	}
}

func rpcErrorWithMsg(r respInterface, code uint32, msg string) *servers.Response {
	address := r.GetAddress()
	logs.SentryLogicCritical(
		r.AccountID(),
		"Resp Err %s By %d %s", address, code, msg)
	r.SetMsg("no")
	r.SetCode(CODE_ERR, code)
	r.setRng()
	return &servers.Response{
		Code:     address,
		RawBytes: encode(r),
	}
}

func rpcWarn(r respInterface, code uint32) *servers.Response {
	address := r.GetAddress()
	logs.Warn("[%s][%s]Resp Warn By %d",
		r.AccountID(), address, code)
	r.SetMsg("no")
	r.SetCode(CODE_WARN, code)
	r.setRng()
	return &servers.Response{
		Code:     address,
		RawBytes: encode(r),
	}
}

func rpcSuccess(r respInterface) *servers.Response {
	return rpcSuccessForceDB(r, false)
}

func rpcSuccessForceDB(r respInterface, force bool) *servers.Response {
	address := r.GetAddress()

	if !game.Cfg.IsRunModeProd() {
		jsons, _ := json.Marshal(r)
		logs.Debug("[%s][%s][Msg]Resp data is %v",
			r.AccountID(),
			address,
			string(jsons))
	}

	r.SetCode(CODE_SUCCESS, 0)
	r.setRng()
	return &servers.Response{
		Code:          address,
		RawBytes:      encode(r),
		ForceDBChange: force,
	}
}

func decode(raw []byte, out interface{}) {
	codec.Decode(raw, out)
}

func encode(value interface{}) []byte {
	return codec.Encode(value)
}

func decAsDict(raw []byte) map[string]interface{} {
	return codec.DecAsDict(raw)
}

func encAsDict(value interface{}) []byte {
	return codec.EncAsDict(value)
}

const CODE_SUCCESS = 0

const (
	CODE_WARN = iota
	_
	CODE_ERR
	CODE_ERR_UNKNOWN
	CODE_ERR_KICK
)

func mkCode(s, v uint32) uint32 {
	return s*100 + v
}

func initReqRsp(
	rsp_address string,
	req_data []byte,
	req reqInterface,
	rsp respInterface,
	p *Account) {
	decode(req_data, req)

	rsp.Init(req.GetPassthroughID(), rsp_address, req.GetHotVer(), p)

	if !game.Cfg.IsRunModeProd() {
		jsons, _ := json.Marshal(req)
		logs.Debug("[%s][%s][Msg]Req data is %v",
			p.AccountID.String(),
			rsp_address,
			string(jsons))
	}
}

// 用来发送主动推送的push协议
func sendPush(
	rsp_address string,
	push notify.NotifySyncMsg,
	p *Account) {
	push.SetAddr(rsp_address)
	p.Account.SendRespByPush(&push)
}

type SyncRespWithRewards struct {
	SyncResp
	RewardID    []string `codec:"rids"`
	RewardCount []uint32 `codec:"rcs"`
	RewardData  []string `codec:"rds"`
}

func (s *SyncRespWithRewards) AddResReward(g *gamedata.CostData2Client) {
	if g == nil {
		return
	}
	for i := 0; i < g.Len(); i++ {
		ok, it, c, d, _ := g.GetItem(i)
		if ok {
			s.addReward2RespDatas(it, c, d)
		}
	}
}

func (s *SyncRespWithRewards) addReward2RespDatas(id string, c uint32, data string) {
	isJade, _ := gamedata.IsJade(id)
	if isJade {
		for idx, rid := range s.RewardID {
			if rid == id {
				s.RewardCount[idx] += c
				return
			}
		}
	}
	s.RewardID = append(s.RewardID, id)
	s.RewardCount = append(s.RewardCount, c)
	s.RewardData = append(s.RewardData, data)
}

func (s *SyncRespWithRewards) MergeReward() {
	_RewardID := make([]string, 0, len(s.RewardID))
	_RewardCount := make([]uint32, 0, len(s.RewardCount))
	_RewardData := make([]string, 0, len(s.RewardData))
	_r2n := make(map[string]uint32, len(s.RewardID))

	for i, id := range s.RewardID {
		data := s.RewardData[i]
		if data != "" || !gamedata.IsFixedIDItemID(id) {
			_RewardID = append(_RewardID, id)
			_RewardCount = append(_RewardCount, s.RewardCount[i])
			_RewardData = append(_RewardData, s.RewardData[i])
			continue
		}
		n, ok := _r2n[id]
		if !ok {
			_r2n[id] = s.RewardCount[i]
		} else {
			_r2n[id] = s.RewardCount[i] + n
		}
	}

	for r, n := range _r2n {
		_RewardID = append(_RewardID, r)
		_RewardCount = append(_RewardCount, n)
		_RewardData = append(_RewardData, "")
	}

	// 整理显示顺序，不然会随机显示
	_blanceReward := make(awards, len(_RewardID))
	for n := 0; n < len(_RewardID); n++ {
		_blanceReward[n] = award{_RewardID[n], n}
	}

	sort.Sort(_blanceReward)

	s.RewardID = make([]string, 0, len(_RewardID))
	s.RewardCount = make([]uint32, 0, len(_RewardID))
	s.RewardData = make([]string, 0, len(_RewardID))
	for _, e := range _blanceReward {
		s.RewardID = append(s.RewardID, e.awardid)
		s.RewardCount = append(s.RewardCount, _RewardCount[e.Idx])
		s.RewardData = append(s.RewardData, _RewardData[e.Idx])
	}
}

type award struct {
	awardid string
	Idx     int
}

type awards []award

func (pq awards) Len() int { return len(pq) }

func (pq awards) Less(i, j int) bool {
	return pq[i].awardid < pq[j].awardid
}

func (pq awards) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

// anticheat
type ReqWithAnticheat struct {
	Req
	Hackjson string `codec:"hackjson"`
}

type RespWithAnticheat struct {
	SyncResp
	CheatedIndex []int `codec:"cheat_idx"` // 哪项作弊检查未通过，-1无效
}

type SyncRespWithRewardsAnticheat struct {
	SyncRespWithRewards
	CheatedIndex []int `codec:"cheat_idx"` // 哪项作弊检查未通过，-1无效
}

func (p *Account) AntiCheatCheck(rsp *SyncRespWithRewardsAnticheat, cheat *ReqWithAnticheat, levelCost int64, typ string) *servers.Response {
	if cheat.Hackjson != "" {
		hacks := []float32{}
		if err := json.Unmarshal([]byte(cheat.Hackjson), &hacks); err != nil {
			if err != nil {
				return rpcErrorWithMsg(rsp,
					1,
					fmt.Sprintf("hack unmarshal err %s", err.Error()))
			}
		}

		rsp.CheatedIndex = p.AntiCheat.CheckFightRelAll(
			p.AccountID.String(),
			hacks,
			p.Account,
			typ,
			levelCost)
		if len(rsp.CheatedIndex) > 0 {
			if isTimeCheat(rsp.CheatedIndex) {
				return rpcWarn(rsp, errCode.YouTimeCheat)
			}
			return rpcWarn(rsp, errCode.YouCheat)
		} else {
			return nil
		}
	} else {
		logs.Info("[Antichest-Empty] Trial  acid %s req.Hackjson is empty",
			p.AccountID.String())
		return nil

	}
}

func isTimeCheat(index []int) bool {
	for _, item := range index {
		if item == account.CheckerPassTime {
			return true
		}
	}
	return false
}

func (p *Account) AntiCheatCheckWithRewards(rsp *SyncRespWithRewardsAnticheat, cheat *ReqWithAnticheat, levelCost int64, typ string) uint32 {
	if cheat.Hackjson != "" {
		hacks := []float32{}
		if err := json.Unmarshal([]byte(cheat.Hackjson), &hacks); err != nil {
			if err != nil {
				return errCode.YouCheat
			}
		}

		rsp.CheatedIndex = p.AntiCheat.CheckFightRelAll(
			p.AccountID.String(),
			hacks,
			p.Account,
			typ,
			levelCost)
		if len(rsp.CheatedIndex) > 0 {
			if isTimeCheat(rsp.CheatedIndex) {
				return errCode.YouTimeCheat
			}
			return errCode.YouCheat
		} else {
			return 0
		}
	} else {
		logs.Info("[Antichest-Empty] Trial  acid %s req.Hackjson is empty",
			p.AccountID.String())
		return 0

	}
}

func (p *Account) AntiCheatCheckWithCode(rsp *RespWithAnticheat, cheat *ReqWithAnticheat, levelCost int64, typ string) uint32 {
	if cheat.Hackjson != "" {
		hacks := []float32{}
		if err := json.Unmarshal([]byte(cheat.Hackjson), &hacks); err != nil {
			if err != nil {
				return errCode.YouCheat
			}
		}

		rsp.CheatedIndex = p.AntiCheat.CheckFightRelAll(
			p.AccountID.String(),
			hacks,
			p.Account,
			typ,
			levelCost)
		if len(rsp.CheatedIndex) > 0 {
			if isTimeCheat(rsp.CheatedIndex) {
				return errCode.YouTimeCheat
			}
			return errCode.YouCheat
		} else {
			return 0
		}
	} else {
		logs.Info("[Antichest-Empty] Trial  acid %s req.Hackjson is empty",
			p.AccountID.String())
		return 0

	}
}
