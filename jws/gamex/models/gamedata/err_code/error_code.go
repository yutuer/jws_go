package errCode

// 返回码定义, 客户端需要1XX返回码不重复

const (
	_                    = iota
	RedPacketNotFound    // 1 不存在的红包
	RedPacketHasGrabbed  // 2 已经抢过这个红包了
	RedPacketErrorBoxId  // 3 错误的宝箱ID
	RedPacketHasClaimed  // 4 已经领过这个宝箱了
	RedPacketCannotClaim // 5 条件未达成， 不能领
	RedPacketReachMax    // 6 红包已经达到上限了

	AddFriendError    // 7 好友系统 添加好友错误
	AddBlackListError // 8 好友系统 添加黑名单错误
	NoFindPlayer      // 9 好友系统 没有找到指定玩家
	GuildRenameError  // 10 军团改名失败
	AddSelfFriend     // 11 好友系统 添加自己为好友
	AddSelfBlackList  // 12 好友系统，将自己拉黑

	GuildGatesEnemyGetRewardErrByNoJoin // 13 您没有参加本次兵临城下，无法领奖
	GuildPosChangeFailedBySamePosition  // 14 该成员已经是该职位了

	HeroTeamWarn           // 15 出阵阵容错误
	HeroDiffHeroAlreadyUse // 16 出奇制胜 武将已经使用过
	RewardFail             // 17 奖励失败

	GuildBossNotAllPassed // 18 BOSS没有全部打完
	GuildBossBigUnlocked  // 19 大BOSS条件未达成

	WSPVPErrorLockAcid       // 20 错误的对手
	WSPVPHasLocked           // 21 该对手已经被锁定
	WSPVPOpponentRankChanged // 22 对手排名已经发生变化
	WSPVPOpponentLockTimeOut // 23 该对手的锁定超时
	WSPVPFormationError      // 24 不合法的阵型
	WSPVPDBError             // 25 DB错误，服务器内部错误
	WSPVPHasClaimedReward    // 26 已经领过奖励了
	WSPVPFailToClaimReward   // 27 条件不足，不能领取奖励
	WSPVPNotTimesLeft        // 28 购买或者刷新次数不足
	WSPVPCannotGetRobotInfo  // 29 无法查看排行内容
	WSPVPNotInRankYet        // 30 已经不在排行帮上了

	OppoSignError       // 31 oppo 签到失败
	OppoDailyQuestError // 32 oppo 每日任务失败
	WSPVPLockingTime    // 33 不在可玩时间范围内

	//公共错误...防止错误码增长过快
	CommonInner          // 34 内部错误
	CommonInitFailed     // 35 初始化失败
	CommonLessMoney      // 36 货币不足
	CommonCountLimit     // 37 次数达到限制
	CommonNotInTime      // 38 不在时间范围内
	CommonConditionFalse // 39 条件不满足
	CommonInvalidParam   // 40 非法参数
	CommonMaxLimit       // 41 达到最大值

	CommonEnd = iota + 59 - CommonInvalidParam //60 公共错误码边界

	HeroDestinyHasActivated   // 61 已经激活这个宿命了
	HeroDestinyConditionError // 62 宿命条件不足

	ALLFRIENDHADGIVEN = 63 // 63 //所有好友已经赠送过礼物

	TBossNeedLvIsNotEnough    = 64 //64  等级不够
	TBossJoinRoomFailed       = 65 //65  加入队伍失败
	TBossJoinTimeTooShort     = 66 //66  距离时间过短
	TBossRoomIsNotExist       = 67 //67  房间不存在
	TBossStartBattleFailed    = 68 //68  开始战斗失败
	TBossHeroChooseOccupied   = 69 //69  已有人选择该种类武将
	TBossKickOneIsNotExist    = 70 //70  要踢的人已经不在
	TBossRoomIsFull           = 71 //71  组队boss房间已满
	TBossAreadyTickRedBox     = 72 //72  另一人已经勾选红宝箱
	TBossRoomInBattle         = 73 //73  房间已经在战斗中
	TBossRoomCantKickInBattle = 74 //74  踢人失败，房间在战斗中
	TBossRoomCantEnter        = 75 //75  权限不足仅邀请能进入
	TBossReadyFailed          = 76 //76  准备失败

)

const (
	none = iota

	RecodeGiftCodeTimeout                 = iota + 100 // 101 兑换码 尝试领取已经过期的礼包会被告知“该礼包已经过期”；
	RecodeGiftCodeTimeNoStart                          // 102 兑换码 尝试领取有效期尚未开始的礼包会被告知“兑换活动尚未开始”；
	RecodeGiftCodeUsed                                 // 103 兑换码 尝试使用用过的兑换码会被告知“该兑换码已被用过，已经失效”；
	RecodeGiftCodeBatchHasExchange                     // 104 兑换码 在领过该批礼包后尝试使用同一批另一组的兑换码，会被告知“你已经领取过XXX大礼包，不能再领”。
	RecodeGiftCodeFormatErr                            // 105 兑换码 兑换码格式错误
	RecodeGiftCodeDataErr                              // 106 兑换码 兑换码所指向的信息错误
	RenameSensitve                                     // 107 改名   敏感词或特殊字符检查失败
	RenameNameHasExit                                  // 108 改名   名字已经存在
	ChatEnterWarn                                      // 109 聊天服务 进入失败
	RecodeGiftCodeBindErr                              // 110 兑换码 该兑换码不能用在当前服务器
	AndroidPayOrderTryTimeOut                          // 111 尝试抓取android未完成的付费订单时，已经过了48小时，超时了
	GuildAlreadyInGuild                                // 112 公会 玩家已在公会中
	GuildNameRepeat                                    // 113 公会 公会名字重复
	GuildNotFound                                      // 114 公会 公会不存在
	GuildFull                                          // 115 公会 公会满
	GuildApplyFull                                     // 116 公会 公会申请列表满
	GuildPositionErr                                   // 117 公会 职务权力不足
	GuildPlayerAlreadyInOther                          // 118 公会 玩家已再其他公会
	GuildPlayerNotFound                                // 119 公会 玩家不在本公会中
	GuildApplicantNotFound                             // 120 公会 申请者不存在了
	GuildChiefNotQuit                                  // 121 公会 会长不能退出
	GuildPlayerNotIn                                   // 122 公会 玩家不在公会中
	GuildWordIllegal                                   // 123 公会 敏感词或特殊字符检查失败
	GuildMemLevelNotEnough                             // 124 公会 玩家等级不足
	GuildWordSensitive                                 // 125 公会 内容包含敏感词
	AddItemFail_MaxCount                               // 126 物品 加物品是超过物品上限
	GuildGateEnemyStartErr                             // 127 公会开启兵临城下活动失败
	GuildGateEnemyHasStarted                           // 128 公会开启兵临城下活动失败:活动已经开启
	GuildGateEnemyRESWarnNoAct                         // 129 兵临城下活动没有开启
	GuildGateEnemyRESWarnCanNotInAct                   // 130 兵临城下活动玩家不能参加活动
	GuildGateEnemyRESWarnEnemyIDErr                    // 131 兵临城下活动杂兵ID错误
	GuildGateEnemyRESWarnEnemyHasFighting              // 132 兵临城下活动杂兵已经被人挑战
	GuildGateEnemyRESErrNoBoss                         // 133 兵临城下活动BossID错误
	GuildGateEnemyRESErrStateErr                       // 134 兵临城下活动玩家状态错误
	GuildGateEnemyRESTimeOut                           // 135 兵临城下活动请求超时
	MailReceiveMailIDErr                               // 136 要领取的邮件ID不存在
	GuildHasOut                                        // 137 主公，您已被踢出公会 募捐时被踢出公会
	GuildHasIn                                         // 138 主公，您已加入公会了 申请时已经加入工会了
	GuildGatesEnemyGetRewardErrByNoGuild               // 139 玩家不在公会中, 无法领取公会奖励
	GuildGatesEnemyGetRewardErrByNoReward              // 140 没有奖励可以领取
	QuestErrIDX                                        // 141 任务数据出错
	QuestErrIDXData                                    // 142 任务数据出错
	QuestErrUnFinishCond                               // 143 不满足接取条件 (多是跨天导致的)
	QuestErrFinish                                     // 144 未知的完成失败错误
	QuestErrGiveData                                   // 145 任务数据出错
	QuestErrGive                                       // 146 任务数据出错
	QuestErrBugFull                                    // 147 背包已满
	GiftByTimeErrNoGift                                // 148 没有在线礼包
	GiftByTimeErrNoCond                                // 149 不满足在线礼包领取条件
	GiftByTimeErrGive                                  // 150 在线礼包奖励错误
	PhoneRegCannotGetCodeHasGotReward                  // 151 当前无法获得验证码,已经领取过奖励
	PhoneRegCannotGetCodeTooFast                       // 152 当前无法获得验证码,玩家账号请求次数过快
	PhoneRegCannotGetCodeTooMuch                       // 153 当前无法获得验证码,玩家账号请求次数过多
	PhoneRegCannotGetCodeByPhone                       // 154 该手机号当天获取验证码次数过多
	PhoneRegPhoneFormatErr                             // 155 手机号格式错误
	PhoneRegUnknownErr                                 // 156 请求错误
	PhoneRegRegCodeErr                                 // 157 验证码输入错误
	Account7DayQuestTimeOut                            // 158 新角色7天活动，任务时间已过
	Account7DaySevGoodCountNotEnough                   // 159 新角色7天活动，全服物品数量不足
	Account7DayQuestTimeNotYet                         // 160 新角色7天活动, 任务时间未到
	CurrNoGVE                                          // 161 当前不在gve活动时间
	GuildApplyGsNotEnough                              // 162 公会 公会申请gs不满足
	GuildIndexIllegal                                  // 163 公会 显示id格式非法
	ActivityTimeOut                                    // 164 活动 未开启
	TitleCondNotCond                                   // 165 称号 条件不满足
	GuildInventoryFull                                 // 166 公会仓库 已满,加物品失败
	GuildBossParamErr                                  // 167 公会Boss 未知错误
	GuildBossNoFound                                   // 168 公会Boss 没有找到Boss
	GuildBossCurrBossHasLocker                         // 169 公会Boss 当前Boss已经锁定
	GuildBossCurrBossNoLocker                          // 170 公会Boss 当前Boss没有被锁定
	GuildBossCount                                     // 171 公会Boss 没有次数
	GateEnemyBuffAlready                               // 172 兵临城下 buff已都被占用
	GuildScienceLevelFull                              // 173 工会科技 已满级
	FenghuoRoomNumWarn                                 // 174 烽火连城 传入参数房间号码不正确
	FenghuoAlreadyInRoom                               // 175
	FenghuoRoomFull                                    // 176
	FenghuoNotJoinable                                 // 177
	FenghuoWaitOthersReady                             // 178
	FenghuoNoMoney                                     // 179
	HeroGachaRaceTimeOut                               // 180 限时神将，活动已过期
	SimplePvpMatchOutOfTime                            // 181 1v1竞技场对手匹配超时
	HeroGachaRaceNotReady                              // 182 限时神将 活动未准备好
	HeroGachaRaceNotStart                              // 183 限时神将 活动未开启
	HeroGachaRaceNoRank                                // 184 限时神将 未上榜
	ClientSync                                         // 185 客户端会getinfo
	MaterialNotEnough                                  // 186 材料不足
	ClickTooQuickly                                    // 187 点击过快了
	ActivityNotValid                                   // 188 (活动，关卡)当前不可用
	GateEnemyActivityOver                              // 189 兵临城下活动，马上就要结束了，不可进了
	NotSupportedThisVersionPlsUpgrade                  // 190 IDS_ERROR_NETWORK_190 服务器跨服使用Warning
	TPVPEnemyLocked                                    // 191 3v3竞技场所挑战的角色被其他人锁定了
	TPVPInvalidFight                                   // 192 3v3竞技场由于挑战时间超时,而结果无效
	ChangeGuildInSameDay                               // 193 您在同一天内更换了军团\n当天不可参与军团活动
	LimitGoodTimeOut                                   // 194 限时礼包过期
	HeroCompanionNotOpen                               // 195 英雄聚义未开启
	HeroCompanionNotAllActive                          // 196 进化时, 聚义武将没有全部激活
	HeroCompanionEvolveMaxLevel                        // 197 进化等级已经达到最大等级
	YouCheat                                           // 198 战斗作弊警告
	YouTimeCheat                                       // 199 战斗通关时间超时cheat

//XXX: 小心,如果要超过200请来找Yin Ze Hong
)

var (
	Warn_Str = []string{
		none: "none",
		RecodeGiftCodeTimeout:            "RecodeGiftCodeTimeout",
		RecodeGiftCodeTimeNoStart:        "RecodeGiftCodeTimeNoStart",
		RecodeGiftCodeUsed:               "RecodeGiftCodeUsed",
		RecodeGiftCodeBatchHasExchange:   "RecodeGiftCodeBatchHasExchange",
		RecodeGiftCodeFormatErr:          "RecodeGiftCodeFormatErr",
		RecodeGiftCodeDataErr:            "RecodeGiftCodeDataErr",
		RenameSensitve:                   "RenameSensitve",
		RenameNameHasExit:                "RenameNameHasExit",
		ChatEnterWarn:                    "ChatEnterWarn",
		RecodeGiftCodeBindErr:            "RecodeGiftCodeBindErr",
		AndroidPayOrderTryTimeOut:        "AndroidPayOrderTryTimeOut",
		GuildAlreadyInGuild:              "GuildAlreadyInGuild",
		GuildNameRepeat:                  "GuildNameRepeat",
		GuildNotFound:                    "GuildNotFound",
		GuildFull:                        "GuildFull",
		GuildApplyFull:                   "GuildApplyFull",
		GuildPositionErr:                 "GuildPositionErr",
		GuildPlayerAlreadyInOther:        "GuildPlayerAlreadyInOther",
		GuildPlayerNotFound:              "GuildPlayerNotFound",
		GuildApplicantNotFound:           "GuildApplicantNotFound",
		GuildChiefNotQuit:                "GuildChiefNotQuit",
		GuildPlayerNotIn:                 "GuildPlayerNotIn",
		GuildWordIllegal:                 "GuildWordIllegal",
		GuildMemLevelNotEnough:           "GuildMemLevelNotEnough",
		GuildWordSensitive:               "GuildWordSensitive",
		AddItemFail_MaxCount:             "AddItemFail_MaxCount",
		Account7DaySevGoodCountNotEnough: "Account7DaySevGoodCountNotEnough",
		CurrNoGVE:                        "CurrNoGVE",
	}
)
