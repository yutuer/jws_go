package storehelper

import (
	//"github.com/aws/aws-sdk-go/aws"
	//"github.com/aws/aws-sdk-go/service/s3"
	"testing"
	//"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	S3_Bucket     = "redis-onland"
	AWS_Region    = "cn-north-1"
	AWS_AccessKey = "AKIAPLNPSQYENX3LGB5A"
	AWS_SecretKey = "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"
)

func TestGet(t *testing.T) {
	s3 := NewStoreS3(AWS_Region, S3_Bucket, AWS_AccessKey, AWS_SecretKey, "", "")
	res, err := s3.Get("2015/04/23/bag:0:0:1")
	if err != nil {
		t.Errorf("Get Err %s", err.Error())
		return
	}
	t.Logf("Info %v", string(res))
}

func TestS3GetAll(t *testing.T) {
	//t.Logf("Start")
	//conn := NewStoreS3(AWS_Region, S3_Bucket, AWS_AccessKey, AWS_SecretKey, "", "")
	//marker := ""
	//turn := 0
	//all := 0
	//for {
	//	logs.Info("Start %d", turn)
	//	params := &s3.ListObjectsInput{
	//		Bucket: aws.String(S3_Bucket), // Required
	//		//Delimiter:    aws.String("Delimiter"),
	//		//EncodingType: aws.String("EncodingType"),
	//		Marker:  aws.String(marker),
	//		MaxKeys: aws.Long(100),
	//		//Prefix:       aws.String("Prefix"),
	//	}
	//
	//	resp, err := conn.s.ListObjects(params)
	//
	//	if err != nil {
	//		// A non-service error occurred.
	//		panic(err)
	//	}
	//
	//	keys := make([]string, 0, 1000)
	//
	//	//max := resp.MaxKeys
	//	for i, v := range resp.Contents {
	//		keys = append(keys, *(v.Key))
	//		logs.Info("key %d %d %d %s", turn, all, i, *(v.Key))
	//		all++
	//	}
	//
	//	if len(keys) == 0 {
	//		break
	//	}
	//	marker = keys[len(keys)-1]
	//
	//	turn++
	//}
}
