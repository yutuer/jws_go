package logics

import "vcs.taiyouxi.net/platform/planx/servers"

func handleAllGenFunc(r *servers.Mux, p *Account) {
	r.HandleFunc("Attr/TestProtoReq", p.TestProto)

	r.HandleFunc("Guild/AssignGuildInventoryReq", p.AssignGuildInventory)
	r.HandleFunc("Guild/GetGuildInventoryLogReq", p.GetGuildInventoryLog)
	r.HandleFunc("Attr/ApplyGuildInventoryReq", p.ApplyGuildInventory)
	r.HandleFunc("Attr/GetGuildInventoryApplyListReq", p.GetGuildInventoryApplyList)
	r.HandleFunc("Attr/ApproveGuildInventoryReq", p.ApproveGuildInventory)
	r.HandleFunc("Attr/ExchangeGuildInventoryReq", p.ExchangeGuildInventory)

	r.HandleFunc("PlayerAttr/IAPPaySuccessReq", p.IAPPaySuccess)
	r.HandleFunc("PlayerAttr/AwardLevelGiftReq", p.AwardLevelGift)

	r.HandleFunc("Attr/GuildBossLockReqReq", p.GuildBossLockReq)
	r.HandleFunc("Attr/GuildBossBeginReqReq", p.GuildBossBeginReq)
	r.HandleFunc("Attr/GuildBossFinishReqReq", p.GuildBossFinishReq)
	r.HandleFunc("Attr/GuildBossHeartBeatReqReq", p.GuildBossHeartBeatReq)

	r.HandleFunc("Guild/AddGuildSciencePointReq", p.AddGuildSciencePoint)
	r.HandleFunc("Guild/GuildSciencePointLogReq", p.GuildSciencePointLog)

	r.HandleFunc("Attr/SetSimplePvpDefAvatarReq", p.SetSimplePvpDefAvatar)
	r.HandleFunc("Attr/SkipTutorialReqReq", p.SkipTutorialReq)

	r.HandleFunc("PlayerAttr/AwardMarketActivityReq", p.AwardMarketActivity)
	r.HandleFunc("Attr/NewAddDestinyGeneralLvReq", p.NewAddDestinyGeneralLv)
	r.HandleFunc("Attr/FenghuoRoomListenEndReq", p.FenghuoRoomListenEnd)
	r.HandleFunc("Attr/FenghuoRoomListenStartReq", p.FenghuoRoomListenStart)
	r.HandleFunc("Attr/TeamPVESingleEnterReq", p.TeamPVESingleEnter)
	r.HandleFunc("Attr/TeamPVESingleLvPassReq", p.TeamPVESingleLvPass)
	r.HandleFunc("Attr/FenghuoRoomCancelReadyReq", p.FenghuoRoomCancelReady)
	r.HandleFunc("Attr/FenghuoRoomChangeMasterReq", p.FenghuoRoomChangeMaster)
	r.HandleFunc("Attr/FenghuoRoomChatReq", p.FenghuoRoomChat)
	r.HandleFunc("Attr/FenghuoRoomCreateReq", p.FenghuoRoomCreate)
	r.HandleFunc("Attr/FenghuoRoomUpdateReq", p.FenghuoRoomUpdate)
	r.HandleFunc("Attr/FenghuoRoomDeleteReq", p.FenghuoRoomDelete)
	r.HandleFunc("Attr/FenghuoRoomReadyReq", p.FenghuoRoomReady)
	r.HandleFunc("Attr/FenghuoRoomStartFightReq", p.FenghuoRoomStartFight)
	r.HandleFunc("Attr/FenghuoRoomMkFightRewardsReq", p.FenghuoRoomMkFightRewards)
	r.HandleFunc("Attr/FenghuoRoomEnterReq", p.FenghuoRoomEnter)
	r.HandleFunc("Attr/FenghuoRoomLeaveReq", p.FenghuoRoomLeave)

	r.HandleFunc("Attr/HeroTalentLevelUpReq", p.HeroTalentLevelUp)
	r.HandleFunc("Attr/HeroSoulLevelUpReq", p.HeroSoulLevelUp)
	r.HandleFunc("Attr/EatBaoziReq", p.EatBaozi)

	r.HandleFunc("Attr/GuildGatesEnemyInspireReq", p.GuildGatesEnemyInspire)

	r.HandleFunc("Attr/BuyExpItemReq", p.BuyExpItem)

	r.HandleFunc("Attr/WannaHeroReq", p.WannaHero)

	r.HandleFunc("Attr/ChangeHeroTeamReq", p.ChangeHeroTeam)

	r.HandleFunc("Attr/FenghuoClientLogReq", p.FenghuoClientLog)

	r.HandleFunc("Attr/HeroGachaRaceChestReq", p.HeroGachaRaceChest)
	r.HandleFunc("Attr/HeroGachaRaceGetReq", p.HeroGachaRaceGet)
	r.HandleFunc("Attr/StarheroReq", p.Starhero)

	//GVG军团战
	r.HandleFunc("Attr/GVGEnterCityReq", p.GVGEnterCity)
	r.HandleFunc("Attr/GVGMatchEnemyReq", p.GVGMatchEnemy)
	r.HandleFunc("Attr/GVGMatchQueryReq", p.GVGMatchQuery)
	r.HandleFunc("Attr/GVGEndFightReq", p.GVGEndFight)
	r.HandleFunc("Attr/GVGLeaveCityReq", p.GVGLeaveCity)
	r.HandleFunc("Attr/GVGGetGuildInfoReq", p.GVGGetGuildInfo)
	r.HandleFunc("Attr/GVGGetGuildMemberInfoReq", p.GVGGetGuildMemberInfo)
	r.HandleFunc("Attr/GVGGetPlayerInfoReq", p.GVGGetPlayerInfo)
	r.HandleFunc("Attr/GVGCancelMatchReq", p.GVGCancelMatch)
	r.HandleFunc("Attr/GetGVGCityDataReq", p.GetGVGCityData)
	r.HandleFunc("Attr/GetGVGGuildScoreReq", p.GetGVGGuildScore)
	r.HandleFunc("Attr/GetGVGGuildRankReq", p.GetGVGGuildRank)
	r.HandleFunc("Attr/GetGVGGuildTotalScoreReq", p.GetGVGGuildTotalScore)

	//Friend System
	r.HandleFunc("Attr/AddBlackListReq", p.AddBlackList)
	r.HandleFunc("Attr/AddFriendReq", p.AddFriend)
	r.HandleFunc("Attr/FindFriendReq", p.FindFriend)
	r.HandleFunc("Attr/RemoveBlackListReq", p.RemoveBlackList)
	r.HandleFunc("Attr/RemoveFriendReq", p.RemoveFriend)
	r.HandleFunc("Attr/UpdateRecentPlayerReq", p.UpdateRecentPlayer)
	r.HandleFunc("Attr/GetRecommendPlayerReq", p.GetRecommendPlayer)

	//军团改名
	r.HandleFunc("Attr/RenameGuildReq", p.RenameGuild)

	r.HandleFunc("Attr/GuildWorshipAvatarInfoReq", p.GuildWorshipAvatarInfo)
	r.HandleFunc("Attr/StartHeroDiffFightReq", p.StartHeroDiffFight)
	r.HandleFunc("Attr/OverHeroDiffFightReq", p.OverHeroDiffFight)
	r.HandleFunc("Attr/BuyGuildBossAbsentRewardReq", p.BuyGuildBossAbsentReward)
	//激活专属兵器
	r.HandleFunc("Attr/ActivateExclusiveWeaponReq", p.ActivateExclusiveWeapon)
	r.HandleFunc("Attr/EvolveExclusiveWeaponReq", p.EvolveExclusiveWeapon)
	r.HandleFunc("Attr/PromoteExclusiveWeaponReq", p.PromoteExclusiveWeapon)
	r.HandleFunc("Attr/ResetExclusiveWeaponReq", p.ResetExclusiveWeapon)

	//军团膜拜
	r.HandleFunc("Attr/WorshipPlayerReq", p.WorshipPlayer)
	r.HandleFunc("Attr/WorshipLogReq", p.WorshipLog)
	r.HandleFunc("Attr/WorshipBoxReq", p.WorshipBox)
	//无双争霸防守阵容
	r.HandleFunc("Attr/SetWSPVPDefenseFormationReq", p.SetWSPVPDefenseFormation)
	r.HandleFunc("Attr/GetMatchOpponentReq", p.GetMatchOpponent)
	r.HandleFunc("Attr/LockWSPVPBattleReq", p.LockWSPVPBattle)
	r.HandleFunc("Attr/BeginWSPVPBattleReq", p.BeginWSPVPBattle)
	r.HandleFunc("Attr/EndWSPVPBattleReq", p.EndWSPVPBattle)
	r.HandleFunc("Attr/UnlockWSPVPBattleReq", p.UnlockWSPVPBattle)
	r.HandleFunc("Attr/GetWSPVPBattleLogReq", p.GetWSPVPBattleLog)
	r.HandleFunc("Attr/ClaimWSPVPRewardReq", p.ClaimWSPVPReward)
	r.HandleFunc("Attr/GetWSPVPPlayerInfoReq", p.GetWSPVPPlayerInfo)
	//
	r.HandleFunc("Attr/OppoSignReq", p.OppoSign)
	r.HandleFunc("Attr/OppoDailyQuestReq", p.OppoDailyQuest)
	r.HandleFunc("Attr/OppoLoginReq", p.OppoLogin)

	//新手引导跳过
	r.HandleFunc("Attr/SetNewHandIgnoreReq", p.SetNewHandIgnore)
	//
	r.HandleFunc("Attr/GetOpRankReq", p.GetOpRank)
	r.HandleFunc("Attr/GetOpRankRewardInfoReq", p.GetOpRankRewardInfo)

	r.HandleFunc("Attr/ExchangePropReq", p.ExchangeProp)
	r.HandleFunc("Attr/GetExchangeShopInfoReq", p.GetExchangeShopInfo)

	//押运粮草活动
	r.HandleFunc("Attr/CSRobPlayerInfoReq", p.CSRobPlayerInfo)
	r.HandleFunc("Attr/CSRobGetRecordsReq", p.CSRobGetRecords)
	r.HandleFunc("Attr/CSRobSetFormationReq", p.CSRobSetFormation)
	r.HandleFunc("Attr/CSRobRandCarReq", p.CSRobRandCar)
	r.HandleFunc("Attr/CSRobBuildCarReq", p.CSRobBuildCar)
	r.HandleFunc("Attr/CSRobSendHelpReq", p.CSRobSendHelp)
	r.HandleFunc("Attr/CSRobReceiveHelpReq", p.CSRobReceiveHelp)
	r.HandleFunc("Attr/CSRobRobItReq", p.CSRobRobIt)
	r.HandleFunc("Attr/CSRobRobResultReq", p.CSRobRobResult)
	r.HandleFunc("Attr/CSRobSendMarqueeReq", p.CSRobSendMarquee)
	r.HandleFunc("Attr/CSRobGuildInfoReq", p.CSRobGuildInfo)
	r.HandleFunc("Attr/CSRobGuildListReq", p.CSRobGuildList)
	r.HandleFunc("Attr/CSRobTeamsListReq", p.CSRobTeamsList)
	r.HandleFunc("Attr/CSRobGuildRankReq", p.CSRobGuildRank)
	r.HandleFunc("Attr/CSRobNationalityRankReq", p.CSRobNationalityRank)
	r.HandleFunc("Attr/CSRobAutoAcceptSetReq", p.CSRobAutoAcceptSet)
	//激活指定的宿命
	r.HandleFunc("Attr/ActivateHeroDestinyReq", p.ActivateHeroDestiny)
	//扫荡远征
	r.HandleFunc("Attr/ExpeditionSweepReq", p.ExpeditionSweep)

	r.HandleFunc("Attr/HeroDiffSweepReq", p.HeroDiffSweep)
	r.HandleFunc("Attr/BuyItemReq", p.BuyItem)

	//黑盒宝箱抽奖
	r.HandleFunc("Attr/DrawBlackGachaReq", p.DrawBlackGacha)
	r.HandleFunc("Attr/GetBlackGachaInfoReq", p.GetBlackGachaInfo)
	r.HandleFunc("Attr/ClaimBlackGachaExtraRewardReq", p.ClaimBlackGachaExtraReward)

	//星图系统
	r.HandleFunc("Attr/AstrologyGetInfoReq", p.AstrologyGetInfo)
	r.HandleFunc("Attr/AstrologyIntoReq", p.AstrologyInto)
	r.HandleFunc("Attr/AstrologyDestroyInHeroReq", p.AstrologyDestroyInHero)
	r.HandleFunc("Attr/AstrologyDestroyInBagReq", p.AstrologyDestroyInBag)
	r.HandleFunc("Attr/AstrologyDestroySkipReq", p.AstrologyDestroySkip)
	r.HandleFunc("Attr/AstrologySoulUpgradeReq", p.AstrologySoulUpgrade)
	r.HandleFunc("Attr/AstrologyAugurReq", p.AstrologyAugur)

	//
	r.HandleFunc("Attr/GiveGiftToFriendReq", p.GiveGiftToFriend)
	r.HandleFunc("Attr/ReceiveGiftFromFriendReq", p.ReceiveGiftFromFriend)
	r.HandleFunc("Attr/BatchGiveGift2FriendReq", p.BatchGiveGift2Friend)
	r.HandleFunc("Attr/BatchReceiveGiftFromFriendReq", p.BatchReceiveGiftFromFriend)
	r.HandleFunc("Attr/GetReceiveGiftInfoReq", p.GetReceiveGiftInfo)
	r.HandleFunc("Attr/GetFriendGiftAcIDReq", p.GetFriendGiftAcID)
	//
	r.HandleFunc("Attr/BindMailRewardsReq", p.BindMailRewards)

	r.HandleFunc("Attr/GetWBInfoReq", p.GetWBInfo)
	r.HandleFunc("Attr/BeginWBReq", p.BeginWB)
	r.HandleFunc("Attr/EndWBReq", p.EndWB)
	r.HandleFunc("Attr/GetWBRankInfoReq", p.GetWBRankInfo)
	r.HandleFunc("Attr/UpdateBattleInfoReq", p.UpdateBattleInfo)
	r.HandleFunc("Attr/UseBuffReq", p.UseBuff)
	r.HandleFunc("Attr/GetWBRankRewardsReq", p.GetWBRankRewards)
	r.HandleFunc("Attr/GetWBPlayerDetailReq", p.GetWBPlayerDetail)
	r.HandleFunc("Attr/SetBuyBuffReminderReq", p.SetBuyBuffReminder)
	//离线资源
	r.HandleFunc("Attr/GetOfflineRecoverInfoReq", p.GetOfflineRecoverInfo)
	r.HandleFunc("Attr/ClaimOfflineRecoverRewardReq", p.ClaimOfflineRecoverReward)
	//武将碎片兑换令牌
	r.HandleFunc("Attr/ExchangeHeroPieceReq", p.ExchangeHeroPiece)
	r.HandleFunc("Attr/DrawHeroPieceGachaReq", p.DrawHeroPieceGacha)
	//英雄灵宠
	r.HandleFunc("Attr/SetStateOfShowMagicPetReq", p.SetStateOfShowMagicPet)
	r.HandleFunc("Attr/ShowMagicPetReq", p.ShowMagicPet)
	r.HandleFunc("Attr/MagicPetLevUpReq", p.MagicPetLevUp)
	r.HandleFunc("Attr/MagicPetStarUpReq", p.MagicPetStarUp)
	r.HandleFunc("Attr/MagicPetChangeTalentReq", p.MagicPetChangeTalent)
	r.HandleFunc("Attr/MagicPetSaveTalentReq", p.MagicPetSaveTalent)
	//获取组队BOSS的队伍信息
	r.HandleFunc("Attr/GetTBTeamListReq", p.GetTBTeamList)
	r.HandleFunc("Attr/TBTeamReadyReq", p.TBTeamReady)
	r.HandleFunc("Attr/CreatTBTeamReq", p.CreatTBTeam)
	r.HandleFunc("Attr/TBTeamJoinSettingReq", p.TBTeamJoinSetting)
	r.HandleFunc("Attr/JoinTBTeamReq", p.JoinTBTeam)
	r.HandleFunc("Attr/GetRedBoxCostHCReq", p.GetRedBoxCostHC)
	r.HandleFunc("Attr/LeaveTBTeamReq", p.LeaveTBTeam)
	r.HandleFunc("Attr/TBTeamKickReq", p.TBTeamKick)
	r.HandleFunc("Attr/TBChooseHeroReq", p.TBChooseHero)
	r.HandleFunc("Attr/GetTBMemberInfoReq", p.GetTBMemberInfo)

	//打开组队BOSS仓库
	r.HandleFunc("Attr/TBOpenStorageReq", p.TBOpenStorage)
	r.HandleFunc("Attr/TBOpenBoxReq", p.TBOpenBox)
	r.HandleFunc("Attr/TBDelBoxReq", p.TBDelBox)
	//组队BOSS战结束
	r.HandleFunc("Attr/TBBattleEndReq", p.TBBattleEnd)
	r.HandleFunc("Attr/TBBattleStartReq", p.TBBattleStart)

	//战阵系统
	r.HandleFunc("Attr/ShowBattleArmyReq", p.ShowBattleArmy)
	r.HandleFunc("Attr/BattleArmyLevUpReq", p.BattleArmyLevUp)
	r.HandleFunc("Attr/BattleArmyChoiceAvatarIDReq", p.BattleArmyChoiceAvatarID)

	//推特、line分享
	r.HandleFunc("Attr/TwitterShareReq", p.TwitterShare)
	r.HandleFunc("Attr/LineShareReq", p.LineShare)

	//直购礼包
	r.HandleFunc("Attr/GetPackageInfoReq",p.GetPackageInfo)
	r.HandleFunc("Attr/GetSpecialPackageInfoReq",p.GetSpecialPackageInfo)
	r.HandleFunc("Attr/ReceiveConditionPackageReq",p. ReceiveConditionPackage)
	r.HandleFunc("Attr/CloseSendInfoReq",p.CloseSendInfo)

	//幸运转盘
	r.HandleFunc("Attr/WheelShowInfoReq", p.WheelShowInfo)
	r.HandleFunc("Attr/UseWheelOneReq", p.UseWheelOne)
}