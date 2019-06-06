package main

//GOOS=linux GOARCH=amd64  go build dynamodb.go
//scp dynamodb ec2-user@10.222.0.246:~/
import (
	"fmt"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/tipay/dynamopay"
)

func CreateAWSSession(region, accessKey, secretKey string, MaxRetries int) *session.Session {
	config := defaults.Config().WithRegion(region).WithMaxRetries(MaxRetries)
	handler := defaults.Handlers()

	providers := []credentials.Provider{
		&credentials.StaticProvider{Value: credentials.Value{
			AccessKeyID:     accessKey,
			SecretAccessKey: secretKey,
			SessionToken:    "",
		}},
		&credentials.EnvProvider{},
		&credentials.SharedCredentialsProvider{Filename: "", Profile: ""},
		defaults.RemoteCredProvider(*config, handler),
	}

	if accessKey == "" ||
		secretKey == "" {
		providers = providers[1:]
	}

	creds := credentials.NewChainCredentials(providers)
	//TODO: For FanYang by YZH， WithDisableComputeChecksums还是必须的吗？
	config = config.
		WithCredentials(creds).
		WithDisableSSL(true).
		WithDisableComputeChecksums(true)

	return session.New(config)

}

type TableInfo struct {
	Name    string
	Hash_t  string
	Range_t string
}

func newTableInfo(des *dynamodb.TableDescription) TableInfo {
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

func InitTableAndCheck(d *dynamodb.DynamoDB, table []string) error {
	tables := table
	//tl, err := d.ListTables(&dynamodb.ListTablesInput{})
	//tables := tl.TableNames[:]
	//if err != nil {
	//	return err
	//}
	//
	table_des := make(map[string]TableInfo, len(tables))

	isExist := make(map[string]bool, len(table))
	for _, t := range table {
		if t != "" {
			isExist[t] = false
		}
	}

	for _, n := range tables {
		table_name := n
		des_table_input := &dynamodb.DescribeTableInput{
			TableName: aws.String(table_name),
		}
		des_table_output, err := d.DescribeTable(des_table_input)

		if err != nil {
			return err
		}
		table_des[table_name] = newTableInfo(des_table_output.Table)
		if _, ok := isExist[table_name]; ok {
			isExist[table_name] = true
		}
	}
	for _, b := range isExist {
		if !b {
			return fmt.Errorf("DynamoDB Init, table %s not exist", table)
		}
	}
	fmt.Println(isExist)
	return nil
}

func setIfNoExists() {
	dbPayMgr := dynamopay.NewPayDynamoDB("LocalAndroidPay",
		"cn-north-1",
		"AKIAPLNPSQYENX3LGB5A",
		"JlXWCZV24sYsLw1+yigXE/I+zgfV4PLkJbyvfFn+")

	err := dbPayMgr.Open()
	if err != nil {
		fmt.Println("Open  err %v", err)
		return
	}
	value := make(map[string]interface{}, 11)
	value["channel"] = "channel"         // 渠道标示ID
	value["channel_uid"] = "channel_uid" // 渠道用户唯一标示,该值从客户端GetUserId()中可获取
	value["good_idx"] = "good_idx"       // 游戏在调用QucikSDK发起支付时传递的游戏方订单,这里会原样传回
	value["order_no"] = "123456"         // 天象唯一订单号
	value["pay_time"] = "qqq"            // 支付时间
	value["money_amount"] = 12           // 成交金额
	value["status"] = 1                  // 充值状态 0 成功 1失败(为1时 应返回FAILUD失败)
	value["product_id"] = 1
	value["ver"] = "v"
	value["extras_params"] = "ex"
	value["tistatus"] = "er" // 通知状态，paid支付成功；delivered支付成功并玩家已拿到
	value["uid"] = "uid"     // accountid
	value["mobile"] = "mobile"
	value["is_test"] = "is_test"
	value["note"] = "note"

	err = dbPayMgr.SetByHashM_IfNoExist("123456", value)
	if err != nil {
		fmt.Println("SetByHashM 1 err %v", err)
		return
	}
	value["receiveTimestamp"] = 3345
	err = dbPayMgr.SetByHashM_IfNoExist("123456", value)
	if err != nil {
		fmt.Println("SetByHashM 2 err %v", err)
		return
	}
}
func main() {
	ddb := CreateAWSSession("cn-north-1",
		"AKIAPLNPSQYENX3LGB5A",
		"JlXWCZV24sYsLw1+yigXE/I+zgfV4PLkJbyvfFn+", 3)
	client := dynamodb.New(ddb)

	tl, err := client.ListTables(&dynamodb.ListTablesInput{})

	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Println(tl)

	for {
		err := InitTableAndCheck(client, []string{"RedeemCode"})
		fmt.Println("InitTableAndCheck %v", err)

		out, err1 := client.GetItem(&dynamodb.GetItemInput{
			ConsistentRead: aws.Bool(true),
			TableName:      aws.String("RedeemCode"),
			Key: map[string]*dynamodb.AttributeValue{
				"code": &dynamodb.AttributeValue{
					S: aws.String("LXQLBFQFEHNN"),
				},
			},
		})
		fmt.Println("GetItem %v, %v", out, err1)
		time.Sleep(time.Second * 2)
	}

}
