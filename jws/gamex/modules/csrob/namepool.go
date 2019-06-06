package csrob

import (
	"fmt"
	"strings"

	"time"

	"vcs.taiyouxi.net/jws/gamex/modules/csrob/safetable"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	poolNameExpire = 10 * time.Minute
)

type poolName struct {
	res *resources

	playerTable *safetable.SafeTable
	guildTable  *safetable.SafeTable
	serverTable *safetable.SafeTable
}

func newPoolName(res *resources) *poolName {
	p := &poolName{}
	p.res = res
	p.playerTable = safetable.NewSafeTable(128)
	p.guildTable = safetable.NewSafeTable(128)
	p.serverTable = safetable.NewSafeTable(8)
	return p
}

//GetPlayerCache 取玩家缓存信息
func (p *poolName) GetPlayerCache(acid string) NamePoolPlayer {
	ret := p.playerTable.Get(acid)
	if nil == ret {
		ret, err := p.res.NamePoolDB.getPlayer(acid)
		if nil != err {
			logs.Error(fmt.Sprint(err))
			return NamePoolPlayer{}
		}
		if nil == ret {
			return NamePoolPlayer{}
		}
		p.playerTable.Set(acid, *ret)
		//10分钟后删除, 用来触发下次从DB取新的
		time.AfterFunc(
			poolNameExpire,
			func() {
				p.playerTable.Del(acid)
			},
		)
		return *ret
	}
	return ret.(NamePoolPlayer)
}

//SetPlayerCache 设置玩家缓存信息
func (p *poolName) SetPlayerCache(acid string, cache NamePoolPlayer) {
	p.playerTable.Set(acid, cache)
	if err := p.res.NamePoolDB.pushPlayer(&cache); nil != err {
		logs.Error(fmt.Sprint(err))
	}
}

//GetPlayerName 取玩家缓存的名字
func (p *poolName) GetPlayerName(acid string) string {
	return p.GetPlayerCache(acid).Name
}

//GetPlayerGuildID 取玩家缓存的公会ID
func (p *poolName) GetPlayerGuildID(acid string) string {
	return p.GetPlayerCache(acid).GuildID
}

//GetPlayerGuildPos 取玩家缓存的公会Pos
func (p *poolName) GetPlayerGuildPos(acid string) int {
	return p.GetPlayerCache(acid).GuildPos
}

//GetPlayerCSName 取玩家的跨服姓名
func (p *poolName) GetPlayerCSName(acid string) string {
	cache := p.GetPlayerCache(acid)
	return p.GetServerCache(cache.Sid).ServerName + "-" + cache.Name
}

//GetGuildCache 取公会缓存
func (p *poolName) GetGuildCache(guildID string) NamePoolGuild {
	ret := p.guildTable.Get(guildID)
	if nil == ret {
		ret, err := p.res.NamePoolDB.getGuild(guildID)
		if nil != err {
			logs.Error(fmt.Sprint(err))
			return NamePoolGuild{}
		}
		if nil == ret {
			return NamePoolGuild{}
		}
		p.guildTable.Set(guildID, *ret)
		//10分钟后删除, 用来触发下次从DB取新的
		time.AfterFunc(
			poolNameExpire,
			func() {
				p.guildTable.Del(guildID)
			},
		)
		return *ret
	}
	return ret.(NamePoolGuild)
}

//SetGuildCache 设置公会缓存
func (p *poolName) SetGuildCache(guildID string, cache NamePoolGuild) {
	p.guildTable.Set(guildID, cache)
	if err := p.res.NamePoolDB.pushGuild(&cache); nil != err {
		logs.Error(fmt.Sprint(err))
	}
}

//GetGuildCSName 取玩家的跨服姓名
func (p *poolName) GetGuildCSName(guildID string) string {
	cache := p.GetGuildCache(guildID)
	return p.GetServerCache(cache.Sid).ServerName + "-" + cache.GuildName
}

//GetServerCache 取服务器信息缓存, 数据来源为etcd表单, 不存数据库
func (p *poolName) GetServerCache(sid uint) NamePoolServer {
	key := fmt.Sprint(sid)

	ret := p.serverTable.Get(key)
	if nil == ret {
		cache := p.getServerFromEtcd(sid)
		p.serverTable.Set(key, cache)
		return cache
	}
	return ret.(NamePoolServer)
}

func (p *poolName) getServerFromEtcd(sid uint) NamePoolServer {
	displayName1 := etcd.GetSidDisplayName(game.Cfg.EtcdRoot, uint(game.Cfg.Gid), sid)
	logs.Info("[CSRob] poolName getServerFromEtcd, gid:%d, sid:%d, name:%s", game.Cfg.Gid, sid, displayName1)
	displayName2 := etcd.ParseDisplayShardName(displayName1)
	index := strings.Index(displayName2, "-")

	serverName := ""
	if index == -1 {
		serverName = displayName2
	} else {
		serverName = displayName2[:index]
	}

	return NamePoolServer{
		Sid:        sid,
		ServerName: serverName,
	}
}
