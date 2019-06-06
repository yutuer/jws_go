package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"path/filepath"

	"io"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/tools/gen_shard_name/shard_init"
)

type gameID struct {
	Name string
	Id   int
}

type shardID struct {
	name        string
	id          int
	gid         int
	displayName string
	showState   string
}

// TODO 这个数据结构在Auth rest api部分也存在，考虑整合！
type ShardInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"dn"`
	ShowState   string `json:"ss"`
}

func dangerInit(c *cli.Context) {
	filePath := c.String("output")
	writer, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}
	protoWriter := NewRESPWriter(writer)

	gameInit(protoWriter)
	shardInit(protoWriter)
}

// 客户端应该拿到的是shardname:displayname的关系
func gameInit(protoWriter *RESPWriter) {
	file, err := os.Open("conf/game_id.csv")
	if err != nil {
		// err is printable
		// elements passed are separated by space automatically
		fmt.Println("Error:", err.Error())
		return
	}
	defer file.Close()

	fmt.Println("读取game csv文件，初始化...")
	r2 := csv.NewReader(file)
	ss, err := r2.ReadAll()

	if err != nil {
		// err is printable
		// elements passed are separated by space automatically
		log.Fatalln("Error csv:", err.Error())
		return
	}
	for _, r := range ss {
		fmt.Println(r)
	}

	var cmds2 Cmds
	cmds2 = cmds2.Add("HMSET", "cfg:gamenames")
	// fmt.Printf("HMSET cfg:gamenames ")
	for i, r := range ss {
		id, err := strconv.Atoi(r[1])
		if err != nil {
			log.Fatalln("danger-init: found game id is not number", i, r)
		}
		cmds2 = cmds2.Add(
			fmt.Sprintf("%d", id),
			fmt.Sprintf("%s", r[0]))
		// fmt.Printf(`%d %q `, id, r[0])
	}
	// fmt.Println("\n")
	cmds2.Println()
	protoWriter.WriteCommand(cmds2...)
}

const (
	etcd_output_dir = "output"
)

func shardInit(protoWriter *RESPWriter) {

	// prepare for etcd
	if err := prepareEtcdOutput(); err != nil {
		log.Fatalln("danger-init: prepareEtcdOutput err", err)
		return
	}

	file, err := os.Open("conf/shard_id.csv")
	if err != nil {
		// err is printable
		// elements passed are separated by space automatically
		fmt.Println("Error:", err.Error())
		return
	}
	defer file.Close()

	fmt.Println("读取shard csv文件，初始化...")
	r2 := csv.NewReader(file)
	ss, err := r2.ReadAll()

	if err != nil {
		// err is printable
		// elements passed are separated by space automatically
		log.Fatalln("Error csv:", err.Error())
		return
	}

	shards := make(map[string]string)
	shardIDs := make(map[int]shardID)
	shardsByGID := make(map[int][]ShardInfo)
	idx := 0
	for i, r := range ss {
		//fmt.Println(r, len(r), r[1])
		if idx != 0 {
			id, err := strconv.Atoi(r[0])
			if err != nil {
				log.Fatalln("danger-init: found  shard id is not number", i, r)
			}
			gid, err := strconv.Atoi(r[1])
			if err != nil {
				log.Fatalln("danger-init: found  game id is not number", i, r)
			}
			ss := ""
			if len(r) >= 4 {
				ss = r[3]
			}

			v := shardID{
				name:        "",
				id:          id,
				gid:         gid,
				displayName: r[2],
				showState:   ss,
			}

			//生成唯一Shard name，稳定算法不需要记录
			if v.name == "" {
				v.name = shard_init.MkShardName(id, gid)
			}

			//检查shard name是否重复
			if _, ok := shards[v.name]; ok {
				log.Fatalln("danger-init: load duplicated shard name. line:", i, r)
			}
			shardinfo, err := json.Marshal(struct {
				ID  int
				GID int
			}{
				ID:  v.id,
				GID: v.gid,
			})
			shards[v.name] = string(shardinfo)

			//检查shard id (sid)是否重复
			if _, ok := shardIDs[v.id]; ok {
				log.Fatalln("danger-init: load duplicated shard id. line:", i, r)
			}
			shardIDs[v.id] = v

			if _, ok := shardsByGID[gid]; !ok {
				shardsByGID[gid] = make([]ShardInfo, 0, 10)
			}
			shardsByGID[gid] = append(shardsByGID[gid], ShardInfo{v.name, v.displayName, v.showState})

			// for etcd
			if err := os.MkdirAll(fmt.Sprintf("%s/%d/%d", etcd_output_dir, gid, id), 0777); err != nil {
				log.Fatalln("danger-init: MkdirAll for etcd err. line:", i, r, err)
				return
			}
			f, err := os.Create(fmt.Sprintf("%s/%d/%d/sn", etcd_output_dir, gid, id))
			if err != nil {
				log.Fatalln("danger-init: create file for etcd err. line:", i, r, err)
				return
			}
			defer f.Close()
			_, err = io.WriteString(f, v.name)
			if err != nil {
				log.Fatalln("danger-init: write file for etcd err. line:", i, r, err)
				return
			}
		}
		//fmt.Println(r, len(r), r[1])
		idx++
	}

	//初始化数据库
	//直接写数据库这个事情，太危险啦！还是生成指令，手动来做吧
	fmt.Println("推送到数据库指令集生成...\n")
	//login 服务器使用数据

	//HSET shards shardname sid
	//for k, v := range shards {
	//fmt.Printf("HSET cfg:shards %s %d\n", k, v)
	//}

	var cmds Cmds
	cmds = cmds.Add("HMSET", "cfg:shards")
	// fmt.Printf("HMSET cfg:shards ")
	for k, v := range shards {
		cmds = cmds.Add(k, v)
		// fmt.Printf("%s %q ", k, v)
	}
	// fmt.Println("\n")
	protoWriter.WriteCommand(cmds...)
	cmds.Println()

	//for k, v := range shardsByGID {
	//b, err := json.Marshal(v)
	//if err != nil {
	//log.Fatalln("clientCfgShards 生成失败：", err.Error())
	//}
	//fmt.Printf("HSET cfg:client_shards %d %s\n", k, string(b))
	//}

	var cmds2 Cmds
	cmds2 = cmds2.Add("HMSET", "cfg:client_shards")
	// fmt.Printf("HMSET cfg:client_shards ")
	for k, v := range shardsByGID {
		b, err := json.Marshal(v)
		if err != nil {
			log.Fatalln("clientCfgShards 生成失败：", err.Error())
		}
		cmds2 = cmds2.Add(
			fmt.Sprintf("%d", k), string(b))

		// fmt.Printf(`%d %q `, k, string(b))
	}
	// fmt.Println("\n")
	protoWriter.WriteCommand(cmds2...)
	cmds2.Println()

	fmt.Println("推送到数据库指令集生成完毕")

}

// 删除output目录
func prepareEtcdOutput() error {
	if _, err := os.Stat(etcd_output_dir); err != nil {
		return nil
	}

	fileNames := make([]string, 0)
	dirNames := make([]string, 0)
	err := filepath.Walk(etcd_output_dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			dirNames = append(dirNames, path)
		} else {
			fileNames = append(fileNames, path)
		}
		return err
	})
	if err != nil {
		return err
	}

	for _, fn := range fileNames {
		if err := os.Remove(fn); err != nil {
			return err
		}
	}
	if err := os.RemoveAll(etcd_output_dir); err != nil {
		return err
	}
	return nil
}

func init() {
	register(&cli.Command{
		Name: "danger-init",
		//ShortName: "",
		Usage:  "根据game_id.csv, shard_id.csv初始化数据库",
		Action: dangerInit,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "output, o",
				Value: "danger-init-redis.bin",
				Usage: "The redis protocol for `redis-cli --pipe` usage",
			},
		},
	})
}
