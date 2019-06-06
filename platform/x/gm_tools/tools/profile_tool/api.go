package profile_tool

import (
	//"fmt"
	//"strings"
	//"strconv"
	"encoding/json"
	"errors"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	gmConfig "vcs.taiyouxi.net/platform/x/gm_tools/config"
	//"vcs.taiyouxi.net/platform/x/gm_tools/util"
	"strconv"
	"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/platform/x/tool_account2json/account2json"
	"vcs.taiyouxi.net/platform/x/tool_json2account/json2account"
)

func RegCommands() {
	gm_command.AddGmCommandHandle("getNickNameFromRedisByACID", getNickNameFromRedisByACID)
	gm_command.AddGmCommandHandle("getInfoByNickName", getInfoByNickName)
	gm_command.AddGmCommandHandle("getAllInfoByAccountID", getAllInfoByAccountID)
	gm_command.AddGmCommandHandle("setAllInfoByAccountID", setAllInfoByAccountID)
	gm_command.AddGmCommandHandle("cleanUnEquipItems", cleanUnEquipItems)
}

type nickToClient struct {
	Nick string `json:"name"`
	Aid  string `json:"pid"`
}

func getNickNameFromRedisByACID(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Trace("getProfileFromRedis server is:%v,accountid is :%v,params is :%v", server, accountid, params)
	cfg := gmConfig.Cfg.GetServerCfgFromName(server)
	res := make([]nickToClient, 0, len(params))
	for _, pid := range params {
		if pid == "System" {
			continue
		}
		playerNickName, err := GetNickNameFromRedisByACID(cfg.RedisName, server, "profile:"+pid)
		if err == nil {
			res = append(res, nickToClient{string(playerNickName), pid})
		} else {
			logs.Error("err:%v", err)
			res = append(res, nickToClient{"Err by : " + err.Error(), pid})
		}
	}

	j, err := json.Marshal(res)
	if err != nil {
		return err
	}
	c.SetData(string(j))
	return nil
}
func getInfoByNickName(c *gm_command.Context, server, accountid string, params []string) error {
	cfg := gmConfig.Cfg.GetServerCfgFromName(server)
	logs.Info("getInfoByNickName %s %v --> %v", server, params, cfg)

	res := make([]nickToClient, 0, len(params))
	for _, n := range params {
		if n == "System" {
			continue
		}
		pid, err := GetAcountIDFromRedis(cfg.RedisName, server, n)
		if err == nil {
			res = append(res, nickToClient{n, pid})
		} else {
			res = append(res, nickToClient{n, "Err by : " + err.Error()})
		}
	}

	j, err := json.Marshal(res)
	if err != nil {
		return err
	}
	c.SetData(string(j))
	return nil
}

type accountData struct {
	ProfileData string `json:"ProfileData"`
	BagData     string `json:"BagData"`
	GeneralData string `json:"GeneralData"`
	StoreData   string `json:"StoreData"`
	TmpData     string `json:"TmpData"`
}

func getAllInfoByAccountID(c *gm_command.Context, server, accountid string, params []string) error {
	cfg := gmConfig.Cfg.GetServerCfgFromName(server)
	logs.Info("getInfoByNickName  %s %v --> %v", server, params, cfg)

	conn := getRedisConn(cfg.RedisName)
	defer conn.Close()
	if conn.IsNil() {
		return errors.New("No Server Exit")
	}

	str := account2json.Imp(conn, accountid)

	c.SetData(string(str))
	return nil
}

func setAllInfoByAccountID(c *gm_command.Context, server, accountid string, params []string) error {
	cfg := gmConfig.Cfg.GetServerCfgFromName(server)

	logs.Info("setAllInfoByAccountID  %s %v --> %v", server, params, cfg)

	conn := getRedisConn(cfg.RedisName)
	defer conn.Close()
	if conn.IsNil() {
		return errors.New("No Server Exit")
	}
	res := json2account.Imp(conn, accountid, []byte(params[0]))

	if res != "" {
		c.SetData(res)
	} else {
		c.SetData("ok")
	}
	return nil
}

type AvatarEquip struct {
	Curr_equips  []uint32 `json:"curr"`
	Lv_upgrade   []uint32 `json:"up"`   // 装备强化等级
	Lv_evolution []uint32 `json:"ev"`   // 精炼等级
	Lv_star      []uint32 `json:"star"` // 装备星级
	Star_Bless   []uint32 `json:"sb"`   // 祝福值
}

type ProfileData struct {
	aEquip AvatarEquip

	AEquip string `json:"avatarEquip"`
}

func (p *ProfileData) IsHasEquip(ids string) bool {
	idn, err := strconv.Atoi(ids)

	if err != nil {
		return true
	}

	id := uint32(idn)

	if id == 0 || bag.IsFixedID(id) {
		return true
	}

	for _, eid := range p.aEquip.Curr_equips {
		if eid == id {
			return true
		}
	}
	return false
}

type BagData struct {
	old map[string]string
}

func (b *BagData) LoadData(strs map[string]string) error {
	b.old = strs
	return nil
}

func (b *BagData) CleanUnEquipItems(p *ProfileData) {
	newItems := make(map[string]string, len(b.old))
	for ids, eq := range b.old {
		if p.IsHasEquip(ids) {
			newItems[ids] = eq
		}
	}
	b.old = newItems
}

func (p *ProfileData) LoadEquip() error {
	err := json.Unmarshal([]byte(p.AEquip), &p.aEquip)
	if err != nil {
		logs.Error("Unmarshal Profile  Err By %s", err.Error())
		return err
	}
	logs.Trace("LoadEquip %v", p.aEquip)
	return nil
}

func (b *BagData) SaveData() ([]byte, error) {
	res, err := json.Marshal(b.old)
	if err != nil {
		return []byte{}, err
	}

	return res, nil
}

func cleanUnEquipItems(c *gm_command.Context, server, accountid string, params []string) error {
	cfg := gmConfig.Cfg.GetServerCfgFromName(server)

	data_ProfileData, err := GetProfileFromRedis(cfg.RedisName, "profile:"+accountid)
	if err != nil {
		logs.Error("GetProfileFromRedis profile_data Err By %s", err.Error())
		return err
	}

	data_BagData, err := GetProfileFromRedis(cfg.RedisName, "bag:"+accountid)
	if err != nil {
		logs.Error("GetProfileFromRedis profile_data Err By %s", err.Error())
		return err
	}

	logs.Trace("GetProfileFromRedis %v", string(data_ProfileData))
	logs.Trace("GetProfileFromRedis %v", string(data_BagData))

	bags := make(map[string]string, 640)
	err = json.Unmarshal(data_BagData, &bags)
	if err != nil {
		logs.Error("Unmarshal Bags  Err By %s", err.Error())
		return err
	}

	logs.Trace("Unmarshal Bags %v", bags)

	profiles := ProfileData{}
	bagDatas := BagData{}

	err = json.Unmarshal(data_ProfileData, &profiles)
	if err != nil {
		logs.Error("Unmarshal Profile  Err By %s", err.Error())
		return err
	}

	err = profiles.LoadEquip()
	if err != nil {
		logs.Error("LoadEquip Profile  Err By %s", err.Error())
		return err
	}

	err = bagDatas.LoadData(bags)
	if err != nil {
		logs.Error("LoadData Bag  Err By %s", err.Error())
		return err
	}

	// 清除所有没有装备的装备
	bagDatas.CleanUnEquipItems(&profiles)

	newBagDatas, err := bagDatas.SaveData()
	if err != nil {
		logs.Error("bagDatas SaveData  Err By %s", err.Error())
		return err
	}

	logs.Trace("datas %v", string(newBagDatas))

	err = ModPrfileFromRedis(cfg.RedisName, "bag:"+accountid, string(newBagDatas))
	if err != nil {
		return err
	}

	return nil
}
