package dynamodb

import (
	"errors"
	"os"

	"fmt"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	DDB "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/cenk/backoff"
	"vcs.taiyouxi.net/platform/planx/util/awshelper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type DynamoDB struct {
	client        *DDB.DynamoDB
	noRetryClient *DDB.DynamoDB
	//TODO: YZH DynamoDB table_des的存在很奇怪
	table_des map[string]TableInfo
}

func (d *DynamoDB) Client() *DDB.DynamoDB {
	return d.client
}

func (d *DynamoDB) Connect(
	region,
	accessKeyID,
	secretAccessKey,
	sessionToken string) error {

	if region == "" {
		region = os.Getenv("AWS_REGION")
	}
	// For https://github.com/aws/aws-sdk-ruby/issues/243
	mySession := awshelper.CreateAWSSession(region, accessKeyID, secretAccessKey, 3)
	d.client = DDB.New(mySession)

	noRetrySession := awshelper.CreateAWSSession(region, accessKeyID, secretAccessKey, 0)
	d.noRetryClient = DDB.New(noRetrySession)

	logs.Info("DynamoDB Connect ok")
	return nil
}

func (d *DynamoDB) Set(table_name string,
	hash, range_value, value interface{}) error {
	hash_typ, range_typ, ok := d.GetHashType(table_name)
	if !ok {
		return errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	put_input := &DDB.PutItemInput{
		TableName: aws.String(table_name),
		Item: map[string]*DDB.AttributeValue{
			hash_typ:  CreateAttributeValue(hash),
			range_typ: CreateAttributeValue(range_value),
			ValueType: CreateAttributeValue(value),
		},
		ReturnConsumedCapacity: aws.String("TOTAL"),
		Expected:               map[string]*DDB.ExpectedAttributeValue{},
	}
	_, err := d.client.PutItem(put_input)
	return err
}

func (d *DynamoDB) Get(table_name string,
	hash, range_value interface{}) (map[string]interface{}, error) {
	hash_typ, range_typ, ok := d.GetHashType(table_name)
	if !ok {
		return nil, errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	get_item := &DDB.GetItemInput{
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(table_name),
		Key: map[string]*DDB.AttributeValue{
			hash_typ:  CreateAttributeValue(hash),
			range_typ: CreateAttributeValue(range_value),
		},
	}
	get_item_out, err := d.client.GetItem(get_item)
	if err != nil {
		return nil, err
	} else {
		re := make(map[string]interface{}, len(get_item_out.Item))
		for i, v := range get_item_out.Item {
			re[i] = GetItemValue(v)
		}
		return re, nil
	}
}

func (d *DynamoDB) SetByHash(table_name string, hash, value interface{}) error {
	hash_typ, _, ok := d.GetHashType(table_name)
	if !ok {
		return errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	put_input := &DDB.PutItemInput{
		TableName: aws.String(table_name),
		Item: map[string]*DDB.AttributeValue{
			hash_typ:  CreateAttributeValue(hash),
			ValueType: CreateAttributeValue(value),
		},
		ReturnConsumedCapacity: aws.String("TOTAL"),
		Expected:               map[string]*DDB.ExpectedAttributeValue{},
	}
	_, err := d.client.PutItem(put_input)
	return err
}

func (d *DynamoDB) SetByHashB(table_name string, hash interface{}, value []byte) error {
	hash_typ, _, ok := d.GetHashType(table_name)
	if !ok {
		return errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	logs.Warn("SetByHashB %v", value)

	put_input := &DDB.PutItemInput{
		TableName: aws.String(table_name),
		Item: map[string]*DDB.AttributeValue{
			hash_typ: CreateAttributeValue(hash),
			ValueType: &DDB.AttributeValue{
				B: value,
			},
		},
		ReturnConsumedCapacity: aws.String("TOTAL"),
		Expected:               map[string]*DDB.ExpectedAttributeValue{},
	}
	_, err := d.client.PutItem(put_input)
	return err
}

func (d *DynamoDB) DelKey(table_name string, hash interface{}) error {
	hash_typ, _, ok := d.GetHashType(table_name)
	if !ok {
		return errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	logs.Warn("Del Key %v from %s", hash, table_name)

	put_input := &DDB.DeleteItemInput{
		TableName: aws.String(table_name),
		Key: map[string]*DDB.AttributeValue{
			hash_typ: CreateAttributeValue(hash),
		},
		ReturnConsumedCapacity: aws.String("TOTAL"),
		Expected:               map[string]*DDB.ExpectedAttributeValue{},
	}
	_, err := d.client.DeleteItem(put_input)
	return err
}

func (d *DynamoDB) DelKeyRange(table_name string, hash interface{}, rangeKey interface{}) error {
	hash_typ, ramge_typ, ok := d.GetHashType(table_name)
	if !ok {
		return errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	logs.Warn("Del Key %v from %s", hash, table_name)

	put_input := &DDB.DeleteItemInput{
		TableName: aws.String(table_name),
		Key: map[string]*DDB.AttributeValue{
			hash_typ:  CreateAttributeValue(hash),
			ramge_typ: CreateAttributeValue(rangeKey),
		},
		ReturnConsumedCapacity: aws.String("TOTAL"),
		Expected:               map[string]*DDB.ExpectedAttributeValue{},
	}
	_, err := d.client.DeleteItem(put_input)
	return err
}

// 上层业务要遵守dynamodb batch write某些限制条件 , not support bool 暂时
func (d *DynamoDB) BatchSetByHashM(table_name string, hash interface{}, values []map[string]interface{}) (error, int) {
	hash_typ, _, ok := d.GetHashType(table_name)
	if !ok {
		return errors.New(fmt.Sprintf("No table Info, %s", table_name)), len(values)
	}
	PutRequests := make([]*DDB.WriteRequest, 0, len(values))
	for _, value := range values {
		item := make(map[string]*DDB.AttributeValue, len(value))
		item[hash_typ] = CreateAttributeValue(hash)
		for typ, v := range value {
			item[typ] = CreateAttributeValue(v)
		}
		PutRequests = append(PutRequests, &DDB.WriteRequest{
			PutRequest: &DDB.PutRequest{
				Item: item,
			},
		})
	}
	params := &DDB.BatchWriteItemInput{
		RequestItems: map[string][]*DDB.WriteRequest{ // Required
			table_name: PutRequests,
		},
		ReturnConsumedCapacity: aws.String("TOTAL"),
	}

	resp, err := d.client.BatchWriteItem(params)
	if err != nil {
		logs.Error("BatchWriteItem Err by %s", err.Error())
		return err, len(values)
	}
	if len(resp.UnprocessedItems) > 0 {
		b := NewExponentialBackOffSleepFirst()
		err = backoff.RetryNotify(func() error {
			newPutRequests := PutRequests[0:0]
			for _, v := range resp.UnprocessedItems {
				for _, failed := range v {
					if failed.PutRequest != nil || failed.DeleteRequest != nil {
						newPutRequests = append(newPutRequests, failed)
					}
				}
			}
			nparams := &DDB.BatchWriteItemInput{
				RequestItems: map[string][]*DDB.WriteRequest{ // Required
					table_name: newPutRequests,
				},
				ReturnConsumedCapacity: aws.String("TOTAL"),
			}
			nresp, nerr := d.client.BatchWriteItem(nparams)
			resp = nresp
			if nerr != nil {
				return nerr
			}

			if len(resp.UnprocessedItems) > 0 {
				return fmt.Errorf("UnprocessedItems")
			}
			return nil
		}, b, func(e error, d time.Duration) {
		})
		return err, len(resp.UnprocessedItems)
	}
	return nil, 0
}

func (d *DynamoDB) SetByHashM(table_name string, hash interface{},
	values map[string]interface{}) error {
	hash_typ, _, ok := d.GetHashType(table_name)
	if !ok {
		return errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	items := make(map[string]*DDB.AttributeValue, len(values)+1)
	items[hash_typ] = CreateAttributeValue(hash)

	for typ, v := range values {
		items[typ] = CreateAttributeValue(v)
	}

	put_input := &DDB.PutItemInput{
		TableName: aws.String(table_name),
		Item:      items,
		ReturnConsumedCapacity: aws.String("TOTAL"),
		Expected:               map[string]*DDB.ExpectedAttributeValue{},
	}
	_, err := d.client.PutItem(put_input)
	return err
}

func (d *DynamoDB) SetByHashM_IfNoExist(table_name string, hash interface{},
	values map[string]interface{}) error {
	hash_typ, _, ok := d.GetHashType(table_name)
	if !ok {
		return errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	items := make(map[string]*DDB.AttributeValue, len(values)+1)
	items[hash_typ] = CreateAttributeValue(hash)

	for typ, v := range values {
		items[typ] = CreateAttributeValue(v)
	}

	put_input := &DDB.PutItemInput{
		TableName: aws.String(table_name),
		Item:      items,
		ReturnConsumedCapacity: aws.String("TOTAL"),
		//Expected:               map[string]*DDB.ExpectedAttributeValue{},

		ExpressionAttributeNames: map[string]*string{
			"#hash_typ": aws.String(fmt.Sprintf("%s", hash_typ)),
		},
		ConditionExpression: aws.String("attribute_not_exists(#hash_typ)"),
	}
	_, err := d.client.PutItem(put_input)
	return err
}

func (d *DynamoDB) GetByHashM(table_name string, hash interface{}) (map[string]interface{}, error) {
	hash_typ, _, ok := d.GetHashType(table_name)
	if !ok {
		return nil, errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	get_item := &DDB.GetItemInput{
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(table_name),
		Key: map[string]*DDB.AttributeValue{
			hash_typ: CreateAttributeValue(hash),
		},
	}
	get_item_out, err := d.client.GetItem(get_item)
	if err != nil {
		return nil, err
	} else {
		re := make(map[string]interface{}, len(get_item_out.Item))
		for i, v := range get_item_out.Item {
			re[i] = GetItemValue(v)
		}
		return re, nil
	}
}

func (d *DynamoDB) GetByHash(table_name string, hash interface{}) (interface{}, error) {
	hash_typ, _, ok := d.GetHashType(table_name)
	if !ok {
		return "", errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	get_item := &DDB.GetItemInput{
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(table_name),
		Key: map[string]*DDB.AttributeValue{
			hash_typ: CreateAttributeValue(hash),
		},
	}
	get_item_out, err := d.client.GetItem(get_item)
	if get_item_out == nil || get_item_out.Item == nil {
		return nil, err
	}
	items := get_item_out.Item

	//logs.Warn("GetByHash %v %v %v", items[ValueType].B, table_name, hash_typ)

	return GetItemValue(items[ValueType]), err
}

func (d *DynamoDB) QueryByHash(table_name string,
	hash interface{}) ([]interface{}, error) {

	hash_typ, _, ok := d.GetHashType(table_name)
	if !ok {
		return []interface{}{}, errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	query_keys := make(map[string]*DDB.Condition)
	query_keys[hash_typ] = &DDB.Condition{
		AttributeValueList: []*DDB.AttributeValue{CreateAttributeValue(hash)},
		ComparisonOperator: aws.String("EQ"),
	}

	in := &DDB.QueryInput{
		TableName:     aws.String(table_name),
		KeyConditions: query_keys,
	}

	query_out, err := d.client.Query(in)

	re := make([]interface{}, 0, len(query_out.Items))

	for _, value := range query_out.Items {
		re = append(re, GetItemValue(value[ValueType]))

	}
	return re[:], err
}

func (d *DynamoDB) UpdateByHash(
	table_name string,
	hash interface{},
	values map[string]interface{}) error {

	hash_typ, _, ok := d.GetHashType(table_name)
	if !ok {
		return errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	updates := make(map[string]*DDB.AttributeValueUpdate, len(values))
	for typ, v := range values {
		v_a := CreateAttributeValue(v)
		updates[typ] = &DDB.AttributeValueUpdate{
			Action: aws.String("PUT"),
			Value:  v_a,
		}
	}

	update_item := &DDB.UpdateItemInput{
		AttributeUpdates: updates,
		TableName:        aws.String(table_name),
		Key: map[string]*DDB.AttributeValue{
			hash_typ: CreateAttributeValue(hash),
		},
		ReturnConsumedCapacity: aws.String("TOTAL"),
		Expected:               map[string]*DDB.ExpectedAttributeValue{},
		ReturnValues:           aws.String("UPDATED_NEW"),
	}
	_, err := d.client.UpdateItem(update_item)
	return err
}

func (d *DynamoDB) Incr(table_name string,
	hash interface{},
	item_type string,
	value int64) (int64, error) {
	hash_typ, _, ok := d.GetHashType(table_name)
	if !ok {
		return -1, errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	update_input := &DDB.UpdateItemInput{
		TableName: aws.String(table_name),
		Key: map[string]*DDB.AttributeValue{
			hash_typ: CreateAttributeValue(hash),
		},
		ExpressionAttributeNames: map[string]*string{
			"#Q": &item_type,
		},
		ExpressionAttributeValues: map[string]*DDB.AttributeValue{
			":incr": CreateAttributeValue(value),
		},
		UpdateExpression: aws.String("SET #Q = #Q + :incr"),

		ReturnConsumedCapacity: aws.String("TOTAL"),
		ReturnValues:           aws.String("UPDATED_NEW"),
	}
	update_output, err := d.client.UpdateItem(update_input)
	if err != nil {
		return -1, err
	}
	v, ok := update_output.Attributes[item_type]
	if !ok {
		return -2, errors.New("No type Return")
	}
	v_any := GetItemValue(v)
	v_int, ok := v_any.(int64)
	if !ok {
		return -3, errors.New("type not int64")
	}
	return v_int, nil
}

type ScanHander func(idx int, key interface{}, data interface{}) error

func (d *DynamoDB) rescan(ticker *backoff.Ticker, params *DDB.ScanInput) (*DDB.ScanOutput, error) {
	var err error
	for _ = range ticker.C {
		output, err := d.client.Scan(params)
		if err == nil {
			ticker.Stop()
			return output, err
		}
	}

	ticker.Stop()
	return nil, err
}

func (d *DynamoDB) Scan(name string, scan_len int64, hander ScanHander, b *backoff.ExponentialBackOff) error {
	hash_typ, _, ok := d.GetHashType(name)
	if !ok {
		return errors.New(fmt.Sprintf("No table Info, %s", name))
	}

	var last_key map[string]*DDB.AttributeValue
	var all int = 0
	for {
		params := &DDB.ScanInput{
			TableName: aws.String(name),
			AttributesToGet: []*string{
				aws.String(hash_typ),
				aws.String(ValueType),
			},

			Limit: aws.Int64(scan_len),
		}
		if last_key != nil {
			params.ExclusiveStartKey = last_key
		}
		output, err := d.client.Scan(params)
		if err != nil {
			ticker := backoff.NewTicker(b)
			output, err = d.rescan(ticker, params)
		}

		if err != nil {
			return err
		}

		for _, item := range output.Items {
			all++
			key := GetItemValue(item[hash_typ])
			value := GetItemValue(item[ValueType])

			herr := hander(all, key, value)
			if herr != nil {
				return herr
			}
		}

		last_key = output.LastEvaluatedKey
		if last_key == nil {
			break
		}

	}

	return nil
}

type DynamoKV struct {
	K interface{}
	V interface{}
}

// need to go ParallelScan()
func (d *DynamoDB) ParallelScan(
	name string,
	scan_len, scan_idx, scan_all int64,
	channel chan DynamoKV,
	b *backoff.ExponentialBackOff) error {

	hash_typ, _, ok := d.GetHashType(name)
	if !ok {
		return errors.New(fmt.Sprintf("No table Info, %s", name))
	}

	var last_key map[string]*DDB.AttributeValue
	var all int = 0
	for {

		params := &DDB.ScanInput{
			TableName: aws.String(name),
			AttributesToGet: []*string{
				aws.String(hash_typ),
				aws.String(ValueType),
			},

			Limit:         aws.Int64(scan_len),
			Segment:       aws.Int64(scan_idx),
			TotalSegments: aws.Int64(scan_all),
		}

		if last_key != nil {
			params.ExclusiveStartKey = last_key
		}

		output, err := d.client.Scan(params)
		if err != nil {
			ticker := backoff.NewTicker(b)
			output, err = d.rescan(ticker, params)
		}

		if err != nil {
			return err
		}

		for _, item := range output.Items {
			all++
			key := GetItemValue(item[hash_typ])
			value := GetItemValue(item[ValueType])

			channel <- DynamoKV{
				key,
				value,
			}
		}

		last_key = output.LastEvaluatedKey
		if last_key == nil {
			break
		}

	}

	return nil
}
