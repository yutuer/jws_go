package gen2csharp

import "vcs.taiyouxi.net/platform/x/tiprotogen/def"

func (g *genner2Csharp) GetReqBase(data *dsl.ProtoDef) string {
	if data.Cheat {
		return "AbstractReqWithAnticheatMsg"
	}
	return "AbstractReqMsg"
}

func (g *genner2Csharp) GetRspBase(data *dsl.ProtoDef) string {
	if data.Cheat {
		switch data.Rsp.Base {
		case "WithSync":
			return "AbstractRspWithAnticheatMsg"
		case "WithRewards":
			return "AbstractRspWithRewardAnticheatMsg"
		default:
			return "AbstractRspWithAnticheatMsg"
		}
	} else {
		switch data.Rsp.Base {
		case "WithSync":
			return "AbstractRspMsg"
		case "WithRewards":
			return "AbstractRspWithRewardMsg"
		default:
			return "AbstractRspMsg"
		}
	}
}
