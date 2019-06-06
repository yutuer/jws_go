package redis_helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

type ManualDBDump interface {
	// val是自身，返回序列化后的基础类型
	ToDB(val interface{}) interface{}
	// data是序列化后的基础类型，obj是自身的value
	FromDB(data interface{}, obj reflect.Value) error
}

/////////////////
//Helper function for DumpToHashDB of Redis
func init() {
}

type XArgs struct {
	redis.Args
}

func (args XArgs) Add(value ...interface{}) XArgs {
	args.Args = args.Args.Add(value...)
	return args
}

func (args XArgs) AddFlat(v interface{}) XArgs {
	args.Args = args.Args.AddFlat(v)
	return args
}

//语意上这个函数是只为了HMSET而存在的
func (args XArgs) AddHash(v interface{}, dirtyHash map[string]interface{}) (
	res XArgs, newDirtyHash map[string]interface{}, chged []string) {

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			logs.Warn("AddHash Ptr rv %v is Nil!", rv)
			return args, nil, nil
		} else if !rv.IsValid() {
			logs.Warn("AddHash Ptr rv %v is not Valid!", rv)
			return args, nil, nil
		} else {
			rv = rv.Elem()
		}
	}

	switch rv.Kind() {
	case reflect.Struct:
		//logs.Trace("AddHash Struct!")
		args.Args, newDirtyHash, chged = flattenStruct(args.Args, rv, dirtyHash)
	case reflect.Map:
		//logs.Trace("AddHash Map!")
		for _, k := range rv.MapKeys() {
			lv := rv.MapIndex(k)
			lvi := lv.Interface()
			if needJson(lv) {
				jval, err := json.Marshal(lvi)
				if err != nil {
					panic(err)
				}
				args.Args = append(args.Args, k.Interface(), jval)
			} else {
				args.Args = append(args.Args, k.Interface(), lvi)
			}
		}
	default:
		args.Args = append(args.Args, v)
	}
	return args, newDirtyHash, chged
}

func needJson(v reflect.Value) bool {
	switch v.Interface().(type) {
	case string, int, int64, float64, bool, nil, []byte:
		return false
	default:
		return true
	}
	return false
}

func ToDB(val interface{}) interface{} {
	rval := reflect.ValueOf(val)
	if rval.CanAddr() {
		if structConvert, ok := rval.Addr().Interface().(ManualDBDump); ok {
			return structConvert.ToDB(val)
		}
	}

	if structConvert, ok := rval.Interface().(ManualDBDump); ok {
		return structConvert.ToDB(val)
	} else {
		return ToDB_default(val)
	}
}

func ToDB_default(val interface{}) interface{} {
	rval := reflect.ValueOf(val)
	is_need_json := needJson(rval)
	if is_need_json {
		jval, err := json.Marshal(val)
		if err != nil {
			panic(err)
		}
		return jval
	} else {
		return val
	}
}

func FromDB(src_data interface{}, dest reflect.Value) error {
	if dest.CanAddr() {
		if structConvert, ok := dest.Addr().Interface().(ManualDBDump); ok {
			return structConvert.FromDB(src_data, dest)
		}
	}

	if structConvert, ok := dest.Interface().(ManualDBDump); ok {
		return structConvert.FromDB(src_data, dest)
	} else {
		return FromDB_default(src_data, dest)
	}
}

func FromDB_default(src_data interface{}, dest reflect.Value) error {
	is_need_json := needJson(dest)
	if is_need_json {
		data, ok := src_data.([]byte)
		if !ok {
			return errors.New("fromDB data is not []byte!")
		}

		res := reflect.New(dest.Type()).Interface()

		err := json.Unmarshal(data, &res)
		if err != nil {
			return err
		}
		if res == nil {
			return fmt.Errorf("fromDB failed unMarshal json from %s", data)
		}
		dest.Set(reflect.ValueOf(res).Elem())
		return nil
	} else {
		return errors.New("Can not json!")
	}
}

func GetInItProfileDBKey(key string) string {
	dbkey, err := db.ParseProfileDbKey(key)
	if err != nil {
		logs.Error("key %s format error", key)
		return ""
	}
	dbkey.Account.UserId = db.NewUserIDWithName("init")

	return fmt.Sprintf("init:%s", dbkey)
}

func GenDirtyHash(v interface{}) map[string]interface{} {
	_, ret, _ := XArgs{}.AddHash(v, nil)
	return ret
}

func DumpToHashDBCmcBufferCheckDirty(cb redis.CmdBuffer, key string,
	v interface{}, dirtyHash map[string]interface{}) (error, map[string]interface{}, []string) {
	args, newDirtyHash, chged := XArgs{}.Add(key).AddHash(v, dirtyHash)
	cmds := args.Args
	if len(cmds) <= 1 {
		return nil, newDirtyHash, chged
	}
	err := cb.Send("HMSET", cmds...)
	if err != nil {
		logs.Error("Dump To CmdBuffer failed, key:%s. err:%s, cmds: %v", key, err.Error(), cmds)
		return err, nil, nil
	}
	logs.Debug("Dump To %s CmdBuffer success.", key)
	return nil, newDirtyHash, chged
}

func DumpToHashDBCmcBuffer(cb redis.CmdBuffer, key string, v interface{}) error {
	args, _, _ := XArgs{}.Add(key).AddHash(v, nil)
	cmds := args.Args
	err := cb.Send("HMSET", cmds...)
	if err != nil {
		logs.Error("Dump To CmdBuffer failed, key:%s. err:%s, cmds: %v", key, err.Error(), cmds)
		return err
	}
	//logs.Debug("Dump To %s CmdBuffer success.", key)
	return nil
}

//func DumpToHashDB(db redis.Conn, key string, v interface{}) error {
//	if db == nil {
//		return fmt.Errorf("DumpToHashDB db nil")
//	}
//	cmds := XArgs{}.Add(key).AddHash(v).Args
//	_, err := redis.String(db.Do("HMSET", cmds...))
//	if err != nil {
//		logs.Error("Dump To %s failed. err:%s, cmds: %v", key, err.Error(), cmds)
//		return err
//	}
//	logs.Debug("Dump To %s success.", key)
//	return nil
//}

// RESTORE_ERR_Profile_No_Data 表示玩家第一次登陆游戏，没有存档，这不视为Bug
// 外面的逻辑需要根据此判断是否是第一次登陆游戏
var RESTORE_ERR_Profile_No_Data = errors.New("Profile_No_Data")

func RestoreFromHashDB(db redis.Conn, key string, v interface{}, isNeedInit, logInfo bool) error {
	if db == nil {
		return fmt.Errorf("RestoreFromHashDB db nil")
	}
	// T990 TASK 增加服务器拷贝存档功能
	// 初始化存档时如果是开发模式的话，会以“init:XXXX:0:0:1”中的数据作为初始数据
	// 如果“init:XXXX:0:0:1”没有的话就还是用默认空存档
	if isNeedInit {
		key = GetInItProfileDBKey(key)
	}

	values, err := redis.Values(db.Do("HGETALL", key))
	if err != nil {
		logs.Warn("Restore To %s failed. err:%s", key, err.Error())
		return err
	}

	if len(values) == 0 {
		// 初始化存档不在这里实现了
		//if (!is_need_init) && devMode {
		// 初始化存档时如果是开发模式的话，会以“init:XXXX:0:0:1”中的数据作为初始数据
		//	RestoreFromHashDB(db, key, v, true)
		//	return RESTORE_ERR_Profile_No_Data
		//} else {
		//如果“init:XXXX:0:0:1”没有的话就还是用默认空存档
		return RESTORE_ERR_Profile_No_Data
		//}
	}

	//logs.Trace("Restore To %s, with value: %v", key, values)
	logs.Trace("Restore To %s", key)
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	} else {
		logs.Error("RestoreFromHashDB need a Ptr %s", key)
		return errors.New("RestoreFromHashDB need a Ptr")
	}

	// 结构升级(小)
	// 这里要现根据需要升级存档字符串
	//

	switch rv.Kind() {
	case reflect.Map:
		err = scanMap(values, rv)
		if err != nil {
			logs.Warn("Restore To %s failed in ScanMap. err:%s", key, err.Error())
			return err
		}
	default:
		err = scanStruct(values, rv)
		if err != nil {
			logs.Warn("Restore To %s failed in ScanStruct. err:%s", key, err.Error())
			return err
		}
	}
	if logInfo {
		logs.Info("Restore To %s success.", key)
	}
	return nil
}

func RedisSaveDataToOther(db redispool.RedisPoolConn, from, to string) bool {
	logs.Info("RedisSaveDataToOtherKey : %s -> %s", from, to)

	lua_src := `
		local from = KEYS[1]
		local to = KEYS[2]

		if redis.call("EXISTS", to) == 1 then
		  redis.call("DEL", to)
		end

		redis.call("RESTORE", to, 0, redis.call("DUMP", from))

		return "OK"
	`

	rc := db.RawConn()
	save_data_to_other_src := redis.NewScript(2, lua_src)
	save_data_to_other_src.Load(rc)

	res, err := redis.String(save_data_to_other_src.Do(rc, from, to))
	logs.Warn("RedisSaveDataToOtherKey Re : %v  %v", err, res)

	return res == "OK"
}

func PanicIfErr(err error) bool {
	if err == nil {
		return false
	}

	if err == RESTORE_ERR_Profile_No_Data {
		return true
	}

	//NewAccount数据加载后如何处理玩家数据加载错误,
	//如果玩家数据不存在是不会引发错误的。这里的错误应该是数据库自身的错误。
	//此外此函数是通过GetMux|Player playerProcessor调用，是用户goroutine级别的错误，不会影响其他玩家
	panic(err)
	return false
}
