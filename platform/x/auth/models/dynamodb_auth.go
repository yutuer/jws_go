package models

import (
	"errors"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	DDB "github.com/aws/aws-sdk-go/service/dynamodb"
	. "vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	HadRole = "had"
)

type AuthDynamoDB struct {
	*DynamoDB
}

type AuthUserShardInfo struct {
	GidSid  string `json:"gidsid" bson:"gid_sid,omitempty"`
	HasRole string `json:"role" bson:"has_role,omitempty"`
}

func (d *AuthDynamoDB) QueryUserShardInfo(table_name, uid string) ([]AuthUserShardInfo, error) {
	hash_typ, _, ok := d.GetHashType(table_name)
	if !ok {
		return nil, errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	query_keys := make(map[string]*DDB.Condition)
	query_keys[hash_typ] = &DDB.Condition{
		AttributeValueList: []*DDB.AttributeValue{CreateAttributeValue(uid)},
		ComparisonOperator: aws.String("EQ"),
	}

	in := &DDB.QueryInput{
		TableName:     aws.String(table_name),
		KeyConditions: query_keys,
	}

	query_out, err := d.Client().Query(in)
	if !ok {
		return nil, err
	}

	re := make([]AuthUserShardInfo, 0, len(query_out.Items))

	for _, i := range query_out.Items {
		uid, uok := i["Uid"]
		gid_sid, sok := i["gid_sid"]
		hasRole, hok := i["has_role"]
		if !uok || !sok || !hok {
			logs.Error("QueryUserShardInfo info error by %s!", uid)
			continue
		}

		rec := AuthUserShardInfo{}
		gs, sok := GetItemValue(gid_sid).(string)
		hr, hok := GetItemValue(hasRole).(string)
		if !sok || !hok {
			logs.Error("QueryUserShardInfo info typ err by %s for %v %v!", uid, gid_sid, hasRole)
			continue
		}
		rec.GidSid = gs
		rec.HasRole = hr
		re = append(re, rec)
	}
	return re, nil
}

func (d *AuthDynamoDB) SetUserShardInfo(table_name, uid, gidSidStr string, hasRole string) error {
	hash_typ, range_typ, ok := d.GetHashType(table_name)
	if !ok {
		return errors.New(fmt.Sprintf("No table Info, %s", table_name))
	}

	updates := make(map[string]*DDB.AttributeValueUpdate, 1)
	v_a := CreateAttributeValue(hasRole)
	updates["has_role"] = &DDB.AttributeValueUpdate{
		Action: aws.String("PUT"),
		Value:  v_a,
	}

	update_item := &DDB.UpdateItemInput{
		AttributeUpdates: updates,
		TableName:        aws.String(table_name),
		Key: map[string]*DDB.AttributeValue{
			hash_typ:  CreateAttributeValue(uid),
			range_typ: CreateAttributeValue(gidSidStr),
		},
		ReturnConsumedCapacity: aws.String("TOTAL"),
		Expected:               map[string]*DDB.ExpectedAttributeValue{},
		ReturnValues:           aws.String("UPDATED_NEW"),
	}
	_, err := d.Client().UpdateItem(update_item)
	return err
}
