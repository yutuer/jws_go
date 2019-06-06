package main

import (
	"fmt"
	"math"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	//"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

func SetServiceRetry(s *aws.Service) {
	s.DefaultMaxRetries = 2 //这里修改了service中的默认重试次数

	//返回一个重试时间间隔
	s.RetryRules = func(r *aws.Request) time.Duration {
		fmt.Println("yzh retry", r.RetryCount)
		delay := time.Duration(math.Pow(2, float64(r.RetryCount))) * 50
		return delay * time.Millisecond
	}
}

func main() {
	config := *aws.DefaultConfig

	//id := "AKIAO6TYDXICE34WUFAQ"
	//secret := "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"
	//NewCredentialInfo := credentials.NewStaticCredentials(id, secret, "")
	//config.Credentials = NewCredentialInfo

	//config.LogLevel = 1

	//config.MaxRetries = 3 //影响默认service代码中的重试次数

	config.Region = "cn-north-1"
	k := kinesis.New(&config)
	SetServiceRetry(k.Service)

	fmt.Println("###########")
	//params := &kinesis.ListStreamsInput{
	//ExclusiveStartStreamName: aws.String("dev-log"),
	//Limit: aws.Long(1),
	//}
	lso, err := k.ListStreams(nil)
	if err != nil {
		fmt.Printf("error: %s", err.Error())
	}

	fmt.Println(lso.String())
	paramsit := &kinesis.GetShardIteratorInput{
		ShardID:           aws.String("shardId-000000000000"), // Required
		ShardIteratorType: aws.String("LATEST"),               // Required
		StreamName:        aws.String("dev-dev"),              // Required
		//StartingSequenceNumber: aws.String("SequenceNumber"),
	}
	respit, errit := k.GetShardIterator(paramsit)
	{
		err := errit
		resp := respit
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				// Generic AWS error with Code, Message, and original error (if any)
				fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
				if reqErr, ok := err.(awserr.RequestFailure); ok {
					// A service error occurred
					fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
				}
			} else {
				// This case should never be hit, the SDK should always return an
				// error which satisfies the awserr.Error interface.
				fmt.Println(err.Error())
			}
		}

		// Pretty-print the response data.
		fmt.Println(awsutil.StringValue(resp))
	}

	{
		params := &kinesis.GetRecordsInput{
			ShardIterator: respit.ShardIterator, // Required
			Limit:         aws.Long(2),
		}
		for {
			time.Sleep(1 * time.Millisecond)
			resp, err := k.GetRecords(params)

			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					// Generic AWS error with Code, Message, and original error (if any)
					fmt.Println(awsErr.Code(), "**", awsErr.Message(), awsErr.OrigErr())
					if reqErr, ok := err.(awserr.RequestFailure); ok {
						// A service error occurred
						fmt.Println(reqErr.Code(), "*", reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
					}
				} else {
					// This case should never be hit, the SDK should always return an
					// error which satisfies the awserr.Error interface.
					fmt.Println("#", err.Error())
				}
				break
			}
			if resp != nil {
				params.ShardIterator = resp.NextShardIterator
				// Pretty-print the response data.
				for _, r := range resp.Records {
					fmt.Println(string(r.Data))
				}
			}
		}
	}
}
