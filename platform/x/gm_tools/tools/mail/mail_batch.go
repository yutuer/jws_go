package mail

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/store"
)

const (
	CSV_MAIL_INDEX = iota
	CSV_MAIL_UID
	CSV_MAIL_SERVER
	CSV_MAIL_TITLE
	CSV_MAIL_INFO
	CSV_MAIL_TIME_BEGIN
	CSV_MAIL_TIME_END
	CSV_MAIL_TAG
	CSV_MAIL_REWARD
)

// 单行数据
const (
	_                          = iota
	CSV_MAIL_ERROR_TIME_FORMAT //
	CSV_MAIL_ERROR_ITEM_FORMAT //
	CSV_MAIL_ERROR_ID          //
	CSV_MAIL_ERROR_SERVER_NAME
	CSV_MAIL_ERROR_SEND
	CSV_MAIL_ERROR_ACID
)

type BatchDealResult struct {
	SuccessCount      int
	FailArr           []SingleBatchDeal
	CurrentBatchIndex int
}

type BatchMailSuccessInfo struct {
	BatchIndex int
	InfoList   []SuccessMailInfo
}

type SuccessMailInfo struct {
	CsvIndex int
	Uid      string
	Idx      int64
}

type SingleBatchDeal struct {
	Id       int
	FailCode int
}

type SendMailInfo struct {
	csvIndex    int
	serverName  string
	profileAcid string
	mailReward  timail.MailReward
}

var BatchMailIndex int64

// 返回发送结果
func sendMailByBatch(records [][]string) (BatchDealResult, error) {
	ret := BatchDealResult{FailArr: make([]SingleBatchDeal, 0, len(records))}
	for _, csvData := range records {
		csvIndex, serverName, profileMail, mailReward, errCode := parseCsv2Mail(csvData)
		if errCode != 0 {
			ret.FailArr = append(ret.FailArr, SingleBatchDeal{Id: csvIndex, FailCode: errCode})
			continue
		}
		err := SendMailWithAddon(serverName, profileMail, &mailReward, csvIndex)
		if err != nil {
			logs.Warn("fail to send mail in batch, %d, %d, %V", serverName, profileMail, mailReward)
			if err.Error() == "NoServerName" {
				ret.FailArr = append(ret.FailArr, SingleBatchDeal{Id: csvIndex, FailCode: CSV_MAIL_ERROR_SERVER_NAME})
			} else {
				ret.FailArr = append(ret.FailArr, SingleBatchDeal{Id: csvIndex, FailCode: CSV_MAIL_ERROR_SEND})
			}
		}
	}
	ret.SuccessCount = len(records) - len(ret.FailArr)
	return ret, nil
}

func sendMailByBatch2(records [][]string, batchMailIdx int) (BatchDealResult, error) {
	lock := sync.RWMutex{}
	lock.Lock()
	ret := BatchDealResult{FailArr: make([]SingleBatchDeal, 0, len(records))}
	info := &BatchMailSuccessInfo{InfoList: make([]SuccessMailInfo, 0), BatchIndex: batchMailIdx}
	lock.Unlock()
	sendMailMap := make(map[string][]SendMailInfo, 0)
	for _, csvData := range records {
		csvIndex, serverName, profileMail, mailRewardInfo, errCode := parseCsv2Mail(csvData)
		if errCode != 0 {
			ret.FailArr = append(ret.FailArr, SingleBatchDeal{Id: csvIndex, FailCode: errCode})
			continue
		}
		sendMail := SendMailInfo{
			csvIndex:    csvIndex,
			serverName:  serverName,
			profileAcid: profileMail,
			mailReward:  mailRewardInfo,
		}
		if list, ok := sendMailMap[serverName]; ok {
			sendMailMap[serverName] = append(list, sendMail)
		} else {
			sendMailMap[serverName] = append([]SendMailInfo{}, sendMail)
		}
	}
	waitter := util.WaitGroupWrapper{}
	logs.Info("send mail map %v", sendMailMap)
	for key, value := range sendMailMap {
		serverName := key
		mails := value
		waitter.Wrap(func() {
			logs.Info("send mail by server name %s", serverName)
			tempKeys, failedMails, err := sendMailByServerName(serverName, mails)
			if err != nil {
				// 全部失败
				lock.Lock()
				for _, tmail := range mails {
					ret.FailArr = append(ret.FailArr, SingleBatchDeal{Id: tmail.csvIndex, FailCode: CSV_MAIL_ERROR_SEND})
				}
				lock.Unlock()
				logs.Error("batch send mail err ", err)
			} else {
				// 部分成功， 部分失败
				for _, failCsvIndex := range failedMails {
					ret.FailArr = append(ret.FailArr, SingleBatchDeal{Id: failCsvIndex, FailCode: CSV_MAIL_ERROR_SEND})
				}
				lock.Lock()
				info.InfoList = append(info.InfoList, tempKeys...)
				lock.Unlock()
			}
		})
	}
	waitter.Wait()
	saveBatchMail(info, BatchMailIndex)
	ret.SuccessCount = len(records) - len(ret.FailArr)
	ret.CurrentBatchIndex = int(BatchMailIndex)
	return ret, nil
}

func saveBatchMail(info *BatchMailSuccessInfo, batchIndex int64) {
	successInfoKey := fmt.Sprintf("batch_mail_info_%d", info.BatchIndex)
	infoBytes := bytes.NewBuffer([]byte{})
	csvW := csv.NewWriter(infoBytes)
	for _, mail := range info.InfoList {
		csvW.Write([]string{fmt.Sprintf("%d", mail.CsvIndex), mail.Uid, fmt.Sprintf("id%d", mail.Idx)})
	}
	csvW.Flush()
	logs.Warn("<Batch Mail> info save db key=%s, info=%v", successInfoKey, string(infoBytes.Bytes()))
	if err := store.Set(store.NormalBucket, successInfoKey, infoBytes.Bytes()); err != nil {
		logs.Error("<Batch Mail> info save db err, %s, %v", err.Error(), *info)
	}
	if err := store.Set(store.NormalBucket, "batch_mail_index", []byte(fmt.Sprintf("%d", batchIndex))); err != nil {
		logs.Error("<Batch Mail> index save db err, %s, %v", err.Error(), batchIndex)
	}
}

func sendMailByServerName(serverName string, mails []SendMailInfo) ([]SuccessMailInfo, []int, error) {
	cfg := CommonCfg.GetServerCfgFromName(serverName)
	for i := range mails {
		// TODO潜在的BUG是同一个人的邮件超过1000封
		mails[i].mailReward.Idx = timail.MkMailId(timail.Mail_Send_By_Sys, int64(i)%timail.Mail_Id_Gen_Base)
	}
	if cfg.RedisName == "" ||
		CommonCfg.AWS_Region == "" ||
		cfg.MailDBName == "" {
		logs.Error("SendMail Err No Name %v %s|%s|%s|%s", serverName,
			CommonCfg.AWS_Region,
			cfg.MailDBName,
			CommonCfg.AWS_AccessKey,
			CommonCfg.AWS_SecretKey)
		return nil, nil, errors.New("NoServerName")
	}

	m, err := mailhelper.NewMailDriver(mailhelper.MailConfig{
		AWSRegion:    CommonCfg.AWS_Region,
		DBName:       cfg.MailDBName,
		AWSAccessKey: CommonCfg.AWS_AccessKey,
		AWSSecretKey: CommonCfg.AWS_SecretKey,
		MongoDBUrl:   cfg.MailMongoUrl,
		DBDriver:     cfg.MailDBDriver,
	})

	if err != nil {
		return nil, nil, err
	}

	successMailMap := make(map[timail.MailKey]int, 0)

	for _, mail := range mails {
		successMailMap[timail.MailKey{Uid: mail.profileAcid, Idx: mail.mailReward.Idx}] = mail.csvIndex
	}

	failMailIds := make([]timail.MailKey, 0)

	mailLen := len(mails)
	for i := 0; i <= mailLen/25; i++ {
		startIndex := i * 25
		lastIndex := (i + 1) * 25
		if lastIndex > mailLen {
			lastIndex = mailLen
		}
		if startIndex < lastIndex {
			failIds := sendMailBy25(mails[startIndex:lastIndex], m)
			failMailIds = append(failMailIds, failIds...)
		}
	}

	for _, mail := range failMailIds {
		delete(successMailMap, timail.MailKey{Uid: mail.Uid, Idx: mail.Idx})
	}

	successMails := make([]SuccessMailInfo, 0)
	for key, value := range successMailMap {
		successMails = append(successMails, SuccessMailInfo{
			CsvIndex: value,
			Uid:      key.Uid,
			Idx:      key.Idx,
		})
	}

	failCsvIndex := make([]int, 0)
	for _, failMail := range failMailIds {
		for _, mail := range mails {
			if failMail.Idx == mail.mailReward.Idx && failMail.Uid == mail.profileAcid {
				failCsvIndex = append(failCsvIndex, mail.csvIndex)
				break
			}
		}
	}
	return successMails, failCsvIndex, err
}

func sendMailBy25(mails []SendMailInfo, m timail.Timail) []timail.MailKey {
	mailAcids := make([]string, 0)
	mailRewards := make([]timail.MailReward, 0)

	for _, mail := range mails {
		mailAcids = append(mailAcids, mail.profileAcid)
		mailRewards = append(mailRewards, mail.mailReward)
	}

	err, failMailIds := m.BatchWriteMails(mailAcids, mailRewards)
	if err != nil {
		logs.Error("batch write mail err", err)
		retFails := make([]timail.MailKey, 0)
		for i := 0; i < len(mailAcids); i++ {
			retFails = append(retFails, timail.MailKey{
				Uid: mailAcids[i],
				Idx: mailRewards[i].Idx,
			})
		}
		return retFails
	} else {
		return failMailIds
	}
}

// 返回 index serverName, mailProfile, mail, errCode
func parseCsv2Mail(csvData []string) (int, string, string, timail.MailReward, int) {
	mailReward := timail.MailReward{}
	mailReward.Param = []string{csvData[CSV_MAIL_TITLE], csvData[CSV_MAIL_INFO]}
	mailReward.Reason = "GMSend"
	if csvData[CSV_MAIL_TAG] == "" {
		mailReward.Tag = "{}"
	} else {
		mailReward.Tag = fmt.Sprintf("{\"ver\":\"%s\"}", csvData[CSV_MAIL_TAG])
	}
	index, err := strconv.ParseInt(csvData[CSV_MAIL_INDEX], 10, 32)
	if err != nil {
		return 0, "", "", mailReward, CSV_MAIL_ERROR_ID
	}
	var errCode int
	mailReward.TimeBegin, mailReward.TimeEnd, errCode = parseTime(csvData[CSV_MAIL_TIME_BEGIN], csvData[CSV_MAIL_TIME_END])
	if errCode != 0 {
		return int(index), "", "", mailReward, errCode
	}
	mailReward.ItemId, mailReward.Count, errCode = parseItems(csvData[CSV_MAIL_REWARD])
	if errCode != 0 {
		return int(index), "", "", mailReward, errCode
	}
	colonCount := strings.Count(csvData[CSV_MAIL_UID], ":")
	if colonCount != 2 {
		return int(index), "", "", mailReward, CSV_MAIL_ERROR_ACID
	}
	mailUid := fmt.Sprintf("profile:%s", csvData[CSV_MAIL_UID])
	return int(index), csvData[CSV_MAIL_SERVER], mailUid, mailReward, 0
}

func parseTime(start, end string) (int64, int64, int) {
	startSt, err := time.ParseInLocation("20060102_15:04", start, util.ServerTimeLocal)
	if err != nil {
		logs.Warn("parse time err, %s, %v", start, err)
		return 0, 0, CSV_MAIL_ERROR_TIME_FORMAT
	}
	endSt, err := time.ParseInLocation("20060102_15:04", end, util.ServerTimeLocal)
	if err != nil {
		logs.Warn("parse time err, %s, %v", end, err)
		return 0, 0, CSV_MAIL_ERROR_TIME_FORMAT
	}
	return startSt.Unix(), endSt.Unix(), 0
}

func parseItems(itemStr string) ([]string, []uint32, int) {
	if itemStr == "" {
		return nil, nil, 0
	}
	itemArray := strings.Split(itemStr, ";")
	itemIds := make([]string, len(itemArray))
	itemCounts := make([]uint32, len(itemArray))
	for i, item := range itemArray {
		itemDetail := strings.Split(item, ",")
		if len(itemDetail) != 2 {
			return nil, nil, CSV_MAIL_ERROR_ITEM_FORMAT
		}
		itemIds[i] = itemDetail[0]
		count, err := strconv.ParseInt(itemDetail[1], 10, 32)
		if err != nil {
			return nil, nil, CSV_MAIL_ERROR_ITEM_FORMAT
		}
		itemCounts[i] = uint32(count)
	}
	return itemIds, itemCounts, 0
}

func delAllMail(csvData [][]string) (BatchDealResult, error) {
	ret := BatchDealResult{FailArr: make([]SingleBatchDeal, 0)}
	for _, csvData := range csvData {
		uid := fmt.Sprintf("profile:%s", csvData[CSV_MAIL_UID])
		serverName := csvData[CSV_MAIL_SERVER]
		csvIndexStr := csvData[CSV_MAIL_INDEX]
		csvIndex, _ := strconv.Atoi(csvIndexStr)
		cfg := CommonCfg.GetServerCfgFromName(serverName)
		m, err := mailhelper.NewMailDriver(mailhelper.MailConfig{
			AWSRegion:    CommonCfg.AWS_Region,
			DBName:       cfg.MailDBName,
			AWSAccessKey: CommonCfg.AWS_AccessKey,
			AWSSecretKey: CommonCfg.AWS_SecretKey,
			MongoDBUrl:   cfg.MailMongoUrl,
			DBDriver:     cfg.MailDBDriver,
		})
		if err != nil {
			logs.Error("del all mail err %s", uid)
			ret.FailArr = append(ret.FailArr, SingleBatchDeal{Id: csvIndex, FailCode: CSV_MAIL_ERROR_SERVER_NAME})
			continue
		}
		mailList, err := m.LoadAllMailByGM(uid)
		if err != nil {
			logs.Error("load all mail err %s", uid)
			ret.FailArr = append(ret.FailArr, SingleBatchDeal{Id: csvIndex, FailCode: CSV_MAIL_ERROR_ACID})
			continue
		}
		toDel := make([]int64, 0)
		for _, itMail := range mailList {
			toDel = append(toDel, itMail.Idx)
		}
		m.SyncMail(uid, nil, toDel)
	}
	ret.SuccessCount = len(csvData) - len(ret.FailArr)
	return ret, nil
}

func getBatchMailDetail(c *gm_command.Context, server, accountid string, params []string) error {
	batchIndex := params[0]
	successInfoKey := fmt.Sprintf("batch_mail_info_%s", batchIndex)
	logs.Warn("<Batch Mail> info get save db key=%s", successInfoKey)
	retBytes, err := store.Get(store.NormalBucket, successInfoKey)
	if err != nil {
		return err
	}
	c.SetData(string(retBytes))
	logs.Debug("<Batch Mail> get batch mail detail data, %s", c.GetData())
	return nil
}

func getBatchMailIndex(c *gm_command.Context, server, accountid string, params []string) error {
	c.SetData(fmt.Sprintf("%d", BatchMailIndex))
	return nil
}

func batchDelMail(csvData [][]string) (BatchDealResult, error) {
	ret := BatchDealResult{FailArr: make([]SingleBatchDeal, 0)}
	// TODO 多个邮件一起删除
	for i, csvData := range csvData {
		uid := csvData[1]
		idx, err := strconv.ParseInt(csvData[2][2:], 10, 64)
		if err != nil {
			ret.FailArr = append(ret.FailArr, SingleBatchDeal{Id: i, FailCode: CSV_MAIL_ERROR_ID})
			continue
		}
		var ac db.Account
		if strings.HasPrefix(uid, "profile:") {
			ac, err = db.ParseAccount(uid[8:])
		} else {
			ac, err = db.ParseAccount(uid)
		}

		if err != nil {
			ret.FailArr = append(ret.FailArr, SingleBatchDeal{Id: i, FailCode: CSV_MAIL_ERROR_SERVER_NAME})
			continue
		}
		cfg := CommonCfg.GetServerCfgFromName(ac.ServerString())
		m, err := mailhelper.NewMailDriver(mailhelper.MailConfig{
			AWSRegion:    CommonCfg.AWS_Region,
			DBName:       cfg.MailDBName,
			AWSAccessKey: CommonCfg.AWS_AccessKey,
			AWSSecretKey: CommonCfg.AWS_SecretKey,
			MongoDBUrl:   cfg.MailMongoUrl,
			DBDriver:     cfg.MailDBDriver,
		})
		if err != nil {
			logs.Error("del all mail err %s", uid)
			ret.FailArr = append(ret.FailArr, SingleBatchDeal{Id: i, FailCode: CSV_MAIL_ERROR_SEND})
			continue
		}

		m.SyncMail(uid, nil, []int64{int64(idx)})
	}
	ret.SuccessCount = len(csvData) - len(ret.FailArr)
	return ret, nil
}
