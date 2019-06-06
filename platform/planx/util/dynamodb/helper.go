package dynamodb

import (
	"fmt"
	"strconv"

	"vcs.taiyouxi.net/platform/planx/util/logs"

	"github.com/aws/aws-sdk-go/aws"
	DDB "github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	ValueType = "value"
)

// Create new AttributeValue from the type of value
func CreateAttributeValue(v interface{}) *DDB.AttributeValue {
	switch t := v.(type) {
	case string:
		return &DDB.AttributeValue{
			S: aws.String(t),
		}
	case int, int32, int64, uint, uint32, uint64, float32, float64:
		return &DDB.AttributeValue{
			N: aws.String(fmt.Sprint(t)),
		}
	default:
		return &DDB.AttributeValue{}
	}
}

// Retrieve value from DynamoDB type
func GetItemValue(val *DDB.AttributeValue) interface{} {
	if val == nil {
		return nil
	}
	switch {
	case val.N != nil:
		data, _ := strconv.ParseInt(*val.N, 10, 64)
		return data
	case val.S != nil && val.S != nil:
		return *val.S
	case val.BOOL != nil && val.BOOL != nil:
		return *val.BOOL
	case val.B != nil && len(val.B) > 0:
		return val.B
	case len(val.M) > 0:
		return val.M
	case val.NS != nil && len(val.NS) > 0:
		var data []int64
		for _, vString := range val.NS {
			var vInt int64 = 0
			if vString != nil {
				vInt, _ = strconv.ParseInt(*vString, 10, 64)
			}
			data = append(data, vInt)
		}
		return data
	case val.SS != nil && len(val.SS) > 0:
		var data []string
		for _, vString := range val.SS {
			if vString != nil {
				data = append(data, "")
			} else {
				data = append(data, *vString)
			}
		}
		return data
	case val.BS != nil && len(val.BS) > 0:
		var data [][]byte
		for _, vBytes := range val.BS {
			data = append(data, vBytes)
		}
		return data
	case val.L != nil && len(val.L) > 0:
		var data []interface{}
		for _, vAny := range val.L {
			if vAny != nil {
				data = append(data, GetItemValue(vAny))
			} else {
				data = append(data, nil)
			}
		}
		return data
	}
	return nil
}

// Convert DynamoDB Item to map data
func Unmarshal(item map[string]*DDB.AttributeValue) map[string]interface{} {
	data := make(map[string]interface{})
	for key, val := range item {
		data[key] = GetItemValue(val)
	}
	return data
}

//
func GetFromAnys2String(typ, name string, data map[string]interface{}) (string, bool) {
	data_any, name_ok := data[name]
	if !name_ok {
		//logs.Error("getFromAnys2String %s Err by no %s", typ, name)
		return "", false
	}

	data_str, ok := data_any.(string)
	if !ok {
		logs.Warn("getFromAnys2String %s Err by %s no string", typ, name)
		return "", false
	}

	return data_str, true
}

//
func GetFromAnys2Bytes(typ, name string, data map[string]interface{}) ([]byte, bool) {
	data_any, name_ok := data[name]
	if !name_ok {
		//logs.Error("GetFromAnys2Bytes %s Err by no %s", typ, name)
		return []byte{}, false
	}

	data_bytes, ok := data_any.([]byte)
	if !ok {
		logs.Warn("GetFromAnys2Bytes %s Err by %s no byte", typ, name)
		return []byte{}, false
	}

	return data_bytes, true
}

//
func GetFromAnys2int64(typ, name string, data map[string]interface{}) (int64, bool) {
	data_any, name_ok := data[name]
	if !name_ok {
		//logs.Error("getFromAnys2int64 %s Err by no %s", typ, name)
		return 0, false
	}

	data_int64, ok := data_any.(int64)
	if !ok {
		logs.Warn("getFromAnys2int64 %s Err by %s no string", typ, name)
		return 0, false
	}

	return data_int64, true
}

func printDynamoDBBatchCapacity(loginfo string, resp *DDB.BatchWriteItemOutput) {
	if resp != nil && resp.ConsumedCapacity != nil && len(resp.ConsumedCapacity) > 0 {
		var allcap float64
		for _, c := range resp.ConsumedCapacity {
			allcap += *c.CapacityUnits
		}
		logs.Warn("%s CapacityUnits %f", loginfo, allcap)
	}
}
