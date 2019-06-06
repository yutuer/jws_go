package redis_helper

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"crypto/md5"

	"fmt"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

type fieldSpec struct {
	name  string
	index []int
	//omitEmpty bool
}

type structSpec struct {
	m map[string]*fieldSpec
	l []*fieldSpec
}

var (
	structSpecMutex  sync.RWMutex
	structSpecCache  = make(map[reflect.Type]*structSpec)
	defaultFieldSpec = &fieldSpec{}
)

func structSpecForType(t reflect.Type) *structSpec {

	structSpecMutex.RLock()
	ss, found := structSpecCache[t]
	structSpecMutex.RUnlock()
	if found {
		return ss
	}

	structSpecMutex.Lock()
	defer structSpecMutex.Unlock()
	ss, found = structSpecCache[t]
	if found {
		return ss
	}

	ss = &structSpec{m: make(map[string]*fieldSpec)}
	compileStructSpec(t, make(map[string]int), nil, ss)
	structSpecCache[t] = ss
	return ss
}

func compileStructSpec(t reflect.Type, depth map[string]int, index []int, ss *structSpec) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		switch {
		case f.PkgPath != "":
			// Ignore unexported fields.
		case f.Anonymous:
			// TODO: Handle pointers. Requires change to decoder and
			// protection against infinite recursion.
			if f.Type.Kind() == reflect.Struct {
				compileStructSpec(f.Type, depth, append(index, i), ss)
			}
		default:
			fs := &fieldSpec{name: f.Name}
			tag := f.Tag.Get("redis")
			p := strings.Split(tag, ",")
			if len(p) > 0 {
				if p[0] == "-" {
					continue
				}
				if len(p[0]) > 0 {
					fs.name = p[0]
				}
				for _, s := range p[1:] {
					switch s {
					//case "omitempty":
					//  fs.omitempty = true
					default:
						panic(errors.New("redigo: unknown field flag " + s + " for type " + t.Name()))
					}
				}
			}
			d, found := depth[fs.name]
			if !found {
				d = 1 << 30
			}
			switch {
			case len(index) == d:
				// At same depth, remove from result.
				delete(ss.m, fs.name)
				j := 0
				for i := 0; i < len(ss.l); i++ {
					if fs.name != ss.l[i].name {
						ss.l[j] = ss.l[i]
						j += 1
					}
				}
				ss.l = ss.l[:j]
			case len(index) < d:
				fs.index = make([]int, len(index)+1)
				copy(fs.index, index)
				fs.index[len(index)] = i
				depth[fs.name] = len(index)
				ss.m[fs.name] = fs
				ss.l = append(ss.l, fs)
			}
		}
	}
}

func (ss *structSpec) fieldSpec(name []byte) *fieldSpec {
	return ss.m[string(name)]
}

func ensureLen(d reflect.Value, n int) {
	if n > d.Cap() {
		d.Set(reflect.MakeSlice(d.Type(), n, n))
	} else {
		d.SetLen(n)
	}
}

func cannotConvert(d reflect.Value, s interface{}) error {
	return errors.New("Scan cannot convert")
}

func convertAssignBytes(d reflect.Value, s []byte) (err error) {
	switch d.Type().Kind() {
	case reflect.Float32, reflect.Float64:
		var x float64
		x, err = strconv.ParseFloat(string(s), d.Type().Bits())
		d.SetFloat(x)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var x int64
		x, err = strconv.ParseInt(string(s), 10, d.Type().Bits())
		d.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var x uint64
		x, err = strconv.ParseUint(string(s), 10, d.Type().Bits())
		d.SetUint(x)
	case reflect.Bool:
		var x bool
		x, err = strconv.ParseBool(string(s))
		d.SetBool(x)
	case reflect.String:
		d.SetString(string(s))
	case reflect.Slice:
		if d.Type().Elem().Kind() != reflect.Uint8 {
			err = cannotConvert(d, s)
		} else {
			d.SetBytes(s)
		}
	default:
		err = cannotConvert(d, s)
	}
	return
}

func convertAssignInt(d reflect.Value, s int64) (err error) {
	switch d.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		d.SetInt(s)
		if d.Int() != s {
			err = strconv.ErrRange
			d.SetInt(0)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if s < 0 {
			err = strconv.ErrRange
		} else {
			x := uint64(s)
			d.SetUint(x)
			if d.Uint() != x {
				err = strconv.ErrRange
				d.SetUint(0)
			}
		}
	case reflect.Bool:
		d.SetBool(s != 0)
	default:
		err = cannotConvert(d, s)
	}
	return
}

func convertAssignValue(d reflect.Value, s interface{}) (err error) {
	switch s := s.(type) {
	case []byte:
		err = convertAssignBytes(d, s)
	case int64:
		err = convertAssignInt(d, s)
	default:
		err = cannotConvert(d, s)
	}
	return err
}

func convertAssignValues(d reflect.Value, s []interface{}) error {
	if d.Type().Kind() != reflect.Slice {
		return cannotConvert(d, s)
	}
	ensureLen(d, len(s))
	for i := 0; i < len(s); i++ {
		if err := convertAssignValue(d.Index(i), s[i]); err != nil {
			return err
		}
	}
	return nil
}

func convertAssign(d interface{}, s interface{}) (err error) {
	// Handle the most common destination types using type switches and
	// fall back to reflection for all other types.
	switch s := s.(type) {
	case nil:
		// ingore
	case []byte:
		switch d := d.(type) {
		case *string:
			*d = string(s)
		case *int:
			*d, err = strconv.Atoi(string(s))
		case *bool:
			*d, err = strconv.ParseBool(string(s))
		case *[]byte:
			*d = s
		case *interface{}:
			*d = s
		case nil:
			// skip value
		default:
			if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
				err = cannotConvert(d, s)
			} else {
				err = convertAssignBytes(d.Elem(), s)
			}
		}
	case int64:
		switch d := d.(type) {
		case *int:
			x := int(s)
			if int64(x) != s {
				err = strconv.ErrRange
				x = 0
			}
			*d = x
		case *bool:
			*d = s != 0
		case *interface{}:
			*d = s
		case nil:
			// skip value
		default:
			if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
				err = cannotConvert(d, s)
			} else {
				err = convertAssignInt(d.Elem(), s)
			}
		}
	case []interface{}:
		switch d := d.(type) {
		case *[]interface{}:
			*d = s
		case *interface{}:
			*d = s
		case nil:
			// skip value
		default:
			if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
				err = cannotConvert(d, s)
			} else {
				err = convertAssignValues(d.Elem(), s)
			}
		}
	default:
		err = cannotConvert(reflect.ValueOf(d), s)
	}
	return
}

func flattenStruct(args redis.Args, v reflect.Value, dirtyHash map[string]interface{}) (
	redis.Args, map[string]interface{}, []string) {

	ss := structSpecForType(v.Type())
	fields := ss.m //TODO by YZH XArgs需要类似redigo里面的缓存 T4229
	ret := make(map[string]interface{}, 16)
	chg := make([]string, 0, 16)
	for _, field := range fields {
		name := field.name
		//logs.Trace("[flattenStruct] %s, %s", field.Tag("redis"), name)
		val := v.FieldByIndex(field.index)
		if val.CanInterface() {
			jval := ToDB(val.Interface())

			hash := dirtyValue(jval)
			ret[name] = hash
			if dirtyHash != nil {
				if oldHash, ok := dirtyHash[name]; ok && oldHash == hash {
					continue
				}
			}
			chg = append(chg, name)
			args = append(args, name, jval)
		}
	}

	return args, ret, chg
}

func scanStruct(src []interface{}, rv reflect.Value) error {
	d := rv
	if d.Kind() != reflect.Struct {
		return errors.New("scanStruct dest no struct")
	}
	ss := structSpecForType(d.Type())

	if len(src)%2 != 0 {
		return errors.New("redigo: ScanStruct expects even number of values in values")
	}

	for i := 0; i < len(src); i += 2 {
		s := src[i+1]
		if s == nil {
			continue
		}
		name, ok := src[i].([]byte)
		if !ok {
			return errors.New("redigo: ScanStruct key not a bulk string value")
		}
		fs := ss.fieldSpec(name)
		if fs == nil {
			continue
		}

		field := d.FieldByIndex(fs.index)

		if needJson(field) {
			if err := FromDB(s, field); err != nil {
				return err
			}
		} else {
			if err := convertAssignValue(field, s); err != nil {
				return err
			}
		}
	}
	return nil
}

func scanMap(src []interface{}, rv reflect.Value) error {
	valType := rv.Type()
	valKeyType := valType.Key()
	valElemType := valType.Elem()

	mapType := reflect.MapOf(valKeyType, valElemType)
	valMap := reflect.MakeMap(mapType)
	var err error
	for len(src) > 0 {
		key := reflect.New(valKeyType)
		elem := reflect.New(valElemType)
		src, err = redis.Scan(src, key.Interface(), elem.Interface())
		if err != nil {
			return err
		}
		valMap.SetMapIndex(reflect.Indirect(key), reflect.Indirect(elem))
	}
	rv.Set(valMap)
	return nil
}

func dirtyValue(v interface{}) interface{} {
	switch reflect.ValueOf(v).Interface().(type) {
	case string:
		s, _ := v.(string)
		r := md5.Sum([]byte(s))
		return string(r[:])
	case []byte:
		by, _ := v.([]byte)
		r := md5.Sum(by)
		return string(r[:])
	}
	r := md5.Sum([]byte(fmt.Sprintf("%v", v)))
	return string(r[:])
}
