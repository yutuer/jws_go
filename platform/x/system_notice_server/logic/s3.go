package logic

import (
	"strings"

	"vcs.taiyouxi.net/platform/planx/util/cloud_db"
	"vcs.taiyouxi.net/platform/planx/util/cloud_db/s3db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	db cloud_db.CloudDb
)

func InitS3(region, bucket, accessKey, secretKey string) {
	db = s3db.NewStoreS3(region, bucket,
		accessKey, secretKey, "", "")
	db.Open()
}

func GetNoticeFromS3() []NoticeInfo {
	keys := make([]string, 0)
	lastKey := ""
	for {
		newKeys, err := db.ListObject(lastKey, 100)
		if err != nil {
			logs.Error("get notice from s3 err", err)
			break
		}
		if len(newKeys) == 0 {
			break
		}
		lastKey = newKeys[len(newKeys)-1]
		logs.Debug("get notice from s3 %v", newKeys)
		keys = append(keys, newKeys...)
	}

	logs.Debug("keys from s3 %v", keys)

	retNoticeInfo := make([]NoticeInfo, 0)

	for _, key := range keys {
		if isValidKey(key) {
			gid, version := parseNoticeKey(key)
			if gid == "" || version == "" {
				continue
			}
			bytes, err := db.Get(key)
			if err != nil {
				logs.Error("get notice err", err)
				continue
			}
			noticeStr := string(bytes)
			retNoticeInfo = append(retNoticeInfo, NoticeInfo{
				notice2Client: noticeStr,
				gid:           gid,
				version:       version,
			})
		}
	}
	return retNoticeInfo
}

func isValidKey(key string) bool {
	return strings.HasSuffix(key, ".json")
}

// public/0/gonggao_3.2.0_r.json
// gid, version
func parseNoticeKey(key string) (string, string) {
	splitNotice := strings.Split(key, "/")
	keyLen := len(splitNotice)
	if keyLen < 3 {
		return "", ""
	}
	gid := splitNotice[keyLen-2]
	versionTemp := splitNotice[keyLen-1]
	splitVersion := strings.Split(versionTemp, "_")
	if len(splitVersion) != 3 {
		return "", ""
	}
	if splitVersion[2] != "r.json" {
		return "", ""
	}
	version := splitVersion[1]
	logs.Debug("split notice key gid=%s, version=%s", gid, version)
	return gid, version
}
