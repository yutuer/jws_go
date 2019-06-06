package logics

func (s *SyncResp) mkGeneralInfo(p *Account) {
	generals := &p.GeneralProfile
	if s.general_need_sync {
		s.SyncGeneral = true
		general_data := generals.GetAllGeneral()
		s.SyncGeneralName = make([]string, len(general_data), len(general_data))
		s.SyncGeneralNum = make([]uint32, len(general_data), len(general_data))
		s.SyncGeneralStarLevel = make([]uint32, len(general_data), len(general_data))
		for i := 0; i < len(general_data); i++ {
			s.SyncGeneralName[i] = general_data[i].Id
			s.SyncGeneralStarLevel[i] = general_data[i].StarLv
			s.SyncGeneralNum[i] = general_data[i].Num
		}
	}

	if s.general_rel_need_sync {
		s.SyncGeneralRel = true
		rels := generals.GetAllGeneralRel()
		s.SyncGeneralRelId = make([]string, len(rels), len(rels))
		s.SyncGeneralRelLevel = make([]uint32, len(rels), len(rels))
		for i := 0; i < len(rels); i++ {
			s.SyncGeneralRelId[i] = rels[i].Id
			s.SyncGeneralRelLevel[i] = rels[i].Level
		}
	}

	if s.SyncGeneralQuestNeed {
		generals.GQListUpdate(p.Profile.GetProfileNowTime())
		s.QuestListId = make([]int64, 0, len(generals.QuestList))
		s.QuestListName = make([]string, 0, len(generals.QuestList))
		s.QuestListReved = make([]bool, 0, len(generals.QuestList))
		for _, gql := range generals.QuestList {
			s.QuestListId = append(s.QuestListId, gql.QuestId)
			s.QuestListName = append(s.QuestListName, gql.QuestCfgId)
			s.QuestListReved = append(s.QuestListReved, gql.IsRec)
		}
		s.QuestListNextRefTime = generals.QuestListNextRefTime
		s.QuestRevId = make([]int64, 0, len(generals.QuestRec))
		s.QuestRevName = make([]string, 0, len(generals.QuestRec))
		s.QuestRevFinishTime = make([]int64, 0, len(generals.QuestRec))
		s.QuestRevGeneralNum = make([]int, 0, len(generals.QuestRec))
		s.QuestRevGenerals = make([]string, 0, len(generals.QuestRec)*3)
		for _, gqr := range generals.QuestRec {
			s.QuestRevId = append(s.QuestRevId, gqr.QuestId)
			s.QuestRevName = append(s.QuestRevName, gqr.QuestCfgId)
			s.QuestRevFinishTime = append(s.QuestRevFinishTime, gqr.FinishTime)
			s.QuestRevGeneralNum = append(s.QuestRevGeneralNum, len(gqr.GeneralIds))
			for _, g := range gqr.GeneralIds {
				s.QuestRevGenerals = append(s.QuestRevGenerals, g)
			}
		}
	}
}
