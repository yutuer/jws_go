package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"io"
	"os"
	"strconv"
	"time"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

var ivyMap map[string]guild_info.GuildInventory

// oldId -> newId
var lootIdMap = map[string]string{
	"GVE_BOSS_JD":           "LostGood_JD",
	"GVE_BOSS_ZL1":          "LostGood_ZL1",
	"GVE_BOSS_ZL2":          "LostGood_ZL2",
	"GVG_REVENUE_GIFT_1_1":  "LostGood_1_1",
	"GVG_REVENUE_GIFT_3_1":  "LostGood_3_1",
	"GVG_REVENUE_GIFT_4_1":  "LostGood_4_1",
	"GVG_REVENUE_GIFT_6_1":  "LostGood_6_1",
	"GVG_REVENUE_GIFT_7_1":  "LostGood_7_1",
	"GVG_REVENUE_GIFT_8_1":  "LostGood_8_1",
	"GVG_REVENUE_GIFT_9_1":  "LostGood_9_1",
	"GVG_REVENUE_GIFT_10_1": "LostGood_10_1",
}

func main() {
	app := cli.NewApp()

	app.Version = "0.0.0"
	app.Name = "SaveInventoryTo310"
	app.Usage = "save inventory from 240 to 310"
	app.Author = "libingbing"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "gid",
			Value: "",
			Usage: "大区",
		},
	}

	app.Action = func(context *cli.Context) {
		serverCfgs := loadServerCfg()
		//fmt.Println(serverCfgs)
		fileName := fmt.Sprintf("lost_inventory_%s.txt", context.String("gid"))
		ivyMap = loadOldInventory(fileName)
		for _, cfg := range serverCfgs {
			fmt.Println("start process server ", cfg.Gid, ":", cfg.Sid)
			saveIvy(cfg)
		}
	}

	app.Run(os.Args)
}

type ServerRedisCfg struct {
	Gid  string
	Sid  string
	Url  string
	Db   int
	Auth string
}

const (
	CSV_GID = iota
	CSV_SID
	CSV_URL
	CSV_DB
	CSV_AUTH
)

func loadServerCfg() []ServerRedisCfg {
	serverFile, err := os.Open("server.csv")
	if err != nil {
		panic(err)
	}
	csvReader := csv.NewReader(serverFile)
	csvAll, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}
	//fmt.Println(csvAll)
	cfgs := make([]ServerRedisCfg, 0)
	for _, str := range csvAll {
		db, err := strconv.Atoi(str[CSV_DB])
		if err != nil {
			panic(err)
		}
		cfg := ServerRedisCfg{
			Gid:  str[CSV_GID],
			Sid:  str[CSV_SID],
			Url:  str[CSV_URL],
			Db:   db,
			Auth: str[CSV_AUTH],
		}
		cfgs = append(cfgs, cfg)
	}
	return cfgs
}

type GuildLostInventory struct {
	GuildUid      string
	LostInventory string
}

func loadOldInventory(fileName string) map[string]guild_info.GuildInventory {
	lostFile, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	lostReader := bufio.NewReader(lostFile)
	inventories := make(map[string]guild_info.GuildInventory)
	for {
		str, err := lostReader.ReadString('\n')
		if err != nil && err != io.EOF {
			panic(err)
		} else if err == io.EOF {
			break
		} else {
			inventory := &GuildLostInventory{}
			err = json.Unmarshal([]byte(str), inventory)
			if err != nil {
				panic(err.Error() + str)
			}
			ivy := &guild_info.GuildInventory{}
			err = json.Unmarshal([]byte(inventory.LostInventory), ivy)
			if err != nil {
				panic(err)
			}
			inventories[inventory.GuildUid] = *ivy
		}
	}
	return inventories
}

func saveIvy(redisCfg ServerRedisCfg) {
	conn, err := redis.Dial("tcp", redisCfg.Url,
		redis.DialConnectTimeout(1*time.Second),
		redis.DialReadTimeout(5*time.Second),
		redis.DialWriteTimeout(5*time.Second),
		redis.DialPassword(redisCfg.Auth),
		redis.DialDatabase(redisCfg.Db))
	if err != nil {
		fmt.Println("connect redis err", err)
		return
	}
	guildKeys, err := scanGuildKey(conn)
	if err != nil {
		fmt.Println("scan guild err", err)
		return
	}
	//fmt.Println(guildKeys)
	fmt.Println("got guild key, ", len(guildKeys))
	if guildKeys == nil || len(guildKeys) == 0 {
		return
	}
	for _, guildKey := range guildKeys {
		if ivy, ok := ivyMap[guildKey]; !ok {
			continue
		} else {
			newIvy := modifyInventory(ivy)
			if newIvy == nil {
				//fmt.Println("modify inventory nil ")
				continue
			}
			newIvyJson, err := json.Marshal(newIvy)
			if err != nil {
				fmt.Println("json marshal guild err ", err)
				continue
			}
			_, err = conn.Do("HSET", guildKey, "LostInventory", newIvyJson)
			if err != nil {
				fmt.Println("save guild err ", err)
				continue
			}
			fmt.Println("save guild ok", guildKey)
		}
	}
}

func modifyInventory(inventory guild_info.GuildInventory) *guild_info.GuildInventory {
	if len(inventory.Loots) == 0 && len(inventory.PrepareLoots) == 0 {
		return nil
	}
	inventory.NextRefTime = 0
	inventory.NextResetTime = 0
	inventory.LostActive = true
	changeInventory(&inventory)
	return &inventory
}

func changeInventory(ivy *guild_info.GuildInventory) {
	allIvyMap := make(map[string]*guild_info.GuildInventoryLoot)
	for i, loot := range ivy.PrepareLoots {
		if lt, ok := allIvyMap[loot.LootId]; ok {
			lt.Count += loot.Count
			if lt.Count > 999 {
				lt.Count = 999
			}
		} else {
			allIvyMap[loot.LootId] = &ivy.PrepareLoots[i]
		}
	}

	for i, loot := range ivy.Loots {
		if lt, ok := allIvyMap[loot.Loot.LootId]; ok {
			lt.Count += loot.Loot.Count
			if lt.Count > 999 {
				lt.Count = 999
			}
		} else {
			allIvyMap[loot.Loot.LootId] = &ivy.Loots[i].Loot
		}
	}

	lootList := make([]guild_info.GuildInventoryLootAndMem, 0)
	for _, value := range allIvyMap {
		lootList = append(lootList, guild_info.GuildInventoryLootAndMem{
			Loot: *value,
		})
	}
	ivy.PrepareLoots = nil
	ivy.Loots = lootList
	for i, loot := range ivy.Loots {
		if newId, ok := lootIdMap[loot.Loot.LootId]; ok {
			ivy.Loots[i].Loot.LootId = newId
		}
	}
}

func scanGuildKey(conn redis.Conn) ([]string, error) {
	index := 0
	guildKeys := make([]string, 0)
	for {
		reply, err := conn.Do("SCAN", index, "MATCH", "guild:Info:*")
		if err != nil {
			fmt.Println(err)
			break
		}
		reply1, err := redis.Values(reply, nil)
		if len(reply1) != 2 {
			fmt.Println("bad result ", reply1, err)
			break
		}
		index, err = redis.Int(reply1[0], nil)
		if err != nil {
			fmt.Println(err)
			break
		}
		//fmt.Println("next index ", index)
		tempKeys, err := redis.Strings(reply1[1], nil)
		if err != nil {
			fmt.Println(err)
			break
		}
		guildKeys = append(guildKeys, tempKeys...)
		if index == 0 {
			break
		}
	}

	return guildKeys, nil
}
