package ban_account

import (
	"errors"
	"fmt"
	"strings"
	//"strconv"
	"crypto/sha1"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/astaxie/beego/httplib"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	gmConfig "vcs.taiyouxi.net/platform/x/gm_tools/config"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/rank_del"
	"vcs.taiyouxi.net/platform/x/gm_tools/util"
)

type BatchDealResult struct {
	SuccessCount int
	FailArr      []SingleBatchDeal
}

type SingleBatchDeal struct {
	Id       int
	FailCode int
}
type BatchBanSuccessInfo struct {
	InfoList []SuccessBanInfo
}

type SuccessBanInfo struct {
	CsvIndex int
	Uid      string
	Idx      int64
}
type BanAccountInfo struct {
	csvIndex    int
	serverName  string
	profileAcid string
	mailReward  timail.MailReward
}

// 单行数据
const (
	_                          = iota
	CSV_MAIL_ERROR_TIME_FORMAT //
	CSV_MAIL_ERROR_ITEM_FORMAT //
	CSV_MAIL_ERROR_ID          //
	CSV_MAIL_ERROR_SERVER_NAME
	CSV_MAIL_ERROR_SEND
	CSV_MAIL_ERROR_ACID
	BanTime = "31536000"
)

func RegCommands() {
	gm_command.AddGmCommandHandle("banAccount", banAccount)
	gm_command.AddGmCommandHandle("getAllBanAccount", getAllBanAccount)
}

func BatchBanAccounts(csvData [][]string) (BatchDealResult, error) {
	lock := sync.RWMutex{}
	lock.Lock()
	ret := BatchDealResult{FailArr: make([]SingleBatchDeal, 0, len(csvData))}
	lock.Unlock()
	for i, csvData := range csvData {
		logs.Trace("新的csvData为：%v", csvData)
		var uid string
		if strings.HasPrefix(csvData[1], "profile:") {
			uid = csvData[1][8:]
		} else {
			uid = csvData[1]
		}

		acid, err := db.ParseAccount(uid)

		if err != nil {
			ret.FailArr = append(ret.FailArr, SingleBatchDeal{Id: i, FailCode: CSV_MAIL_ERROR_SERVER_NAME})
			continue
		}
		server := fmt.Sprintf("%d", acid.GameId) + ":" + fmt.Sprintf("%d", acid.ShardId)
		err = banAccount(nil, server, uid, []string{BanTime, "System"})
		if err != nil {
			logs.Error("封禁玩家acid:%v时出错，错误信息为:%v", uid, err)
			ret.FailArr = append(ret.FailArr, SingleBatchDeal{Id: i, FailCode: CSV_MAIL_ERROR_ACID})
			continue
		}
		err = rank_del.DelPlayerRanks(&gm_command.Context{}, server, uid)
		if err != nil {
			logs.Error("删除玩家acid:%v榜单时出错，错误信息为:%v", uid, err)
			ret.FailArr = append(ret.FailArr, SingleBatchDeal{Id: i, FailCode: CSV_MAIL_ERROR_ACID})
			continue
		}
		logs.Trace("玩家:%v封禁删榜成功", uid)
		ret.SuccessCount = len(csvData) - len(ret.FailArr)
	}
	ret.SuccessCount = len(csvData) - len(ret.FailArr)
	return ret, nil
}

func banAccount(c *gm_command.Context, server, accountid string, params []string) error {
	cfg := gmConfig.Cfg.GetServerCfgFromName(server)
	if cfg == (gmConfig.ServerCfg{}) {
		return errors.New("No server: " + server)
	}
	logs.Info("banAccount  %s %s %v--> %v", accountid, server, params, cfg)
	acids := strings.Split(accountid, ":")
	if len(acids) < 3 {
		return errors.New("acidErr")
	}
	gid := acids[0]
	uid := acids[2]
	reason := "IDS_ACCOUNT_CLOSURE"
	if len(params) >= 3 {
		reason = params[2]
	}
	url := fmt.Sprintf("%s%s?time=%s&gid=%s&reason=%s", cfg.AuthApi, uid, params[0], gid, reason)

	logs.Info(url)
	_, err := util.HttpGet(url)
	if err != nil {
		return err
	}

	if gid == "205" { // 港澳台
		if params[0] != "1" { // 封禁
			if err := MuteHMT(true, accountid); err != nil {
				return err
			}
		} else { // 解禁
			if err := MuteHMT(false, accountid); err != nil {
				return err
			}
		}
	}

	return nil
}

func getAllBanAccount(c *gm_command.Context, server, accountid string, params []string) error {
	return nil
}

const (
	AppKey    = "YOUME57CE6A710CBA82B140C189D58B1745115C002AA4"
	AppSecret = "5ph4gaF/LjA7Lohmn91TAUOMCUUW8jjprOvlkCe3ez0VHtVRs5syzGnMHqlEsN5r/F/2UC4gPheOGqVqAvGpmo8MIAlRCTIEqkI5LCLbBG8c5jm4pCAaZtMCqeUe+KgrBDu0JfhlCDPCjIK38vzsGEz8V4aIVCnIdQTwdJlKgUsBAAE="
)

func MuteHMTSec(uid string, muteTime int64) error {
	bbdy, err := genMuteSecBody(uid, muteTime)
	if err != nil {
		return err
	}
	err = muteHMT(bbdy)
	return err

}

func MuteHMT(isMute bool, uid string) error {
	bbdy, err := genMuteBody(isMute, uid)
	if err != nil {
		return err
	}
	err = muteHMT(bbdy)
	return err
}

func muteHMT(bbdy []byte) error {
	now_t := time.Now().Unix()
	data := AppSecret + fmt.Sprint(now_t)
	t := sha1.New()
	io.WriteString(t, data)
	checksum := fmt.Sprintf("%x", t.Sum(nil))

	url := "https://sgapi.youme.im/v1/im/forbid_im_user_send_channel_msg?" + "appkey=" + AppKey + "&identifier=wangliang@taiyouxi.cn&curtime=" + fmt.Sprintf("%d", now_t) + "&checksum=" + checksum
	req := httplib.Post(url).SetTimeout(5*time.Second, 5*time.Second)

	req.Header("Content-Type", "application/json")

	logs.Debug("MuteHMT %s", string(bbdy))
	req.Body(bbdy)

	var r ret
	err := req.ToJson(&r)
	if err != nil {
		return err
	}
	if r.ActionStatus != "OK" {
		return fmt.Errorf("%v", r)
	}
	logs.Warn("MuteHMT success isMute %s", string(bbdy))
	return nil
}

func genMuteBody(mute bool, uid string) ([]byte, error) {
	uid = strings.Replace(uid, ":", "", -1)
	uid = strings.Replace(uid, "-", "", -1)
	if mute {
		bdy := body_ban{

			UserList: []user{
				{uid},
			},
		}
		bbdy, err := json.Marshal(bdy)
		if err != nil {
			return nil, err
		}
		return bbdy, nil
	} else {
		bdy := body_free{
			ShutUpTime: 0,
			UserList: []user{
				{uid},
			},
		}
		bbdy, err := json.Marshal(bdy)
		if err != nil {
			return nil, err
		}
		return bbdy, nil
	}
}

// -1 represent infinity max
func genMuteSecBody(uid string, time int64) ([]byte, error) {
	uid = strings.Replace(uid, ":", "", -1)
	uid = strings.Replace(uid, "-", "", -1)
	var bdy interface{}
	if time == -1 {
		bdy = body_ban{
			UserList: []user{
				{uid},
			},
		}
	} else {
		bdy = body_ban_sec{
			ShutUpTime: time,
			UserList: []user{
				{uid},
			},
		}
	}
	bbdy, err := json.Marshal(bdy)
	if err != nil {
		return nil, err
	}
	return bbdy, nil
}

type body_free struct {
	ShutUpTime int
	UserList   []user
}
type body_ban struct {
	UserList []user
}

type body_ban_sec struct {
	UserList   []user
	ShutUpTime int64
}

type user struct {
	UserID string
}

type ret struct {
	ActionStatus string
	ErrorCode    int
	ErrorInfo    string
}
