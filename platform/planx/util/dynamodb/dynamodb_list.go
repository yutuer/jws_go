package dynamodb

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	DDB "github.com/aws/aws-sdk-go/service/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//TODO: YZH 这里扫描了所有的数据表,很奇怪
type TableInfo struct {
	Name    string
	Hash_t  string
	Range_t string
}

func newTableInfo(des *DDB.TableDescription) TableInfo {
	t := TableInfo{}
	t.Name = *des.TableName
	for _, v := range des.KeySchema {
		if *v.KeyType == "HASH" {
			t.Hash_t = *v.AttributeName
		} else if *v.KeyType == "RANGE" {
			t.Range_t = *v.AttributeName
		}
	}
	return t
}

func (d *DynamoDB) IsTableHasCreate(name string) bool {
	_, ok := d.table_des[name]
	return ok
}

func (d *DynamoDB) CreateTable(name, hash_t, range_t string) error {
	if d.IsTableHasCreate(name) {
		// 已经有的表就不创建了
		return nil
	}

	cti := DDB.CreateTableInput{
		TableName: aws.String(name),
	}

	cti.KeySchema = []*DDB.KeySchemaElement{
		&DDB.KeySchemaElement{
			AttributeName: aws.String(hash_t),
			KeyType:       aws.String("HASH")},
		&DDB.KeySchemaElement{
			AttributeName: aws.String(range_t),
			KeyType:       aws.String("RANGE")},
	}

	cti.AttributeDefinitions = []*DDB.AttributeDefinition{
		&DDB.AttributeDefinition{
			AttributeName: aws.String(hash_t),
			AttributeType: aws.String("N")},
		&DDB.AttributeDefinition{
			AttributeName: aws.String(range_t),
			AttributeType: aws.String("N")},
	}

	cti.ProvisionedThroughput = &DDB.ProvisionedThroughput{
		ReadCapacityUnits:  aws.Int64(10),
		WriteCapacityUnits: aws.Int64(10),
	}

	_, err := d.client.CreateTable(&cti)
	return err
}

func (d *DynamoDB) listTables() ([]*string, error) {
	list_table_input := &DDB.ListTablesInput{}
	list_table_output, err := d.client.ListTables(list_table_input)
	return list_table_output.TableNames[:], err
}

func (d *DynamoDB) describeTable(table_name string) error {
	des_table_input := &DDB.DescribeTableInput{
		TableName: aws.String(table_name),
	}
	des_table_output, err := d.client.DescribeTable(des_table_input)
	if err == nil {
		d.table_des[table_name] = newTableInfo(des_table_output.Table)
		logs.Debug("DynamoDB describeTable %s", table_name)
	}
	return err
}

func (d *DynamoDB) GetHashType(table_name string) (string, string, bool) {
	des, ok := d.table_des[table_name]
	if !ok {
		return "", "", false
	}

	return des.Hash_t, des.Range_t, true
}

func (d *DynamoDB) CreateHashTable(name, hash, hash_typ string) error {
	_, ok := d.table_des[name]
	if ok {
		// 已经有的表就不创建了
		return nil
	}

	cti := DDB.CreateTableInput{
		TableName: aws.String(name),
	}

	cti.KeySchema = []*DDB.KeySchemaElement{
		&DDB.KeySchemaElement{
			AttributeName: aws.String(hash),
			KeyType:       aws.String("HASH")},
	}

	cti.AttributeDefinitions = []*DDB.AttributeDefinition{
		&DDB.AttributeDefinition{
			AttributeName: aws.String(hash),
			AttributeType: aws.String(hash_typ)},
	}

	//TODO 手动建表应该支持配置

	cti.ProvisionedThroughput = &DDB.ProvisionedThroughput{
		ReadCapacityUnits:  aws.Int64(10),
		WriteCapacityUnits: aws.Int64(10),
	}

	_, err := d.client.CreateTable(&cti)
	return err
}

func (d *DynamoDB) InitTable() error {
	tables, err := d.listTables()
	logs.Debug("tables: %v", tables)
	if err != nil {
		return err
	}
	d.table_des = make(map[string]TableInfo, len(tables))

	for _, n := range tables {
		err = d.describeTable(*n)
		if err != nil {
			return err
		}
	}
	return nil
}

// InitTableAndCheck 检查数据库的存在性,如果传入参数都是空字符串
// 例如: []string{"",""},则刚好不会报错
func (d *DynamoDB) InitTableAndCheck(table []string) error {
	if table == nil {
		return nil
	}
	//tables, err := d.listTables()
	//if err != nil {
	//	return err
	//}
	tables := table
	if d.table_des == nil {
		d.table_des = make(map[string]TableInfo, len(tables))
	}
	isExist := make(map[string]bool, len(table))
	for _, t := range table {
		if t != "" {
			isExist[t] = false
		}
	}

	for _, n := range tables {
		err := d.describeTable(n)
		if err != nil {
			return err
		}
		if _, ok := isExist[n]; ok {
			isExist[n] = true
		}
	}
	for _, b := range isExist {
		if !b {
			return fmt.Errorf("DynamoDB Init, table %s not exist", table)
		}
	}
	return nil
}
