package mail

import (
	"fmt"

	"strings"

	"encoding/csv"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"strconv"
	"time"
	gmTimeUtil "vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/store"
	gmConfig "vcs.taiyouxi.net/platform/x/gm_tools/config"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/ban_account"
	"vcs.taiyouxi.net/platform/x/gm_tools/util"
)

// 测试邮件显示BUG的数据, TODO 删除
func loadTestJson() string {
	path := config.NewConfigPath("testJson.json")
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err.Error()
	} else {
		return string(file)
	}
}

func RouteMailGet(r *gin.Engine) {
	r.GET("/api/v1/mail/:server_name/*mail_profile_name", func(c *gin.Context) {
		mail_profile_name := c.Param("mail_profile_name")
		server_name := c.Param("server_name")
		mail_profile_name = util.DeleteBackslash(mail_profile_name)
		logs.Info("mail get %s %s", server_name, mail_profile_name)

		data, err := GetAllMail(server_name, mail_profile_name)
		if err == nil {
			c.String(200, data)
		} else {
			c.String(400, err.Error())
		}
	})

	r.POST("/api/v1/mail/:server_name/*mail_profile_name", func(c *gin.Context) {
		mail_profile_name := c.Param("mail_profile_name")
		server_name := c.Param("server_name")
		mail_profile_name = util.DeleteBackslash(mail_profile_name)
		logs.Info("mail get %s %s", server_name, mail_profile_name)

		s_old := struct {
			Info             string
			Title            string
			ItemId           []string
			Count            []uint32
			TimeBegin        int64
			TimeBeginString  string
			TimeEnd          int64
			TimeEndString    string
			CTimeBegin       int64
			CTimeBeginString string
			CTimeEnd         int64
			CTimeEndString   string
			Reason           string
			Tag              string
			IsGetted         bool
			Idx              int64
			Accounts         []string
			SendAllSer       bool
			Lang             string
			MultiLang        string
		}{}
		err := c.Bind(&s_old)

		logs.Info("AddMailInfo before timelocal change : %s %v %s, %v", server_name, s_old.SendAllSer,
			mail_profile_name, s_old)
		//转换传入的时间，服务器存储传入时间字符串对应的时区时间
		begintime, err := time.ParseInLocation("2006/01/02 15:04:05", s_old.TimeBeginString, gmTimeUtil.ServerTimeLocal)
		if err != nil {
			logs.Error("TimeBeginString cannot be Parsed by location:%v", err)
			return
		}
		s_old.TimeBegin = begintime.Unix()
		endtime, err := time.ParseInLocation("2006/01/02 15:04:05", s_old.TimeEndString, gmTimeUtil.ServerTimeLocal)
		if err != nil {
			logs.Error("TimeEndString cannot be Parsed by location:%v", err)
			return
		}
		s_old.TimeEnd = endtime.Unix()

		if s_old.CTimeBegin != 0 {
			cbtime, err := time.ParseInLocation("2006/01/02 15:04:05", s_old.CTimeBeginString, gmTimeUtil.ServerTimeLocal)
			if err != nil {
				logs.Error("CreateBeforeString cannot be Parsed by location:%v", err)
				return
			}
			s_old.CTimeBegin = cbtime.Unix()
		}

		if s_old.CTimeEnd != 0 {
			cetime, err := time.ParseInLocation("2006/01/02 15:04:05", s_old.CTimeEndString, gmTimeUtil.ServerTimeLocal)
			if err != nil {
				logs.Error("CreateEndString cannot be Parsed by location:%v", err)
				return
			}
			s_old.CTimeEnd = cetime.Unix()
		}

		logs.Info("AddMailInfo after timelocal change : %s %v %s, %v", server_name, s_old.SendAllSer,
			mail_profile_name, s_old)

		s := timail.MailReward{
			ItemId:       s_old.ItemId,
			Count:        s_old.Count,
			TimeBegin:    s_old.TimeBegin,
			TimeEnd:      s_old.TimeEnd,
			Reason:       s_old.Reason,
			Tag:          s_old.Tag,
			IsGetted:     s_old.IsGetted,
			Idx:          s_old.Idx,
			Account2Send: s_old.Accounts,
			CreateBefore: s_old.CTimeBegin,
			CreateEnd:    s_old.CTimeEnd,
		}

		s.Param = []string{s_old.Title, s_old.Info, s_old.MultiLang, s_old.Lang}

		s.Reason = "GMSend"
		if s.Tag == "" {
			s.Tag = "{}"
		}

		if err != nil {
			c.String(400, err.Error())
			logs.Error("POST Bind failed :%v", err)
			return
		}

		if strings.HasPrefix(mail_profile_name, "all") {
			var hc_c uint32
			for i, id := range s.ItemId {
				if strings.Contains(id, "VI_HC") {
					hc_c += s.Count[i]
				}
			}
			if hc_c > 500 {
				c.String(401, "VI_HC can not more then 500")
				return
			}
		}
		if server_name != "" && s_old.SendAllSer {
			allser := gmConfig.Cfg.GetGidAllServer(server_name)
			err_str := ""
			for _, ser := range allser {
				ss_p := strings.Split(mail_profile_name, ":")
				ss_s := strings.Split(ser.ServerName, ":")
				mail_profile_name_ser := strings.Join([]string{ss_p[0], ss_s[1]}, ":")
				err = SendMail(ser.ServerName, mail_profile_name_ser, &s)
				if err != nil {
					logs.Error("Send gid mail fail, ser %s %s",
						ser.ServerName, err.Error())
					err_str += fmt.Sprintf("%s fail, ", ser.ServerName)
				}
			}
			if err_str == "" {
				c.String(200, string("ok"))
			} else {
				c.String(401, err_str)
			}
		} else {
			err = SendMail(server_name, mail_profile_name, &s)
			if err == nil {
				c.String(200, string("ok"))
			} else {
				c.String(401, err.Error())
			}
		}
	})

	r.POST("/api/v1/batchmail", func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			logs.Warn("fail to load file")
			c.String(401, err.Error())
			return
		}
		filename := header.Filename
		logs.Debug("receive csv file, %s", filename)
		csvFile := csv.NewReader(file)
		csvData, err := csvFile.ReadAll()
		if err != nil {
			logs.Warn("fail to parse csv file, %v", err)
			c.String(401, err.Error())
			return
		}
		batchMailIdx, err := strconv.Atoi(csvData[0][0])
		if err != nil {
			logs.Warn("fail to parse mail index, %v", err)
			c.String(401, err.Error())
			return
		}
		if batchMailIdx <= int(BatchMailIndex) {
			logs.Warn("batch mail index too low, %d < %d", batchMailIdx, BatchMailIndex)
			c.String(401, "batch mail index too low")
			return
		}
		BatchMailIndex = int64(batchMailIdx)
		result, err := sendMailByBatch2(csvData[1:], batchMailIdx)
		if err != nil {
			logs.Warn("fail to send mail, %v", err)
			c.String(401, err.Error())
			return
		}
		retJson, err := json.Marshal(result)
		if err != nil {
			logs.Warn("fail to marshal, %v", err)
			c.String(401, err.Error())
			return
		}
		c.String(200, string(retJson))
	})

	r.POST("/api/v1/batch_del_mail", func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			logs.Warn("fail to load file")
			c.String(401, err.Error())
			return
		}
		filename := header.Filename
		logs.Debug("receive csv file, %s", filename)
		csvFile := csv.NewReader(file)
		csvData, err := csvFile.ReadAll()
		if err != nil {
			logs.Warn("fail to parse csv file, %v", err)
			c.String(401, err.Error())
			return
		}

		result, err := batchDelMail(csvData)
		if err != nil {
			logs.Warn("fail to send mail, %v", err)
			c.String(401, err.Error())
			return
		}
		retJson, err := json.Marshal(result)
		if err != nil {
			logs.Warn("fail to marshal, %v", err)
			c.String(401, err.Error())
			return
		}
		c.String(200, string(retJson))
	})

	r.POST("/api/v1/banAccountsAndDelRanks", func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			logs.Error("fail to load file")
			c.String(401, err.Error())
			return
		}
		filename := header.Filename
		logs.Debug("receive csv file, %s", filename)
		csvFile := csv.NewReader(file)
		csvData, err := csvFile.ReadAll()
		if err != nil {
			logs.Warn("fail to parse csv file, %v", err)
			c.String(401, err.Error())
			return
		}
		result, err := ban_account.BatchBanAccounts(csvData[1:])
		if err != nil {
			logs.Warn("fail to send mail, %v", err)
			c.String(401, err.Error())
			return
		}
		retJson, err := json.Marshal(result)
		if err != nil {
			logs.Warn("fail to marshal, %v", err)
			c.String(401, err.Error())
			return
		}
		c.String(200, string(retJson))
	})
}

func RegCommands() {
	gm_command.AddGmCommandHandle("virtualIAP", VirtualIAP)
	gm_command.AddGmCommandHandle("delVirtualIAP", DelVirtualIAP)
	gm_command.AddGmCommandHandle("getVirtualIAP", GetVirtualIAP)

	gm_command.AddGmCommandHandle("delMail", delMail)
	gm_command.AddGmCommandHandle("virtualtrueIAP", VirtualTrueIAP)

	gm_command.AddGmCommandHandle("getBatchMailDetail", getBatchMailDetail)
	gm_command.AddGmCommandHandle("getBatchMailIndex", getBatchMailIndex)

	InitBatchMailIndex()
}

func InitBatchMailIndex() {
	indexBytes, err := store.Get(store.NormalBucket, "batch_mail_index")
	if err == store.ErrNoKey {
		store.Set(store.NormalBucket, "batch_mail_index", []byte("0"))
		logs.Warn("<Start Gmtool> got batch mail index %d", BatchMailIndex)
		return
	}
	if err != nil {
		logs.Error("<Start Gmtool> batch mail index err", err)
		return
	}
	tempIndex := string(indexBytes)
	if tempIndex == "" {
		BatchMailIndex = 0
	} else {
		index, err := strconv.Atoi(tempIndex)
		if err != nil {
			logs.Error("<Start Gmtool> batch mail atoi err", err)
			return
		}
		BatchMailIndex = int64(index)
	}
	logs.Warn("<Start Gmtool> got batch mail index %d", BatchMailIndex)
}
