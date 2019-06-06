package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	authConfig "vcs.taiyouxi.net/platform/x/auth/config"
	"vcs.taiyouxi.net/platform/x/cheat_beholder/accountinfo"
	CBConfig "vcs.taiyouxi.net/platform/x/cheat_beholder/config"
	"vcs.taiyouxi.net/platform/x/cheat_beholder/serverinfo"
	"vcs.taiyouxi.net/platform/x/gift_sender/config"
)

func main() {
	defer logs.Close()
	CBConfig.LoadConfig("conf/config.toml")
	config.CommonConfig.EtcdRoot = CBConfig.CommonConfig.Etcd_Root
	fmt.Printf("CBCEtcd_Root:%v\n", CBConfig.CommonConfig.Etcd_Root)
	fmt.Printf("configEtcd_Root:%v\n", config.CommonConfig.EtcdRoot)
	err := etcd.InitClient(CBConfig.CommonConfig.Etcd_Endpoint)
	if err != nil {
		logs.Error("etcd InitClient err %s", err.Error())
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}
	authConfig.Cfg.Runmode = "test"
	authConfig.Cfg.EtcdRoot = CBConfig.CommonConfig.Etcd_Root
	serverinfo.RegisterAllServerInfo()

	logs.Trace("[cyt]length of Servers:%v", len(serverinfo.RSP.Servers))
	for i, v := range serverinfo.RSP.Servers {
		logs.Trace("Server%d, data:%v", i, v)
	}
	/*file := xlsx.NewFile()
	sheet, err := file.AddSheet("CheatPlayers")
	if err != nil {
		logs.Error("为xls文件添加页签sheet出错%v", err)
	}
	row := sheet.AddRow()
	row.AddCell().SetValue("玩家昵称")
	row.AddCell().SetValue("AcID")
	row.AddCell().SetValue("玩家战力")
	row.AddCell().SetValue("最远关卡id")
	row.AddCell().SetValue("最远关卡推荐战力")
	row.AddCell().SetValue("VIP等级")*/

	file, err := os.Create("cheat_" + time.Now().Format("02-01-2006_15:04:05_PM") + ".csv") //创建文件
	if err != nil {
		logs.Error("%v", err)
		panic(err)
	}
	defer file.Close()

	file.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM

	w := csv.NewWriter(file) //创建一个新的写入文件流

	w.Write([]string{
		"玩家昵称",
		"AcID",
		"玩家战力",
		"最远关卡id",
		"最远关卡推荐战力",
		"VIP等级",
	}) //写入数据
	w.Flush()
	w = csv.NewWriter(file)
	defer func() {
		logs.Trace("[cyt]save file")
		defer w.Flush()
	}()
	for _, v := range serverinfo.RSP.Servers {
		accountinfo.Iterator(v.Serid, w)
	}

}
