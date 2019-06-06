package main

//GOOS=linux GOARCH=amd64  go build dynamodb.go
//scp dynamodb ec2-user@10.222.0.246:~/
import (
	"fmt"
	"log"

	"time"

	"os"

	"encoding/csv"
	"sync"

	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gocarina/gocsv"
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

type PayItem struct {
	AccountName      string `json:"account_name,omitempty"`
	Channel          string `json:"channel,omitempty"`
	ChannelUID       string `json:"channel_uid,omitempty"`
	ExtraParams      string `json:"extras_params,omitempty"`
	GoodIndex        string `json:"good_idx"`
	GoodName         string `json:"good_name"`
	HcBuy            int64  `json:"hc_buy"`
	HcGive           int64  `json:"hc_give"`
	IsTest           string `json:"is_test,omitempty"`
	Mobile           string `json:"mobile"`
	MoneyAmount      string `json:"money_amount"`
	OrderNo          string `json:"order_no"`
	PayTime          string `json:"pay_time"`
	PayTimestamp     string `json:"pay_time_s"`
	Platform         string `json:"platform"`
	ProductID        string `json:"product_id"`
	ReceiveTimestamp int64  `json:"receiveTimestamp"`
	RoleName         string `json:"role_name"`
	SN               int64  `json:"sn"`
	Status           string `json:"status,omitempty"`
	TiStatus         string `json:"tistatus,omitempty"`
	Uid              string `json:"uid,omitempty"`
	Version          string `json:"ver,omitempty"`
}

func main() {
	flag.Parse()
	ddb := CreateAWSSession("cn-north-1",
		"",
		"",
		3)
	client := dynamodb.New(ddb)
	targetdb := flag.Arg(0)
	d := dynamodbattribute.NewDecoder(func(d *dynamodbattribute.Decoder) {
		d.UseNumber = true
	})
	file, err := os.Create(flag.Arg(1))
	if err != nil {
		panic(err)
	}
	w := csv.NewWriter(file)
	var wg sync.WaitGroup
	payChan := make(chan interface{}, 64)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := gocsv.MarshalChan(payChan, w)
		if err != nil {
			panic(err)
		}
	}()
	l := log.New(os.Stderr, "", 0)
	var last_key map[string]*dynamodb.AttributeValue
	var total int64
	for {
		so, err := client.Scan(&dynamodb.ScanInput{
			ReturnConsumedCapacity: aws.String("TOTAL"),
			TableName:              aws.String(targetdb),
			ExclusiveStartKey:      last_key,
			Limit:                  aws.Int64(1000),
		})
		if err != nil {
			fmt.Println("Scan Error, %v", err)
			os.Exit(1)
		}
		//fmt.Println("consume:", so.ConsumedCapacity.GoString(), "count:", *so.ScannedCount)

		for _, i := range so.Items {
			var v PayItem
			errum := d.Decode(&dynamodb.AttributeValue{M: i}, &v)
			if errum == nil {
				payChan <- v
				total++
				//s, _ := json.Marshal(v)
				//fmt.Println(string(s))
			} else {
				panic(errum)
			}
		}

		last_key = so.LastEvaluatedKey
		if last_key == nil {
			break
		}

		//time.Sleep(time.Second * 1)
	}

	l.Println("all:", total)
	close(payChan)
	wg.Wait()
	time.Sleep(time.Second * 1)
}
