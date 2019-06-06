﻿#极无双英雄志H5数据库项目文档
##数据需求
通过H5查看自己的数据，吸引玩家转发。H5画面通过白描上色展现三国典故并将具体故事隐藏在画面你点击可看。
输入信息页：输入区服名称、角色名称加logo
###封面：
文案：极无双英雄志  古往今来英雄胆，留有青史一段香（小）
###第一页：
文案：  露锋芒  英雄露颖在今朝，一试矛兮一试刀。（小）
<font color=ff0000 face="黑体">二零一六年五月一日</font>初入无双乱世之中，<font color=ff0000 face="黑体">将军</font>独行斩将，在<font color=ff0000 face="黑体">群雄逐鹿</font>服威名颇震。
画面：一武将马上扬刀对敌，文字竖着排版
###第二页：
文案：  遣良将 总揽良将分寰宇，思贤若渴真英主。（小）
后来将军四处奔走寻名将，<font color=ff0000 face="黑体">关平</font>首先来投，<font color=ff0000 face="黑体">赵云</font>常伴左右，共集良将<font color=ff0000 face="黑体">23</font>名
###第三页：
文案：  聚军众 四方将士皆来聚，正如龙虎风云会。（小）
几经辗转，将军加入<font color=ff0000 face="黑体">风云会</font>军团，虽临刀光剑影但重手足情深。
###第四页：
文案：  图霸业 气挟风雷无匹敌,志垂日月有光芒。（小）
一路过关斩将，将军等级终达<font color=ff0000 face="黑体">52级</font>，累计经历了<font color=ff0000 face="黑体">78</font>战役，霸业将成。
###第五页：
文案：  扬天下  纵横沙场破奇关，英雄从此震江山。（小）
将军斩杀敌人千万，将军等级终达<font color=ff0000 face="黑体">个人竞技排行榜第5名，跨服争霸排行榜第6名</font>，名扬天下。
###封底：
落版，极无双：这天下由我主宰。下载游戏或查看我的英雄志
userinfo&state=163#wechat_redirect

##完整代码存储在附件中
##需求说明
接口需要传递玩家的建号时间、昵称、服务器名、第一个武将（固定关平）、战力最高的武将（如果是关平则提供战斗力第二高的武将）、武将数量、工会名、玩家等级、战役数量、个人竞技排行榜和跨服争霸排行榜信息。
##代码：
+ 通过初始化函数加载data文件（包含玩家个人信息等数据），调用获取角色信息函数getRoleInfo
```
func Init(e *gin.Engine) {
   config.CommonConfig.EtcdRoot = h5Config.CommonConfig.Etcd_Root
   //加载data文件
   dataRelPath := "data"
   dataAbsPath := filepath.Join(getMyDataPath(), dataRelPath)
   load := func(dfilepath string, loadfunc func(string)) {
      //logs.Info("LoadGameData %s start", filepath)
      loadfunc(filepath.Join("", dataAbsPath, dfilepath))
      logs.Trace("LoadGameData %s success", dfilepath)
   }
   load("level_info.data", loadStageData)
   //加载之后，全局变量gdStageData中为全部已通关关卡的详细信息
   //全局变量gdStagePurchasePolicy中为所有已通关关卡的关卡重置次数购买规则
   //全局变量gdStageTimeLimit中为限时
   load("severgroup.data", loadServeGroup)
   e.GET("/getRoleInfo", getRoleInfo)
}
```
+ getRoleInfo函数获取角色信息，ServerID从外部传入进来，此函数先后调用getACID(获取玩家ACID)和getInfo(获取玩家详细信息)、如果出现问题，返回JSON202、201，其Result为-1
如果正确则返回Json200，其Result为1，info中为玩家详细信息。（info是一个结构体实例，结构稍后介绍）
```

func getRoleInfo(c *gin.Context) {
   _numberID := c.Query("id")

   serverID := c.Query("server_id")

   logs.Debug("id为", _numberID)
   logs.Debug("server_id为", serverID)
   g_sid := strings.Split(serverID, ":")

   if g_sid[0] == "" {
      logs.Debug("putin wrong gid")
      c.JSON(202, accountInfo{Result: -1})
      return
   }
   gid, err := strconv.Atoi(g_sid[0])
   if err != nil {
      logs.Error("putin wrong gid error", err)
   }
   if g_sid[1] == "" {
      logs.Debug("putin wrong sid")
      c.JSON(202, accountInfo{Result: -1})
      return
   }
   sid, err := strconv.Atoi(g_sid[1])
   if err != nil {
      logs.Error("putin wrong sid error", err)
   }
   if _numberID == "" {
      logs.Debug("putin wrong id")
      c.JSON(202, accountInfo{Result: -1})
      return
   }
   numberID, err := strconv.Atoi(_numberID)
   if err != nil {
      logs.Error("putin wrong id error", err)
   }
   acid, err := getACID(numberID, gid, sid)
   if err != nil {
      logs.Error("get ACID err by: %v", err)
      c.JSON(202, accountInfo{Result: -1})
      return
   }
   info, err := getInfo(acid, gid, sid)
   if err != nil {
      logs.Error("get Info err by: %v", err)
      c.JSON(201, accountInfo{Result: -1})
      return
   }
   logs.Debug("%v", info)
   info.Result = 1
   c.JSON(200, info)
}
```
+ 玩家信息结构体accontInfo，最后一个Result属性表示获取的成功/失败，成功时Result为0，失败时为-1
```
type accountInfo struct {
   //建号时间
   CreateTime string `json:"CreateTime"`
   //玩家姓名
   Name string `json:"name"`
   //服务器名字
   SeverName string `json:"severname"`
   //第一个英雄
   FirstHero string `json:"firstHero"`
   //战力最高的英雄
   BestHero string `json:"bestHero"`
   //英雄数量
   HeroNum uint `json:"heroNum"`
   //工会名字
   Gname string `json:"gname"`
   //玩家等级
   Level uint `json:"level"`
   //玩家战役数量
   BattleNum uint `json:"battleNum"`
   //个人排行榜排名
   PvpRank int `json:"pvpRank"`
   //跨服排行榜排名
   A_pvpRank int `json:"A_pvpRank"`
   Result    int `json:"result"`
}
```
+ getACID会获取玩家的账户信息
这样设计是为了应对查找不同信息时不同的玩家id号
```
//获取玩家的ACID,若未找到则报错
func getACID(numberID, gid, sid int) (string, error) {
   acID, err := getAcIDFromNumberID(numberID, gid, sid)
   if err != nil {
      return "", err
   }
   return acID, nil
}
```
+ getInfo获取玩家详细信息，结果存储在一个accountInfo结构体内，当出现不符合道理的数据（理论上不可能，例如没有建号时间，没有等级等应有的信息）时提前终止，返回err，err为未找到的信息详情
工会名称、个人排行榜、争霸排行榜可能为空，也就是玩家没有加入工会、没有打个人竞技、跨服争霸，如果未查找到值，则工会名称为空串、个人排行榜为-1、跨服争霸排行榜为-1。
获取战力最高武将时，如果战力最高的武将为关平，则替换为战斗力第二高的武将。
```
//获取信息，通过acid将玩家所有信息查找出来，放在结构体内
func getInfo(acid string, gid, sid int) (accountInfo, error) {
   profileID := "profile:" + acid
   var result accountInfo = accountInfo{}
   //链接玩家个人信息数据库db
   conn1, err := core.GetGamexProfileReidsConn(uint(gid), uint(sid))
   defer func() {
      if conn1 != nil {
         conn1.Close()
      }
   }()
   if err != nil {
      return result, err
   }
   //链接玩家排行榜信息数据库db
   conn5, err := core.GetGamexRankReidsConn(uint(gid), uint(sid))
   defer func() {
      if conn5 != nil {
         conn5.Close()
      }
   }()
   if err != nil {
      return result, err
   }
   //获取建号时间
   timestamp, err := redis.Int64(conn1.Do("HGET", profileID, "createtime"))
   if err != nil {
      return result, err
   }
   result.CreateTime = getChineseDate(time.Unix(timestamp, 0).Format("2006-01-02"))

   //玩家昵称
   name, err := redis.String(conn1.Do("HGET", profileID, "name"))
   if err != nil {
      return result, err
   }
   result.Name = name
   //服务器名
   severname, err := getEtcdCfg1(uint(gid), uint(sid))
   if err != nil {
      return result, err
   } else {
      result.SeverName = severname
   }
   //第一个英雄
   result.FirstHero = "关平"
   //战力最高英雄
   besthero, err := conn1.Do("HGET", profileID, "data")
   if err != nil {
      return result, err
   } else if besthero != nil {
      var herodata account.ProfileData
      err := json.Unmarshal(besthero.([]byte), &herodata)
      if err != nil {
         return result, err
      } else {
         max_gs := 0     //最高战斗力武将的战斗力
         index_gs := 0   //最高战斗力武将id
         s_index_gs := 0 //第二高战斗力武将的id
         for i := 0; i < helper.AVATAR_NUM_MAX; i++ {
            if herodata.HeroGs[i] > max_gs {
               s_index_gs = index_gs
               max_gs = herodata.HeroGs[i]
               index_gs = i
            }
         }
         if index_gs == serverinfo.HERO_GY { //如果战斗力最高的英雄为关平，则输出第二战斗力武将
            index_gs = s_index_gs

         }

         result.BestHero = serverinfo.GetHeroName(index_gs)
      }
   }
   //总英雄数量
   heros, err := conn1.Do("HGET", profileID, "hero")
   if err != nil {
      return result, err
   } else if heros != nil {

      //武将数量
      var herodata account.PlayerHero

      err := json.Unmarshal(heros.([]byte), &herodata)
      if err != nil {
         return result, err
      } else {
         heronum := 0
         for i := 0; i < helper.AVATAR_NUM_MAX; i++ {
            if herodata.HeroExp[i] > 0 {
               heronum++
            }
         }

         result.HeroNum = uint(heronum)
      }
   }
   //工会名
   gname, err := redis.String(conn1.Do("HGET", profileID, "gname"))
   if err != nil {
      result.Gname = ""
   } else {
      result.Gname = gname
   }
   //玩家等级

   corpdata, err := conn1.Do("HGET", profileID, "corp")
   if err != nil {
      return result, err
   } else {
      var corp account.Corp
      err = json.Unmarshal(corpdata.([]byte), &corp)
      if err != nil {
         return result, err
      }

      result.Level = uint(corp.Level)
   }

   //计算游戏天数
   now := time.Now().Unix()
   ct := timestamp
   if err != nil {
      return result, err
   }
   pro_days := (now-int64(ct))/(24*3600) + 1
   //玩家vip信息
   vipdata, err := conn1.Do("HGET", profileID, "v")
   if err != nil {
      return result, err
   }
   var vip account.VIP
   err = json.Unmarshal(vipdata.([]byte), &vip)

   if err != nil {
      return result, err
   } else {
      //计算通关数量
      valueVip := 0
      switch vip.V {
      case 0:
         valueVip = releation0
      case 1:
         valueVip = releation1
      case 2:
         valueVip = releation2
      case 3:
         valueVip = releation3
      case 4:
         valueVip = releation4
      case 5:
         valueVip = releation5
      case 6:
         valueVip = releation6
      case 7:
         valueVip = releation7
      case 8:
         valueVip = releation8
      case 9:
         valueVip = releation9
      case 10:
         valueVip = releation10
      case 11:
         valueVip = releation11
      case 12:
         valueVip = releation12
      case 13:
         valueVip = releation13
      case 14:
         valueVip = releation14
      case 15:
         valueVip = releation15
      case 16:
         valueVip = releation16
      }
      result.BattleNum = uint(pro_days * int64(valueVip))
   }
   //个人竞技排行榜

   pvprank, err := redis.Int(conn5.Do("zrevrank", fmt.Sprintf("%d:%d:RankSimplePvp", gid, sid), acid))
   logs.Debug("info: %v, %v", fmt.Sprintf("%d:%d", gid, sid)+":RankSimplePvp", acid)
   pvprank++
   if err != nil {
      logs.Error("get pvprank err by %v", err)
      result.PvpRank = -1
   } else {
      result.PvpRank = pvprank
   }

   result.A_pvpRank, err = getwspvpinfo(acid, sid, gid)
   logs.Error("跨服排行榜获取出错", err)
   logs.Debug("map长度为：", len(gdServerGroupSbatch))
   return result, nil

}

func getwspvpinfo(acid string, sid, gid int) (int, error) {
   //跨服争霸排行榜
   groupString := (fmt.Sprintf("%d", gdServerGroupSbatch[uint32(sid)].WspvpGroupID))
   ws_pvpmap, err := getEtcdCfg(uint(gid))
   if err != nil {
      logs.Error("跨服排行榜获取出错1", err)
      return -1, err
   }
   logs.Debug("ws_pvpmap_length,groupString, %d,%s", len(ws_pvpmap), groupString)
   for k, v := range ws_pvpmap {
      logs.Debug("ws_pvpmap,key:%v,value:%v", k, v)
   }
   wsConfig, ok := ws_pvpmap[groupString]
   logs.Debug("ok:", ok)
   if !ok {
      wsConfig = ws_pvpmap["default"]
   }
   logs.Debug("DB", wsConfig.DB)
   connx, err := getRedisConn(wsConfig.AddrPort, wsConfig.DB, wsConfig.Auth)
   defer func() {
      if connx != nil {
         connx.Close()
      }
   }()
   if err != nil {
      logs.Error("跨服排行榜获取出错2", err)
      return -1, err
   }
   logs.Debug("groupString:", groupString)
   A_rankpvp, err := redis.Int(connx.Do("zrank", "rank:"+groupString, acid))
   if err != nil {
      logs.Error("跨服排行榜获取出错3", err)
      return -1, err
   } else {
      return A_rankpvp + 1, nil
   }
}
//获取玩家的ACID,若未找到则报错
func getACID(numberID, gid, sid int) (string, error) {
   acID, err := getAcIDFromNumberID(numberID, gid, sid)
   if err != nil {
      return "", err
   }
   return acID, nil
}

//获取信息，通过acid将玩家所有信息查找出来，放在结构体内
func getInfo(acid string, gid, sid int) (accountInfo, error) {
   profileID := "profile:" + acid
   var result accountInfo = accountInfo{}
   //链接玩家个人信息数据库db
   conn1, err := core.GetGamexProfileReidsConn(uint(gid), uint(sid))
   defer func() {
      if conn1 != nil {
         conn1.Close()
      }
   }()
   if err != nil {
      return result, err
   }
   //链接玩家排行榜信息数据库db
   conn5, err := core.GetGamexRankReidsConn(uint(gid), uint(sid))
   defer func() {
      if conn5 != nil {
         conn5.Close()
      }
   }()
   if err != nil {
      return result, err
   }
   //获取建号时间
   timestamp, err := redis.Int64(conn1.Do("HGET", profileID, "createtime"))
   if err != nil {
      return result, err
   }
   result.CreateTime = getChineseDate(time.Unix(timestamp, 0).Format("2006-01-02"))

   //玩家昵称
   name, err := redis.String(conn1.Do("HGET", profileID, "name"))
   if err != nil {
      return result, err
   }
   result.Name = name
   //服务器名
   severname, err := getEtcdCfg1(uint(gid), uint(sid))
   if err != nil {
      return result, err
   } else {
      result.SeverName = severname
   }
   //第一个英雄
   result.FirstHero = "关平"
   //战力最高英雄
   besthero, err := conn1.Do("HGET", profileID, "data")
   if err != nil {
      return result, err
   } else if besthero != nil {
      var herodata account.ProfileData
      err := json.Unmarshal(besthero.([]byte), &herodata)
      if err != nil {
         return result, err
      } else {
         max_gs := 0     //最高战斗力武将的战斗力
         index_gs := 0   //最高战斗力武将id
         s_index_gs := 0 //第二高战斗力武将的id
         for i := 0; i < helper.AVATAR_NUM_MAX; i++ {
            if herodata.HeroGs[i] > max_gs {
               s_index_gs = index_gs
               max_gs = herodata.HeroGs[i]
               index_gs = i
            }
         }
         if index_gs == serverinfo.HERO_GY { //如果战斗力最高的英雄为关平，则输出第二战斗力武将
            index_gs = s_index_gs

         }

         result.BestHero = serverinfo.GetHeroName(index_gs)
      }
   }
   //总英雄数量
   heros, err := conn1.Do("HGET", profileID, "hero")
   if err != nil {
      return result, err
   } else if heros != nil {

      //武将数量
      var herodata account.PlayerHero

      err := json.Unmarshal(heros.([]byte), &herodata)
      if err != nil {
         return result, err
      } else {
         heronum := 0
         for i := 0; i < helper.AVATAR_NUM_MAX; i++ {
            if herodata.HeroExp[i] > 0 {
               heronum++
            }
         }

         result.HeroNum = uint(heronum)
      }
   }
   //工会名
   gname, err := redis.String(conn1.Do("HGET", profileID, "gname"))
   if err != nil {
      result.Gname = ""
   } else {
      result.Gname = gname
   }
   //玩家等级

corpdata, err := conn1.Do("HGET", profileID, "corp")
if err != nil {
   return result, err
} else {
   var corp account.Corp
   err = json.Unmarshal(corpdata.([]byte), &corp)
   if err != nil {
      return result, err
   }

   result.Level = uint(corp.Level)
}

   //计算游戏天数
   now := time.Now().Unix()
   ct := timestamp
   if err != nil {
      return result, err
   }
   pro_days := (now-int64(ct))/(24*3600) + 1
   //玩家vip信息
   vipdata, err := conn1.Do("HGET", profileID, "v")
   if err != nil {
      return result, err
   }
   var vip account.VIP
   err = json.Unmarshal(vipdata.([]byte), &vip)

   if err != nil {
      return result, err
   } else {
      //计算通关数量
      valueVip := 0
      switch vip.V {
      case 0:
         valueVip = releation0
      case 1:
         valueVip = releation1
      case 2:
         valueVip = releation2
      case 3:
         valueVip = releation3
      case 4:
         valueVip = releation4
      case 5:
         valueVip = releation5
      case 6:
         valueVip = releation6
      case 7:
         valueVip = releation7
      case 8:
         valueVip = releation8
      case 9:
         valueVip = releation9
      case 10:
         valueVip = releation10
      case 11:
         valueVip = releation11
      case 12:
         valueVip = releation12
      case 13:
         valueVip = releation13
      case 14:
         valueVip = releation14
      case 15:
         valueVip = releation15
      case 16:
         valueVip = releation16
      }
      result.BattleNum = uint(pro_days * int64(valueVip))
   }
   //个人竞技排行榜

   pvprank, err := redis.Int(conn5.Do("zrevrank", fmt.Sprintf("%d:%d:RankSimplePvp", gid, sid), acid))
   logs.Debug("info: %v, %v", fmt.Sprintf("%d:%d", gid, sid)+":RankSimplePvp", acid)
   pvprank++
   if err != nil {
      logs.Error("get pvprank err by %v", err)
      result.PvpRank = -1
   } else {
      result.PvpRank = pvprank
   }

   result.A_pvpRank, err = getwspvpinfo(acid, sid, gid)
   logs.Error("跨服排行榜获取出错", err)
   logs.Debug("map长度为：", len(gdServerGroupSbatch))
   return result, nil

}

func getwspvpinfo(acid string, sid, gid int) (int, error) {
   //跨服争霸排行榜
   groupString := (fmt.Sprintf("%d", gdServerGroupSbatch[uint32(sid)].WspvpGroupID))
   ws_pvpmap, err := getEtcdCfg(uint(gid))
   if err != nil {
      logs.Error("跨服排行榜获取出错1", err)
      return -1, err
   }
   logs.Debug("ws_pvpmap_length,groupString, %d,%s", len(ws_pvpmap), groupString)
   for k, v := range ws_pvpmap {
      logs.Debug("ws_pvpmap,key:%v,value:%v", k, v)
   }
   wsConfig, ok := ws_pvpmap[groupString]
   logs.Debug("ok:", ok)
   if !ok {
      wsConfig = ws_pvpmap["default"]
   }
   logs.Debug("DB", wsConfig.DB)
   connx, err := getRedisConn(wsConfig.AddrPort, wsConfig.DB, wsConfig.Auth)
   defer func() {
      if connx != nil {
         connx.Close()
      }
   }()
   if err != nil {
      logs.Error("跨服排行榜获取出错2", err)
      return -1, err
   }
   logs.Debug("groupString:", groupString)
   A_rankpvp, err := redis.Int(connx.Do("zrank", "rank:"+groupString, acid))
   if err != nil {
      logs.Error("跨服排行榜获取出错3", err)
      return -1, err
   } else {
      return A_rankpvp + 1, nil
   }
}
```
+ 根据策划提供的数据，玩家通关数量等于通关系数*游戏天数