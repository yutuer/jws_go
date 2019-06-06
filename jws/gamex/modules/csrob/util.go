package csrob

import "fmt"

func makeError(format string, v ...interface{}) error {
	return fmt.Errorf("[CSRob] "+format, v...)
}

type sortEnemyList []GuildEnemy

func (s sortEnemyList) Len() int {
	return len(s)
}
func (s sortEnemyList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s sortEnemyList) Less(i, j int) bool {
	return s[i].Count > s[j].Count
}
