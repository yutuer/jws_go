package util

import "sort"

//SortPair 键值对元素
type SortPair struct {
	Key int
	Val int
}

//SortPairList 键值对list
type SortPairList []SortPair

func (s SortPairList) Len() int {
	return len(s)
}

func (s SortPairList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortPairList) Less(i, j int) bool {
	return s[i].Val > s[j].Val
}

//SortIntMapKeyByValue 从map中按照值排序,返回排序后的key值
func SortIntMapKeyByValue(m map[int]int) []int {
	s := []SortPair{}

	for k, v := range m {
		s = append(s, SortPair{Key: k, Val: v})
	}

	sort.Sort(SortPairList(s))

	ret := []int{}
	for _, elem := range s {
		ret = append(ret, elem.Key)
	}

	return ret
}
