package mail

import (
	//"errors"
	//"time"
	"encoding/json"
	"strconv"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/global_mail"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/x/api_gateway/pay"
)

const (
	MailStateIdle       = timail.MailStateIdle
	MailStateSyncUpdate = timail.MailStateSyncUpdate //玩家在线新生成的邮件==SyncToDynamoDB
	MailStateDel        = timail.MailStateDel        //需要删除DelOnDynamoDB
)

func mkRewardFromMail(m *timail.MailReward) (*gamedata.CostData, uint32) {
	r := &gamedata.CostData{}
	money := beforeReward(m, r)
	for i := 0; i < len(m.ItemId) && i < len(m.Count); i++ {
		r.AddItem(m.ItemId[i], m.Count[i])
	}
	return r, money
}

func beforeReward(m *timail.MailReward, r *gamedata.CostData) (money uint32) {
	var order timail.IAPOrder
	if m.Tag != "" {
		if err := json.Unmarshal([]byte(m.Tag), &order); err != nil {
			//logs.Error("beforeReward mail tag cant unmarsharl %v, %s", m.Tag, err.Error())
			return
		}
	} else {
		return
	}

	// 检查游戏内购订单和金钱量匹配
	idx, err := strconv.Atoi(order.Game_order)
	if err != nil {
		//logs.Error("mail android iap gameorder not int: %v", order)
		return
	}
	amount, err := strconv.ParseFloat(order.Amount, 64)
	if err != nil {
		logs.Error("mail android iap amount not float: %v", order)
		return
	}
	if order.Channel == pay.KOONESTOREChannel || order.Channel == pay.KOIOSChannel || order.Channel == pay.KOGPAndroidChannel {
		switch timail.GetMailSendTyp(m.Idx) {
		case timail.Mail_Send_By_AndroidIAP:
			r.AddIAPGood(uint32(idx), order.Game_order_id, order.Order_no, uint32(amount),
				order.Channel, order.PayTime, logiclog.Android_Platform, uint32(amount),
				gamedata.PackageInfo{order.PkgInfo.PkgId, order.PkgInfo.SubPkgId}, order.PayType)
		case timail.Mail_Send_By_IOSIAP:
			r.AddIAPGood(uint32(idx), order.Game_order_id, order.Order_no, uint32(amount),
				order.Channel, order.PayTime, logiclog.IOS_Platform, uint32(amount),
				gamedata.PackageInfo{order.PkgInfo.PkgId, order.PkgInfo.SubPkgId}, order.PayType)
		}
		return
	}
	info := gamedata.GetIAPInfo(uint32(idx))
	if info == nil {
		logs.Warn("may be vn pay for idx: %v", idx)
		info = &gamedata.IAPInfo{}
	}
	switch timail.GetMailSendTyp(m.Idx) {
	case timail.Mail_Send_By_AndroidIAP:
		if order.Channel != pay.VNAndroidChannel && order.Channel != pay.VNIOSChannel {
			if info.Android_Rmb_Price != uint32(amount) {
				logs.Error("mail android iap amount not equip config price: %v", order)
				return
			}
		}

		r.AddIAPGood(uint32(idx), order.Game_order_id, order.Order_no, info.Android_Rmb_Price,
			order.Channel, order.PayTime, uutil.Android_Platform, uint32(amount),
			gamedata.PackageInfo{order.PkgInfo.PkgId, order.PkgInfo.SubPkgId}, order.PayType)

	case timail.Mail_Send_By_IOSIAP:
		if order.Channel != pay.VNAndroidChannel && order.Channel != pay.VNIOSChannel {
			if info.IOS_Rmb_Price != uint32(amount) {
				logs.Error("mail ios iap amount not equip config price: %v", order)
				return
			}
		}
		r.AddIAPGood(uint32(idx), order.Game_order_id, order.Order_no, info.IOS_Rmb_Price,
			order.Channel, order.PayTime, logiclog.IOS_Platform, uint32(amount),
			gamedata.PackageInfo{order.PkgInfo.PkgId, order.PkgInfo.SubPkgId}, order.PayType)

	}
	return uint32(amount)
}

//PlayerMail 玩家的所有邮件
type PlayerMail struct {
	mail_reward map[int64]timail.MailReward
	user_id     string

	tmp_mail int64

	ignore_idx []int64
}

func (m *PlayerMail) OnAfterLogin(user_id string) {
	m.user_id = user_id
	m.mail_reward = make(map[int64]timail.MailReward, 32)

	/*
		login之后不主动请求
			err := m.LoadMail()
			if err != nil {
				logs.SentryLogicCritical(user_id, "LoadMail Err by %s", err.Error())
			}
	*/
}

func (m *PlayerMail) LoadMail(server_name, accountID string, createTime int64) error {
	//刷新全服邮件, server_name  should be like 0:0
	all_mails := global_mail.GetGlobalMail(accountID, createTime)

	mails, err := db.LoadAllMail(m.user_id)
	if err != nil {
		return err
	}

	// 先清除掉不是新增的邮件
	mail_reward_new := make(map[int64]timail.MailReward, 32)
	for idx, mail := range m.mail_reward {
		if mail.GetState() == MailStateSyncUpdate {
			mail_reward_new[idx] = mail
		}
	}
	m.mail_reward = mail_reward_new

	for _, mail_from_db := range mails {
		logs.Trace("mail_from_db %v", mail_from_db)
		// by zhangzhen 去掉功能：邮件在玩家创建账号之前发的话，玩家就看不到
		//if createTime < mail_from_db.TimeBegin {
		m.mail_reward[mail_from_db.Idx] = mail_from_db
		//}
	}

	logs.Trace("LoadAllMail From all %v ", all_mails)

	for _, mail_from_db := range all_mails {
		idx := mail_from_db.Idx
		logs.Trace("mail_from_all %v", mail_from_db)
		//检查全服邮件在玩家的数据库中的状态
		isExist, err := db.MailExist(m.user_id, idx)
		if err != nil {
			logs.Error("Player loads global mails, but check in his db failed. %s", err.Error())
			//数据库出现任何异常玩家,则不应该可以领取全服邮件
			continue
		}
		if isExist {
			//如果数据库中已经存在,则玩家不应该可以领取全服邮件
			continue
		}
		// by zhangzhen 去掉功能：邮件在玩家创建账号之前发的话，玩家就看不到
		//if createTime > mail_from_db.TimeBegin {
		//	continue
		//}

		_, ok := m.mail_reward[idx]

		if !ok {
			mail_from_db.SetState(MailStateIdle)
			m.mail_reward[idx] = mail_from_db
		}
	}

	// 军团仓库改版4 需要删除过滤
	////清除失效的MailIdx
	//ignores := []int64{}
	//for idx, mail := range m.mail_reward {
	//	if checkIgnoreMail(mail.IdsID) {
	//		logs.Trace("LoadMail ignore, ids [%d], mail [%v]", mail.IdsID, mail)
	//		ignores = append(ignores, idx)
	//		m.ignore_idx = append(m.ignore_idx, mail.Idx)
	//	}
	//}
	//for _, idx := range ignores {
	//	delete(m.mail_reward, idx)
	//}

	return nil
}

func (m *PlayerMail) ReadMail(idx int64) {
	mail, ok := m.mail_reward[idx]
	if ok {
		mail.IsRead = true
		mail.SetState(MailStateSyncUpdate)
		m.mail_reward[idx] = mail
	}
}

func (m *PlayerMail) ReceiveMail(idx int64) error {
	mail, ok := m.mail_reward[idx]
	if !ok {
		return nil
	}
	logs.Trace("IsNeedNoDelBeforeGetted %v", mail)
	if mail.IsNeedNoDelBeforeGetted() {
		mail.IsGetted = true
		mail.SetState(MailStateSyncUpdate) // 需要更新数据库
		m.mail_reward[idx] = mail

		logs.Trace("IsNeedNoDelBeforeGetted %v", mail)
	} else {
		if mail.GetState() == MailStateSyncUpdate {
			delete(m.mail_reward, idx)
		} else {
			mail.SetState(MailStateDel)
			m.mail_reward[idx] = mail
		}
	}
	return nil
}

func (m *PlayerMail) GetMailReward(idx int64) (*gamedata.CostData, uint32, string, string, string) {
	mail, ok := m.mail_reward[idx]
	if !ok || mail.GetState() == MailStateDel || mail.IsGetted {
		return nil, 0, "", "", ""
	}

	var verTag timail.VerTag
	if mail.Tag != "" {
		if err := json.Unmarshal([]byte(mail.Tag), &verTag); err != nil {
			logs.Error("beforeReward mail tag cant unmarsharl %s, %v, %v, %s",
				m.user_id, mail.Tag, mail, err.Error())
			return nil, 0, "", "", ""
		}

	}
	cd, money := mkRewardFromMail(&mail)
	return cd, money, mail.Reason, verTag.Ver, verTag.Ch
}

func (m *PlayerMail) MkMailId() int64 {
	res := timail.MkMailId(timail.Mail_Send_By_TMP, m.tmp_mail)
	m.tmp_mail++
	// 注意这里假定同一秒内不会发超过1000封邮件
	if m.tmp_mail >= timail.Mail_Id_Gen_Base {
		m.tmp_mail = 0
	}

	return res
}

func (m *PlayerMail) SendMail(ids uint32, param []string, reason string,
	item_id string, count uint32) bool {
	// 在内存中发, 在离线时发给Dynamo
	mail := timail.MailReward{
		Idx:    m.MkMailId(),
		IdsID:  ids,
		Param:  param,
		Reason: reason,
	}
	mail.SetState(MailStateSyncUpdate)

	mail.AddReward(item_id, count)
	m.mail_reward[mail.Idx] = mail
	return true
}

func (m *PlayerMail) SendMailWithRewards(ids uint32, param []string, reason string,
	item_id []string, count []uint32) bool {
	// 在内存中发, 在离线时发给Dynamo
	mail := timail.MailReward{
		Idx:    m.MkMailId(),
		IdsID:  ids,
		Param:  param,
		Reason: reason,
	}
	mail.SetState(MailStateSyncUpdate)
	for idx, itemid := range item_id {
		mail.AddReward(itemid, count[idx])
	}
	logs.Trace("SendMailWithRewards %v --> %v", mail.Idx, mail)
	m.mail_reward[mail.Idx] = mail
	return true
}

func (m *PlayerMail) SyncMailToDynamo() error {
	logs.Trace("SyncMailToDynamo start")
	to_del := make([]int64, 0, len(m.mail_reward))
	mails := make([]timail.MailReward, 0, len(m.mail_reward))
	for _, mail := range m.mail_reward {
		if mail.GetState() == MailStateSyncUpdate {
			mails = append(mails, mail)
		} else if mail.GetState() == MailStateDel {
			to_del = append(to_del, mail.Idx)
		}
	}

	//del ignored mail
	for _, ignoreIdx := range m.ignore_idx {
		to_del = append(to_del, ignoreIdx)
	}

	err := db.SyncMail(m.user_id, mails, to_del)
	if err != nil {
		logs.Debug("SyncMailToDynamo Err %v %v", mails, to_del)
		logs.Error("SyncMailToDynamo Err %s by %s", m.user_id, err.Error())
	} else {
		// 只有成功才变更状态,防止出错
		for idx, mail := range m.mail_reward {
			if mail.GetState() == MailStateSyncUpdate {
				mail.SetState(MailStateIdle)
				m.mail_reward[idx] = mail
			} else if mail.GetState() == MailStateDel {
				delete(m.mail_reward, mail.Idx)
			}
		}

		m.ignore_idx = []int64{}
	}
	logs.Trace("SyncMailToDynamo end")
	return err
}

func (m *PlayerMail) GetAllMail() []timail.MailReward {
	mails := make([]timail.MailReward, 0, len(m.mail_reward))

	for _, mail := range m.mail_reward {
		if mail.IsAvailable() {
			mails = append(mails, mail)
		}
	}

	return mails[:]
}

func (m *PlayerMail) CheckErrorMail() {
	for key, mail := range m.mail_reward {
		var verTag timail.VerTag
		if mail.Tag != "" {
			if err := json.Unmarshal([]byte(mail.Tag), &verTag); err != nil {
				logs.Warn("<Check Mail> mail tag cant unmarsharl %s, %v, %v, %s",
					m.user_id, mail.Tag, mail, err.Error())
				mail.SetState(MailStateDel)
				m.mail_reward[key] = mail
			}
		}
	}
}

//func checkIgnoreMail(ids uint32) bool {
//	return ignoreMailList[ids]
//}
