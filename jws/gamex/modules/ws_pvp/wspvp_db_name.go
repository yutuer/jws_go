package ws_pvp

import "fmt"

func getPersonalTableName(groupId int, acid string) string {
	return fmt.Sprintf("info:%d:%s", groupId, acid)
}

func getRankTableName(groupId int) string {
	return fmt.Sprintf("rank:%d", groupId)
}

func getLockTableName(groupId int) string {
	return fmt.Sprintf("lock:%d", groupId)
}

func getBattleLogTableName(groupId int, acid string) string {
	return fmt.Sprintf("log:%d:%s", groupId, acid)
}

func getRobotTableName(groupId int) string {
	return fmt.Sprintf("robot:%d", groupId)
}

func getBest9RankTableName(groupId int) string {
	return fmt.Sprintf("best9rank:%d", groupId)
}

func getInitKeyTableName(groupId int) string {
	return fmt.Sprintf("wspvp:initkey:%d", groupId)
}
