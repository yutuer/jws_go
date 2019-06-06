package util

import "strings"

func IsNeedOnLand(need_store_heads map[string]bool, key string) bool {
	s := strings.Split(key, ":")
	if len(s) <= 0 {
		return false
	}

	_, ok := need_store_heads[s[0]]
	if !ok {
		if len(s) >= 2 {
			s2 := strings.Join(s[:2], ":")
			_, ok := need_store_heads[s2]
			return ok
		}
	}
	return ok
}
