package sys_public

import (
	"encoding/json"
	"strconv"

	"fmt"

	"os/exec"

	"time"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/secure"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	"vcs.taiyouxi.net/platform/x/gm_tools/config"
)

func RegCommands() {
	gm_command.AddGmCommandHandle("getSysPublic", getSysPublic)
	gm_command.AddGmCommandHandle("sendSysPublic", sendSysPublic)
	gm_command.AddGmCommandHandle("delSysPublic", delSysPublic)
	gm_command.AddGmCommandHandle("releaseSysPublic", releaseSysPublic)
	gm_command.AddGmCommandHandle("releaseSysPublicToRelease", releaseSysPublic)
	gm_command.AddGmCommandHandle("getEndpoint", getEndpoint)
	gm_command.AddGmCommandHandle("setEndpoint", setEndpoint)
	gm_command.AddGmCommandHandle("uploadKS3", uploadKS3)
}

const encodePwd = "c2452p-3Iatj_gncV0sad0f7m02mc90420lposeakgfopai213uCLhmf80234l0i"
const defaultSalt = "34rfdsf3rdkunj;l9"

var DefaultEncode *secure.SecureEncode

func init() {
	DefaultEncode = secure.New(encodePwd, defaultSalt)
}

func getSysPublic(c *gm_command.Context, server, accountid string, params []string) error {
	getSysPublicTime := time.Now()
	res := sysPublicManager.GetAll(params[0])
	c.SetData(string(res))
	logs.Debug("loadSysPublic used %v time",time.Since(getSysPublicTime))
	return nil
}

func sendSysPublic(c *gm_command.Context, server, accountid string, params []string) error {
	sendSysPublicTime := time.Now()
	gid, err := strconv.ParseInt(params[0], 10, 64)
	if err != nil {
		logs.Error("gid sendSysPublic err : %v", err)
		return err
	}
	p, err := strconv.ParseInt(params[1], 10, 64)
	if err != nil {
		logs.Error("gid sendSysPublic err : %v", err)
		return err
	}
	typ, err := strconv.ParseInt(params[2], 10, 64)
	if err != nil {
		logs.Error("gid sendSysPublic err : %v", err)
		return err
	}
	is_sended, err := strconv.ParseInt(params[5], 10, 64)
	if err != nil {
		logs.Error("gid sendSysPublic err : %v", err)
		return err
	}
	logs.Debug("3:%v,4:%v", params[3], params[4])

	tb_Time, err := time.ParseInLocation("2006/01/02 15:04:05", params[3], util.ServerTimeLocal)
	if err != nil {
		logs.Error("tb_Time sendSysPublic err : %v", err)
		return err
	}
	te_Time, err := time.ParseInLocation("2006/01/02 15:04:05", params[4], util.ServerTimeLocal)
	if err != nil {
		logs.Error("te_Time sendSysPublic err : %v", err)
		return err
	}
	tb := tb_Time.Unix()
	te := te_Time.Unix()
	logs.Trace("tb:对应的时间为：%v te:对应的时间为:%v", tb_Time.Format("2006/01/02 15:04:05"), te_Time.Format("2006/01/02 15:04:05"))
	logs.Trace("tb:%v te:%v", tb, te)

	s := NewSysPublic(gid, p, typ, tb, te, is_sended, params[6], params[7], params[8], params[9], params[10], params[11])
	logs.Debug("sendSysPublic read Params using %v time",time.Since(sendSysPublicTime))

	sendSysPublicTime = time.Now()

	logs.Trace("sendSysPublic %s", accountid)

	// 修改逻辑使用
	if accountid != "null" {
		s.Id = accountid
	}

	logs.Trace("sendSysPublic %v", s)

	sysPublicManager.Add(s)

	logs.Debug("sendSysPublic Add NewSysPublic using %v time",time.Since(sendSysPublicTime))
	return nil
}

func delSysPublic(c *gm_command.Context, server, accountid string, params []string) error {
	sysPublicManager.Del(params[0])
	return nil
}

func releaseSysPublic(c *gm_command.Context, server, accountid string, params []string) error {
	gid, err := strconv.ParseInt(params[0], 10, 64)
	if err != nil {
		return err
	}
	return sysPublicManager.SendOnePublic(false, gid, params[1])
}

func releaseSysPublicToRelease(c *gm_command.Context, server, accountid string, params []string) error {
	gid, err := strconv.ParseInt(params[0], 10, 64)
	if err != nil {
		return err
	}
	return sysPublicManager.SendOnePublic(false, gid, params[1])
}

func getEndpoint(c *gm_command.Context, server, accountid string, params []string) error {
	gid, err := strconv.ParseInt(params[0], 10, 64)
	if err != nil && params[0] != "" {
		return err
	}
	ver := params[1]
	res := sysPublicManager.GetEndpoint(gid, ver)
	if res == nil {
		res = &sysEndPoints{}
	}
	res.DataVerEtcd = _getDataVerFromEtcd(gid, ver)
	res.BundleVerEtcd = _getBundleVerFromEtcd(gid, ver)
	res.DataMinEtcd = _getDataMinFromEtcd(gid, ver)
	res.BundleMinEtcd = _getBundleMinFromEtcd(gid, ver)
	j, err := json.Marshal(*res)
	if err != nil {
		return err
	}
	c.SetData(string(j))
	logs.Debug("getEndpoint %s ", string(j))
	return nil
}

func setEndpoint(c *gm_command.Context, server, accountid string, params []string) error {
	gid, err := strconv.ParseInt(params[0], 10, 64)
	if err != nil {
		return err
	}
	s := sysEndPoints{}
	s.ServerEndpoint = params[2]
	s.ChatEndpoint = params[3]
	s.DataVer = params[4]
	s.Whitelistpwd = params[5]
	s.BIOption = params[6]
	s.TokenTime = params[7]
	s.TimeoutTime = params[8]
	s.BundleVer = params[9]
	s.BundleVerEtcd = params[10]
	s.PrUrl = params[11]
	s.PayUrls = params[12]
	s.GZIPSize = params[13]
	s.DataMin = params[14]
	s.BundleMin = params[15]
	if s.GZIPSize == "" {
		s.GZIPSize = "1024"
	}
	logs.Debug("EndPoint Param: %v, %s", params, params[13])
	ver := params[1]
	sysPublicManager.SetEndpoint(gid, ver, s)
	if err := setDataVerFromEtcd(gid, ver, s.DataVer); err != nil {
		return err
	}
	if err := setDataMinFromEtcd(gid, ver, s.DataMin); err != nil {
		return err
	}
	if err := setBundleVerFromEtcd(gid, ver, s.BundleVer); err != nil {
		return err
	}
	if err := setBundleMinFromEtcd(gid, ver, s.BundleMin); err != nil {
		return err
	}
	return nil
}

func _getDataVerFromEtcd(gid int64, ver string) string {
	key := fmt.Sprintf("%s/%d/%s/%s", config.Cfg.GameSerEtcdRoot, gid, etcd.KeyGlobalClientDataVer, ver)
	da, err := etcd.Get(key)
	if err != nil {
		return ""
	}
	return da
}

func _getDataMinFromEtcd(gid int64, ver string) string {
	key := fmt.Sprintf("%s/%d/%s/%s", config.Cfg.GameSerEtcdRoot, gid, etcd.KeyGlobalClientDataVerMIn, ver)
	da, err := etcd.Get(key)
	if err != nil {
		return ""
	}
	return da
}

func setDataVerFromEtcd(gid int64, ver, data_ver string) error {
	if ver == "" { // 若版本号没有，则什么也不做
		return nil
	}
	if data_ver != "" {
		if _, err := strconv.Atoi(data_ver); err != nil {
			return err
		}
	}
	key := fmt.Sprintf("%s/%d/%s/%s", config.Cfg.GameSerEtcdRoot, gid, etcd.KeyGlobalClientDataVer, ver)
	return etcd.Set(key, data_ver, 0)
}

func setDataMinFromEtcd(gid int64, ver, data_min string) error {
	if ver == "" { // 若版本号没有，则什么也不做
		return nil
	}
	if data_min != "" {
		if _, err := strconv.Atoi(data_min); err != nil {
			return err
		}
	}
	key := fmt.Sprintf("%s/%d/%s/%s", config.Cfg.GameSerEtcdRoot, gid, etcd.KeyGlobalClientDataVerMIn, ver)
	return etcd.Set(key, data_min, 0)
}

func _getBundleVerFromEtcd(gid int64, ver string) string {
	key := fmt.Sprintf("%s/%d/%s/%s", config.Cfg.GameSerEtcdRoot, gid, etcd.KeyGlobalClientBundleVer, ver)
	da, err := etcd.Get(key)
	logs.Debug("_getBundleVerFromEtcd key %s %s", key, da)
	if err != nil {
		return ""
	}
	return da
}

func _getBundleMinFromEtcd(gid int64, ver string) string {
	key := fmt.Sprintf("%s/%d/%s/%s", config.Cfg.GameSerEtcdRoot, gid, etcd.KeyGlobalClientBundleVerMin, ver)
	da, err := etcd.Get(key)
	logs.Debug("_getBundleMinFromEtcd key %s %s", key, da)
	if err != nil {
		return ""
	}
	return da
}

func setBundleVerFromEtcd(gid int64, ver, data_ver string) error {
	if ver == "" { // 若版本号没有，则什么也不做
		return nil
	}
	if data_ver != "" {
		if _, err := strconv.Atoi(data_ver); err != nil {
			return err
		}
	}
	key := fmt.Sprintf("%s/%d/%s/%s", config.Cfg.GameSerEtcdRoot, gid, etcd.KeyGlobalClientBundleVer, ver)
	logs.Debug("setBundleVerFromEtcd key %s %s", key, data_ver)
	return etcd.Set(key, data_ver, 0)
}

func setBundleMinFromEtcd(gid int64, ver, bundle_min string) error {
	if ver == "" { // 若版本号没有，则什么也不做
		return nil
	}
	if bundle_min != "" {
		if _, err := strconv.Atoi(bundle_min); err != nil {
			return err
		}
	}
	key := fmt.Sprintf("%s/%d/%s/%s", config.Cfg.GameSerEtcdRoot, gid, etcd.KeyGlobalClientBundleVerMin, ver)
	logs.Debug("setBundleVerFromEtcd key %s %s", key, bundle_min)
	return etcd.Set(key, bundle_min, 0)
}

func uploadKS3(c *gm_command.Context, server, accountid string, params []string) error {
	//cmd := "/usr/bin/java -jar ./confd/ks3up-2.0.2.jar -c ./jinshan_CDN.conf start"
	cmdRes := exec.Command("sh", "upload_CDN.sh")
	res, err := cmdRes.Output()
	if err != nil {
		logs.Error("uploadKS3 err", err)
		return err
	} else {
		logs.Info(string(res))
	}
	return nil
}
