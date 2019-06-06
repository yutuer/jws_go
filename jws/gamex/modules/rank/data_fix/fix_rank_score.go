package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

const (
	PowerBase        = 100000
	SectionNum       = 5
	Redis_ZADD_Banch = 100
)

func main() {
	if 3 > len(os.Args) {
		fmt.Println("too less param")
		return
	}

	filename := os.Args[1]
	list := parseFile(filename)
	//fmt.Println("---", list)

	cmd := os.Args[2]

	switch cmd {
	case "rename":
		for _, t := range list {
			fmt.Printf("back %s:%s\n", t.gameid, t.shardid)
			backOneServer(t)
		}
	case "do":
		for _, t := range list {
			fmt.Printf("do %s:%s\n", t.gameid, t.shardid)
			doOneServer(t)
		}
	default:
		fmt.Println("unkown cmd")
	}
}

func backOneServer(t target) {
	conn := getConn(t)
	if nil == conn {
		fmt.Println("  no connection")
		return
	}
	defer conn.Close()

	renameRank(conn, t, "RankCorpHeroStar")
	renameRank(conn, t, "RankCorpTrial")

	renameRank(conn, t, "RankCorpHeroDiff:TU")
	renameRank(conn, t, "RankCorpHeroDiff:ZHAN")
	renameRank(conn, t, "RankCorpHeroDiff:HU")
	renameRank(conn, t, "RankCorpHeroDiff:SHI")
}

func renameRank(conn redis.Conn, t target, rn string) {
	//合并排行数据 redis key
	table_name := fmt.Sprintf("%s:%s:%s", t.gameid, t.shardid, rn)
	back_name := "Back:" + table_name

	//检查key是否存在
	//rets, err := redis.Strings(conn.Do("KEYS", table_name))
	ret, err := redis.String(conn.Do("type", table_name))
	if nil != err {
		fmt.Printf("    !! do redis KEYS [%s] failed,\n", table_name, err)
		return
	}
	if "zset" != ret {
		fmt.Printf("    renameRank [%s] not exist\n", table_name)
		return
	}

	//删除key的内容:用改名的方法
	_, err = conn.Do("RENAME", table_name, back_name)
	if nil != err {
		fmt.Println("    !! do redis RENAME failed,", err)
	}
	fmt.Printf("    renameRank [%s] over\n", table_name)
}

func doOneServer(t target) {
	conn := getConn(t)
	if nil == conn {
		fmt.Println("  no connection")
		return
	}
	defer conn.Close()
	//fmt.Printf("OK for %s:%s\n", t.gameid, t.shardid)

	fixRank(conn, t, "RankCorpHeroStar")
	fixRank(conn, t, "RankCorpTrial")
}

func fixRank(conn redis.Conn, t target, rn string) {
	//合并排行数据 redis key
	table_name := fmt.Sprintf("%s:%s:%s", t.gameid, t.shardid, rn)
	back_name := "Back:" + table_name

	//检查key是否存在
	//rets, err := redis.Strings(conn.Do("KEYS", back_name))
	ret, err := redis.String(conn.Do("type", back_name))
	if nil != err {
		fmt.Println("    !! do redis KEYS failed,", err)
		return
	}
	if "zset" != ret {
		fmt.Printf("    fixRank [%s] not exist\n", back_name)
		return
	}

	//取出key的内容
	list, err := redis.Strings(conn.Do("ZRANGE", back_name, 0, -1, "WITHSCORES"))
	if nil != err {
		fmt.Printf("    !! fixRank [%s] ZRANGE failed, %v\n", back_name, err)
		return
	}

	//转换分值并组织数据
	packets, count := buildPackets(list)
	fmt.Printf("    fixRank [%s] fix num [%d]\n", table_name, count)

	//数据写会redis
	for _, pack := range packets {
		_, err := conn.Do("ZADD", append([]interface{}{table_name}, pack...)...)
		if nil != err {
			fmt.Printf("    !! fixRank [%s] ZADD failed, list {%v} %v\n", table_name, pack, err)
		}
	}
	fmt.Printf("    fixRank [%s] over\n", table_name)
}

func buildPackets(list []string) ([][]interface{}, int) {
	//组织写入数据
	packet_index := 0
	packet_len := Redis_ZADD_Banch * 2
	packet_inner := 0
	packets := [][]interface{}{}
	for i := 0; i < len(list); i += 2 {
		mem := list[i]
		//计算正确分值
		score, err := strconv.ParseFloat(list[i+1], 64)
		if nil != err {
			fmt.Printf("fixHeroStar ParseFloat [%s:%s] failed, %v\n", list[i], list[i+1], err)
			return [][]interface{}{}, 0
		}
		//如果本身是正确的,跳过这个数值对
		if score < PowerBase {
			score = score * PowerBase
		}

		if 0 == packet_inner%packet_len {
			pack := []interface{}{}
			packets = append(packets, pack)
			packet_index += 1
		}

		packets[packet_index-1] = append(packets[packet_index-1], fmt.Sprintf("%f", score), mem)

		packet_inner++
	}

	return packets, packet_inner
}

func getConn(t target) redis.Conn {
	conn, err := redis.Dial(
		"tcp",
		fmt.Sprintf("%s:%s", t.ip, "6379"),
		redis.DialPassword(t.pwd),
		redis.DialDatabase(t.db),
	)
	if nil != err {
		fmt.Println("  connect redis failed,", err)
		return nil
	}

	return conn
}

type target struct {
	gameid  string
	shardid string
	ip      string
	db      int
	pwd     string
}

func parseFile(filename string) []target {
	file, err := os.Open(filename)
	if nil != err {
		fmt.Println("open file failed,", err)
		os.Exit(1)
	}

	list := []target{}

	reader := bufio.NewReader(file)
	co := true
	for co {
		line, err := reader.ReadString('\n')
		if nil != err && err != io.EOF {
			fmt.Println("read file error,", err)
			os.Exit(1)
		}
		if io.EOF == err {
			//fmt.Printf("read file over, least :[%s]\n", line)
			co = false
		}

		line = strings.TrimSuffix(line, "\n")

		secs := strings.Split(line, ",")
		if SectionNum != len(secs) {
			fmt.Println("uncorrect sections:", line)
			continue
		}

		db, err := strconv.Atoi(secs[3])
		if nil != err {
			fmt.Println("read file parse db id failed,", err)
			os.Exit(1)
		}

		t := target{
			gameid:  secs[0],
			shardid: secs[1],
			ip:      secs[2],
			db:      db,
			pwd:     secs[4],
		}
		list = append(list, t)
	}

	return list
}
