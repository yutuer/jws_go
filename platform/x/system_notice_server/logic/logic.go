package logic

import (
	"sync"

	"encoding/json"

	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/sys_public"
)

type NoticeInfo struct {
	notice2Client  string
	gid            string
	version        string
	noticeInServer sys_public.PublicSetToServer
}

var lock sync.RWMutex
var NoticeList []NoticeInfo

func GetNotice(gid, version string) string {
	lock.RLock()
	defer lock.RUnlock()
	for _, info := range NoticeList {
		if info.gid == gid && info.version == version {
			return info.notice2Client
		}
	}
	return ""
}

func AddNotice(gid, version, newNotice string) {
	lock.Lock()
	defer lock.Unlock()
	for i, info := range NoticeList {
		if info.gid == gid && info.version == version {
			NoticeList[i].notice2Client = newNotice
			return
		}
	}
	NoticeList = append(NoticeList, NoticeInfo{
		notice2Client: newNotice,
		gid:           gid,
		version:       version,
	})
}

func GetAllNotice() []NoticeInfo {
	lock.RLock()
	defer lock.RUnlock()
	return NoticeList
}

func SetNotice(newNoticeInfo []NoticeInfo) {
	lock.Lock()
	defer lock.Unlock()
	NoticeList = newNoticeInfo
}

func InitNotice() {
	updateNotice()
}

func updateNotice() {
	noticeList := GetNoticeFromS3()
	filterNotice := filterExpireNotice(noticeList)
	SetNotice(filterNotice)
	logs.Info("update notice %v", GetAllNotice())
}

func filterExpireNotice(noticeList []NoticeInfo) []NoticeInfo {
	for i, notice := range noticeList {
		noticeInfo := &sys_public.PublicSetToServer{}
		err := json.Unmarshal([]byte(notice.notice2Client), noticeInfo)
		if err != nil {
			logs.Error("filterExpireNotice err", err)
			continue
		}
		noticeInfo.Publics = removeExpireNotice(noticeInfo.Publics)
		noticeInfo.Maintaince = removeExpireNotice(noticeInfo.Maintaince)
		noticeInfo.Forceupdate = removeExpireNotice(noticeInfo.Forceupdate)
		newBytes, err := json.Marshal(noticeInfo)
		if err != nil {
			logs.Error("filterExpireNotice err", err)
			continue
		}
		noticeList[i].notice2Client = string(newBytes)
		noticeList[i].noticeInServer = *noticeInfo
	}
	return noticeList
}

func removeExpireNotice(notices []sys_public.PublicToServer) []sys_public.PublicToServer {
	retNotice := make([]sys_public.PublicToServer, 0)
	nowTime := time.Now().Unix()
	for _, notice := range notices {
		if notice.End > nowTime {
			retNotice = append(retNotice, notice)
		}
	}
	return retNotice
}

func filterInTimeNotice(notices []sys_public.PublicToServer) []sys_public.PublicToServer {
	retNotice := make([]sys_public.PublicToServer, 0)
	nowTime := time.Now().Unix()
	for _, notice := range notices {
		if notice.End > nowTime && notice.SendTime < nowTime {
			retNotice = append(retNotice, notice)
		}
	}
	return retNotice
}

func StartUpdateServer() {
	go func() {
		timer := time.After(time.Second * 60)
		for {
			<-timer
			timer = time.After(time.Second * 60)
			updateNotice2Client()
		}
	}()
}

// 过滤不在时间范围内的公告
func updateNotice2Client() {
	noticeList := GetAllNotice()
	notice2ClientList := make([]string, len(noticeList))
	for i, notice := range noticeList {
		notice.noticeInServer.Publics = filterInTimeNotice(notice.noticeInServer.Publics)
		notice.noticeInServer.Maintaince = filterInTimeNotice(notice.noticeInServer.Maintaince)
		notice.noticeInServer.Forceupdate = filterInTimeNotice(notice.noticeInServer.Forceupdate)
		newBytes, err := json.Marshal(notice.noticeInServer)
		if err != nil {
			logs.Error("filterExpireNotice err", err)
			continue
		}
		notice2ClientList[i] = string(newBytes)
	}
	lock.Lock()
	defer lock.Unlock()
	for i := range notice2ClientList {
		NoticeList[i].notice2Client = notice2ClientList[i]
	}
}
