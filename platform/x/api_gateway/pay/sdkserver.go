package pay

import (
	"fmt"

	"strconv"

	"strings"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/tipay"
	"vcs.taiyouxi.net/platform/planx/util/tipay/pay"
	"vcs.taiyouxi.net/platform/x/api_gateway/config"
)

const (
	MODE_PROD     = "PROD"
	MODE_QA       = "QADEV"
	MODE_PROD_203 = "PRODTEMP"

	typ_quick       = "QUICK"
	typ_quick_mubao = "QUICKMUBAO"
	typ_vivo        = "VIVO"
	typ_samsung     = "SAMSUNG"
	typ_enjoy       = "ENJOY"
	typ_vn          = "VN"
	typ_enjoyko     = "ENJOYKO"
)

var (
	sdkNotifyMap map[string]SdkNotifyFunc
	sdkInitMap   map[string]SdkInitFunc
)

func init() {
	sdkInitMap = map[string]SdkInitFunc{
		typ_samsung: SamsungInit,
	}
	sdkNotifyMap = map[string]SdkNotifyFunc{
		typ_quick:       QuickPayNotify,
		typ_quick_mubao: QuickPayNotify,
		typ_vivo:        VivoPayNotify,
		typ_samsung:     SamsungPayNotify,
		typ_enjoy:       EnjoyPayNotify,
		typ_vn:          VNPayNotify,
		typ_enjoyko:     EnjoyKoPayNotify,
	}
}

type PayWorker struct {
	DBPayRelease pay.PayDB
	DBPayDebug   pay.PayDB

	MailRelease *MailSenderModule
	MailDebug   *MailSenderModule
	// TODO by ljz tmp revise
	MailTemp *MailSenderModule
}

type sdkPayData struct {
	config.PayService
	worker PayWorker
}

func (pd *sdkPayData) IsIOS() bool {
	if pd.Name == "ios" {
		return true
	}
	return false
}

func (pd *sdkPayData) PayDBGetByHashM(isDebug bool, any interface{}) (map[string]interface{}, error) {
	if isDebug {
		return pd.worker.DBPayDebug.GetByHashM(any)
	} else {
		return pd.worker.DBPayRelease.GetByHashM(any)
	}
}

func (pd *sdkPayData) PaySetByHashM(isDebug bool, hash interface{}, values map[string]interface{}) error {
	if isDebug {
		return pd.worker.DBPayDebug.SetByHashM_IfNoExist(hash, values)
	} else {
		return pd.worker.DBPayRelease.SetByHashM_IfNoExist(hash, values)
	}
}

func (pd *sdkPayData) PayUpdateByHash(isDebug bool, hash interface{}, values map[string]interface{}) error {
	if isDebug {
		return pd.worker.DBPayDebug.UpdateByHash(hash, values)
	} else {
		return pd.worker.DBPayRelease.UpdateByHash(hash, values)
	}
}

func (pd *sdkPayData) SendMail(isDebug bool, accountId, info string) error {
	var msm *MailSenderModule
	if isDebug {
		msm = pd.worker.MailDebug
	} else {
		msm = pd.worker.MailRelease
	}

	if msm == nil {
		return fmt.Errorf("SendMail in Debug(%v), not found Mail Dynamo Service. %s, %s",
			isDebug, accountId, info)
	}

	switch pd.Name {
	case "ios":
		return msm.SendIOSIAPMail(accountId, info)
	default:
		return msm.SendAndroidIAPMail(accountId, info)
	}
	return nil
}

// TODO by ljz tmp revise
func (pd *sdkPayData) SendMailTemp(accountId, info string) error {
	var msm *MailSenderModule
	msm = pd.worker.MailTemp
	switch pd.Name {
	case "ios":
		return msm.SendIOSIAPMail(accountId, info)
	default:
		logs.Debug("send mail to accountId: %s, info: %v", accountId, info)
		return msm.SendAndroidIAPMail(accountId, info)
	}
	return nil
}

func RegSdk(g *gin.Engine, Starter func(func()), Stoper func(func())) error {
	for _, service := range config.PayServicesCfg.Services {
		var payWorker PayWorker
		var err error
		if dc, ok := service.Pay[MODE_QA]; ok {
			payWorker.DBPayDebug, err = tipay.NewPayDriver(dc.GetPayDBConfig())
			if err != nil {
				return err
			}
		}
		if dc, ok := service.Pay[MODE_PROD]; ok {
			payWorker.DBPayRelease, err = tipay.NewPayDriver(dc.GetPayDBConfig())
			if err != nil {
				return err
			}
		}

		if dc, ok := service.Mail[MODE_QA]; ok {
			msm := &MailSenderModule{}
			ndc := dc

			if err := msm.Start(ndc); err != nil {
				return err
			}
			payWorker.MailDebug = msm
			//Starter(func() { msm.Start(ndc) })
			//Stoper(func() { msm.Stop() })
		}

		if dc, ok := service.Mail[MODE_PROD]; ok {
			msm := &MailSenderModule{}
			ndc := dc

			if err := msm.Start(ndc); err != nil {
				return err
			}
			payWorker.MailRelease = msm
			//Starter(func() { msm.Start(ndc) })
			//Stoper(func() { msm.Stop() })
		}
		// TODO by ljz tmp revise
		if dc, ok := service.Mail[MODE_PROD_203]; ok {
			if dc.DBName != "" {
				msm := &MailSenderModule{}
				ndc := dc

				if err := msm.Start(ndc); err != nil {
					return err
				}
				payWorker.MailTemp = msm
				logs.Debug("temp mail driver: %v", *(payWorker.MailTemp))
			}
		}

		for k, v := range service.Sdk {
			init, ok := sdkInitMap[k]
			if ok {
				init(sdkPayData{
					PayService: service,
					worker:     payWorker,
				})
			}
			f, ok := sdkNotifyMap[k]
			if !ok {
				return fmt.Errorf("config sdk func not found, %s", k)
			}
			logs.Debug("start listen url %v", v.SdkUrlRelPath)
			g.POST(v.SdkUrlRelPath,
				f(k, sdkPayData{
					PayService: service,
					worker:     payWorker,
				}),
			)
		}

	}
	g.GET("/Health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	return nil
	//g.POST(config.AndroidCfg.SdkUrlRelPath, sdkPayNotify(false))
	//g.POST(config.IOSCfg.SdkUrlRelPath, sdkPayNotify(true))
}

type SdkNotifyFunc func(typ string, data sdkPayData) gin.HandlerFunc
type SdkInitFunc func(data sdkPayData)

func checkOrderNoRepeat(typ string, data sdkPayData, isdbDebug bool, hash_key string) (success bool, errMsg string, isRepeat bool) {
	// 判断重复
	oldValue, err := data.PayDBGetByHashM(isdbDebug, hash_key)
	if err != nil {
		logs.Warn("[%s] PayNotify GetByHashM db err %s %v %s", typ, isdbDebug, err.Error())
		return false, "DB Failed", false
	}
	if oldValue != nil && len(oldValue) > 0 {
		ret := true
		if v, ok := oldValue[key_tistatus]; ok {
			str, ok := v.(string)
			if ok {
				if str != tistatus_delivered {
					//还没有成功发送到Mail,或者上次发送Mail失败了
					//goon
					ret = false
				}
			}
		}
		if ret {
			logs.Warn("[%s] PayNotify order_no repeat %s", typ, hash_key)
			return true, "", true
		}
	}
	return true, "", false
}

// 比较钱和订单需要的钱是否一致
func checkMoneyWithCfg(typ string, data sdkPayData, game_order string, amount uint32, uid string) (bool, string) {
	idx, err := strconv.Atoi(game_order)
	if err != nil {
		logs.Error("[%s] PayNotify Game_order err %s", typ, game_order)
		return false, "GameOrder Param Err"
	}

	cfgInfo := gamedata.GetIAPInfo(uint32(idx))
	if cfgInfo == nil {
		logs.Error("[%s] PayNotify Game_order not found %s", typ, game_order)
		return false, "GameOrder Param Err"
	}
	if data.IsIOS() {
		if cfgInfo.IOS_Rmb_Price != amount {
			logs.Error("[%s] IOSPayNotify price not match, order %s price %d %d",
				typ, game_order, cfgInfo.IOS_Rmb_Price, amount)
			return false, "GameOrder Param Err"
		}
	} else {
		if cfgInfo.Android_Rmb_Price != amount {
			logs.Error("[%s] PayNotify price not match, order %s price %d %d %s",
				typ, game_order, cfgInfo.Android_Rmb_Price, amount, uid)
			return false, "GameOrder Param Err"
		}
	}
	return true, ""
}

func getProperMoneyLevel(typ string, data sdkPayData, game_order string, amount uint32, uid string) (bool, uint32) {
	idx, err := strconv.Atoi(game_order)
	if err != nil {
		logs.Error("[%s] PayNotify Game_order err %s", typ, game_order)
		return false, amount
	}

	cfgInfo := gamedata.GetIAPInfo(uint32(idx))
	if cfgInfo == nil {
		logs.Error("[%s] PayNotify Game_order not found %s", typ, game_order)
		return false, amount
	}
	if data.IsIOS() {
		if cfgInfo.IOS_Rmb_Price <= amount {
			return true, amount - cfgInfo.IOS_Rmb_Price
		}
	} else {
		if cfgInfo.Android_Rmb_Price <= amount {
			return true, amount - cfgInfo.Android_Rmb_Price
		}
	}
	return false, amount
}

type cp_info struct {
	gid        int
	sid        string
	uid        string
	game_order string
	isDBDebug  bool
}

func genCPInfo(typ, Game_order string) *cp_info {
	goss := strings.Split(Game_order, ":")
	if len(goss) < 5 {
		logs.Error("[%s] PayNotify Game_order parse err %s", typ, Game_order)
		return nil
	}
	gid := goss[0]
	shard := goss[1]
	uuid := goss[2]
	uid := fmt.Sprintf("%s:%s:%s", gid, shard, uuid)
	game_order := goss[4]
	isdbDebug := false
	igid, err := strconv.Atoi(gid)
	if err != nil {
		logs.Error("[%s] gid err %s, %s", typ, err.Error(), Game_order)
		return nil
	} else {
		isdbDebug = igid < gid_test_uplimit
	}
	info := &cp_info{
		gid:        igid,
		sid:        shard,
		uid:        uid,
		game_order: game_order,
		isDBDebug:  isdbDebug,
	}
	return info
}
