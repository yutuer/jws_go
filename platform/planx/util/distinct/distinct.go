package distinct

import (
	"vcs.taiyouxi.net/platform/planx/util/distinct/set"
	"fmt"
	"errors"
)

func disinct(info []interface{}) ([]interface{}, error) {
	set := set.New()
	for _, v := range info {
		set.Add(v)
	}
	return set.List(), nil
}
//将任意类型转换为[]interface{},并调用Disinct
func ValuesAndDisinct(reply interface{}) ([]interface{}, error) {
	switch reply := reply.(type) {
	case []uint:
		info := []interface{}{}
		for _,v:=range reply{
			info = append(info,v)
		}
		return disinct(info)
	case []uint8:
		info := []interface{}{}
		for _,v:=range reply{
			info = append(info,v)
		}
		return disinct(info)
	case []uint16:
		info := []interface{}{}
		for _,v:=range reply{
			info = append(info,v)
		}
		return disinct(info)
	case []uint32:
		info := []interface{}{}
		for _,v:=range reply{
			info = append(info,v)
		}
		return disinct(info)
	case []uint64:
		info := []interface{}{}
		for _,v:=range reply{
			info = append(info,v)
		}
		return disinct(info)
	case []int:
		info := []interface{}{}
		for _,v:=range reply{
			info = append(info,v)
		}
		return disinct(info)
	case []int8:
		info := []interface{}{}
		for _,v:=range reply{
			info = append(info,v)
		}
		return disinct(info)
	case []int16:
		info := []interface{}{}
		for _,v:=range reply{
			info = append(info,v)
		}
		return disinct(info)
	case []int32:
		info := []interface{}{}
		for _,v:=range reply{
			info = append(info,v)
		}
		return disinct(info)
	case []int64:
		info := []interface{}{}
		for _,v:=range reply{
			info = append(info,v)
		}
		return disinct(info)
	case []string:
		info := []interface{}{}
		for _,v:=range reply{
			info = append(info,v)
		}
		return disinct(info)
	case []interface{}:
		return disinct(reply)
	default:
		return nil, errors.New("unexpected input,need slice")
	}
	return nil, fmt.Errorf("redigo: unexpected type for Values, got type %T", reply)
}