package timail

import "time"

const (
	//玩家能够加载的最多邮件数
	MAX_Mail_ONE_GET = 400
	//玩家能够回写的最多邮件数
	MAX_Mail_ONE_SEND = 25
	//全服邮件刷新周期 以秒为单位
	Mail_All_Update_Second = 30
)

const (
	Mail_Id_Gen_Base int64 = 1000
	//Reserved Range 用于防止不同系统在同一个时间返送邮件导致range相同
	//作为秒级时间戳的后缀存在
	mail_Reserved_Range int64 = Mail_Id_Gen_Base * 100 // 留两位
)

// 9223372036854775807
// 14406711113016

const (
	//系统发送的全服邮件, Range范围是1000-1999，假定同一秒内不会发送超过999封邮件
	//留了99种类型应该够用了
	_ int64 = iota * Mail_Id_Gen_Base
	Mail_Send_By_Sys
	Mail_Send_By_GM
	Mail_Send_By_Rank_SimplePvp
	Mail_Send_By_Boss
	Mail_Send_By_AndroidIAP
	Mail_Send_By_Single_Player
	Mail_Send_By_TMP
	Mail_Send_By_Rank_ServerOpn_Player_Gs
	Mail_Send_By_Rank_ServerOpn_Player_Rank
	Mail_Send_By_Rank_ServerOpn_Guild // 10
	Mail_Send_By_Rank_ServerOpn_Guild_leader
	Mail_Send_By_TeamPvp
	Mail_Send_By_IOSIAP
	Mail_Send_By_Guild_Inventory
	Mail_Send_By_Pay_FeedBack
	Mail_Send_By_Market_Activity
	Mail_Send_By_Hero_GachaRace
	Mail_Send_By_WeekRank_SimpePvp
	Mail_Send_By_Auto_Change_Chief
	Mail_Send_By_GVG // 20
	Mail_send_By_Guild
	Mail_send_By_RedPacket7Days  // 22
	Mail_send_By_GuildBoss_Death // 23
	Mail_send_By_CSRob           // 24
	Mail_send_By_Black_Gacha     // 25
	Mail_send_By_Debug           // 26
	Mail_send_By_Common          // 27
	Mail_Sender_Count
)

func GetMailSendTyp(mid int64) int64 {
	reserved := mid % mail_Reserved_Range
	return (reserved / Mail_Id_Gen_Base) * Mail_Id_Gen_Base
}

func MkMailIdByTime(time_now, mail_t, addon int64) int64 {
	return time_now*mail_Reserved_Range + mail_t + addon
}

func MkMailId(mail_t, addon int64) int64 {
	time_now := time.Now().Unix()
	return time_now*mail_Reserved_Range + mail_t + addon
}

// 从Mail id 获取Mail id 生成的时间
func GetTimeFromMailID(id int64) int64 {
	return id / mail_Reserved_Range
}

type MailAddonCounter struct {
	i int64
}

var mailAddonMap map[int64]MailAddonCounter

// 注意，要在多线程下使用，自己加锁
func (m *MailAddonCounter) Get() int64 {
	if m.i >= Mail_Id_Gen_Base-1 {
		m.i = 0
	}
	m.i++
	return m.i
}

func MkMailIdAuto(typ int64, timeNow int64) int64 {
	if mailAddonMap == nil {
		mailAddonMap = make(map[int64]MailAddonCounter, Mail_Sender_Count)
	}
	_, ok := mailAddonMap[typ]
	if !ok {
		mailAddonMap[typ] = MailAddonCounter{}
	}
	m, ok := mailAddonMap[typ]

	if m.i >= Mail_Id_Gen_Base-1 {
		m.i = 0
	}
	m.i++
	mailAddonMap[typ] = m

	return timeNow*mail_Reserved_Range + typ + m.i
}
