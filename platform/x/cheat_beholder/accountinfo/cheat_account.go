package accountinfo

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/utils"
	"github.com/golang/protobuf/proto"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	CBConfig "vcs.taiyouxi.net/platform/x/cheat_beholder/config"
	"vcs.taiyouxi.net/platform/x/gift_sender/core"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

type accountInfo struct {
	//玩家昵称，ID，玩家战力，最远关卡id，最远关卡推荐战力，VIP等级
	Name         string
	PlayerACID   string
	PlayerCurrGS int
	MostLevelId  int
	LevelGS      int
	vipLevel     int
}

const (
	scale float64 = 70
)

var (
	gdTrialLvlInfo  map[int32]*ProtobufGen.LEVEL_TRIAL
	gdTrialLvlOrder map[int32]*ProtobufGen.LEVEL_TRIAL
)

func init() {
	//加载data文件
	dataRelPath := "data"
	dataAbsPath := filepath.Join(getMyDataPath(), dataRelPath)
	load := func(dfilepath string, loadfunc func(string)) {
		loadfunc(filepath.Join("", dataAbsPath, dfilepath))
		logs.Trace("LoadGameData %s success", dfilepath)
	}
	load("level_trial.data", loadTrial)
}

func getMyDataPath() string {
	workPath, _ := os.Getwd()
	workPath, _ = filepath.Abs(workPath)
	// initialize default configurations
	AppPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	appConfigPath := filepath.Join(AppPath, "conf")

	if workPath != AppPath {
		if utils.FileExists(appConfigPath) {
			os.Chdir(AppPath)
		} else {
			appConfigPath = filepath.Join(workPath, "conf")
		}
	}
	return appConfigPath
}

func loadTrial(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.LEVEL_TRIAL_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	gdTrialLvlInfo = make(map[int32]*ProtobufGen.LEVEL_TRIAL, len(ar.Items))
	gdTrialLvlOrder = make(map[int32]*ProtobufGen.LEVEL_TRIAL, len(ar.Items))
	var minIndex int32
	minIndex = math.MaxInt32
	var maxIndex int32
	for _, rec := range ar.Items {
		gdTrialLvlInfo[rec.GetLevelID()] = rec
		gdTrialLvlOrder[rec.GetTrialIndex()] = rec
		if rec.GetTrialIndex() < minIndex {
			minIndex = rec.GetTrialIndex()
		}
		if rec.GetTrialIndex() > maxIndex {
			maxIndex = rec.GetTrialIndex()
		}
	}
}
func loadBin(cfgname string) ([]byte, error) {
	errgen := func(err error, extra string) error {
		return fmt.Errorf("gamex.models.gamedata loadbin Error, %s, %s", extra, err.Error())
	}

	//	path := GetDataPath()
	//	appConfigPath := filepath.Join(path, cfgname)

	file, err := os.Open(cfgname)
	if err != nil {
		return nil, errgen(err, "open")
	}

	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return nil, errgen(err, "stat")
	}

	buffer := make([]byte, fi.Size())
	_, err = io.ReadFull(file, buffer) //read all content
	if err != nil {
		return nil, errgen(err, "readfull")
	}

	return buffer, nil
}

func Iterator(Serid string, w *csv.Writer) {
	serids := strings.Split(Serid, ":")
	logs.Trace("[cyt]serids:%v", serids)
	gid, err := strconv.Atoi(serids[0])
	if err != nil {
		logs.Error("[cyt]gid Strconv error :%v", err)
		return
	}
	sid, err := strconv.Atoi(serids[1])
	if err != nil {
		logs.Error("[cyt]sid Strconv error :%v", err)
		return
	}

	iter := 0
	_realSid, err := etcd.Get(fmt.Sprintf("%s/%d/%d/gm/mergedshard", CBConfig.CommonConfig.Etcd_Root, gid, sid))
	if err != nil {
		logs.Error("[cyt]get realSid err :%v", err)
		return
	}
	realSid, err := strconv.Atoi(_realSid)
	if err != nil {
		logs.Error("[cyt]change realSid from string to int err :%v", err)
		return
	}

	//链接区服服务器
	proFileConn, err := core.GetGamexProfileRedisConn(uint(gid), uint(realSid))
	if err != nil {
		logs.Error("[cyt]error when conn gamex profile redis:%v", err)
		return
	}
	defer func() {
		if proFileConn != nil {
			proFileConn.Close()
		}
	}()

	for {
		reply, err := proFileConn.Do("HSCAN", fmt.Sprintf("names:%v:%v", gid, realSid), iter)
		if err != nil {
			logs.Error("[cyt]HSCAN fail error data:%v", err)
			break
		}
		reply1, err := redis.Values(reply, nil)
		if len(reply1) != 2 || err != nil {
			logs.Error("HSCAN result not slice ,fail,error data:%v", err)
			break
		}
		iter, err = redis.Int(reply1[0], nil)
		if err != nil {
			logs.Error("[cyt]HSCAN next index is not integer,fali,error data:%v", err)
			break
		}
		keys, err := redis.Strings(reply1[1], nil)
		if err != nil {
			logs.Error("[cyt]HSCAN result can not change to string,fali,error data:%v", err)
			break
		}

		//keys内为玩家的acids
		for i, key := range keys {
			if i%2 != 0 && !strings.Contains(key, "TPvp") && !strings.Contains(key, "wspvp") {
				MyAccountInfo, err := getAccountInfo(key, proFileConn)
				if err != nil {
					logs.Error("[cyt]get player:%v data file,error data:%v", key, err)
					continue
				}
				logs.Trace("[cyt]player data is :%v", MyAccountInfo)
				//作弊认定
				if scale*float64(MyAccountInfo.LevelGS)/100 >= float64(MyAccountInfo.PlayerCurrGS) {
					logs.Trace("[cyt]affirm palyer:%v cheat", MyAccountInfo.Name)
					w.Write([]string{MyAccountInfo.Name, MyAccountInfo.PlayerACID,
						fmt.Sprintf("%d", MyAccountInfo.PlayerCurrGS),
						fmt.Sprintf("%d", MyAccountInfo.MostLevelId),
						fmt.Sprintf("%d", MyAccountInfo.LevelGS),
						fmt.Sprintf("%d", MyAccountInfo.vipLevel),
					}) //写入数据
				}
			}
		}
		if iter == 0 {
			logs.Debug("[cyt]loop over")
			break
		}
	}
}

func getAccountInfo(acid string, proFileConn redis.Conn) (accountInfo, error) {
	profileID := "profile:" + acid
	var result accountInfo = accountInfo{}

	//玩家ID
	result.PlayerACID = acid

	//玩家详细信息
	_args, err := proFileConn.Do("HMGET", profileID, "name", "data", "trial", "v")
	if err != nil {
		logs.Error("[cyt]HMGET player ACID:%v name,data,trial,vip data error", profileID)
		return accountInfo{}, err
	}
	args, err := redis.Values(_args, nil)
	if err != nil {
		logs.Error("[cyt]HMGET result is not integer array")
		return accountInfo{}, err
	}

	//玩家昵称

	playername, err := redis.String(args[0], nil)

	if err != nil {
		logs.Error("[cyt] get player name error:%v", err)
		return result, err
	}

	result.Name = playername

	//玩家战力
	playergs := args[1]

	if playergs != nil {
		var profileData account.ProfileData
		err = json.Unmarshal(playergs.([]byte), &profileData)
		if err != nil {
			logs.Error("[cyt]Unmarshal player gs data error:%v", err)
			return result, err
		} else {
			result.PlayerCurrGS = profileData.CorpCurrGS_HistoryMax
		}
	}

	//玩家爬塔信息
	playerTrial := args[2]

	if playerTrial != nil {
		var profileTrial account.PlayerTrial
		err = json.Unmarshal(playerTrial.([]byte), &profileTrial)
		if err != nil {
			logs.Error("[cyt]Unmarshal palyer trial data error:%v", err)
			return result, err
		} else {
			result.MostLevelId = int(profileTrial.MostLevelId)
		}
	}

	//最远关卡的推荐战力
	result.LevelGS = int(gdTrialLvlInfo[int32(result.MostLevelId)].GetLevelGS())

	//玩家vip信息
	playerVIP := args[3]

	if playerVIP != nil {
		var profileVIP account.VIP
		err = json.Unmarshal(playerVIP.([]byte), &profileVIP)
		if err != nil {
			logs.Error("[cyt]Unmarshal player vip data error:%v", err)
			return result, err
		} else {
			result.vipLevel = int(profileVIP.V)
		}
	}

	return result, nil
}
