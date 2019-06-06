package gen2golang

import "vcs.taiyouxi.net/platform/x/tiprotogen/def"

func (g *genner2golang) GetReqBase(data *dsl.ProtoDef) string {
	if data.Cheat {
		return "ReqWithAnticheat"
	}
	return "Req"
}

func (g *genner2golang) GetRspBase(data *dsl.ProtoDef) string {
	if data.Cheat {
		switch data.Rsp.Base {
		case "WithSync":
			return "SyncResp"
		case "WithRewards":
			return "SyncRespWithRewardsAnticheat"
		default:
			return "RespWithAnticheat"
		}
	} else {
		switch data.Rsp.Base {
		case "WithSync":
			return "SyncResp"
		case "WithRewards":
			return "SyncRespWithRewards"
		default:
			return "SyncResp"
		}

	}
}
