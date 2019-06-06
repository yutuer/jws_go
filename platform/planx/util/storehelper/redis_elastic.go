package storehelper

import (
	"fmt"

	"encoding/json"
	"strings"

	// "sync"

	"time"

	"github.com/bitly/go-simplejson"
	"gopkg.in/olivere/elastic.v2"
	"vcs.taiyouxi.net/platform/planx/util/account_json"
)

const (
	indx_template = "onland-elastics_shard%s.%s"
)

func NewElasticScanRedis(elasticUrl string, shardId string) *ElasticScanRedis {
	return &ElasticScanRedis{
		elasticUrl: elasticUrl,
		shardId:    shardId,
	}
}

type ElasticScanRedis struct {
	shardId    string
	elasticUrl string

	elasticClient *elastic.Client
	// elasticIndexMutex sync.RWMutex
}

func (es *ElasticScanRedis) Open() error {
	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(es.elasticUrl))
	if err != nil {
		return err
	}
	es.elasticClient = client
	return nil
}

func (es *ElasticScanRedis) Close() error {
	return nil
}

func (es *ElasticScanRedis) Put(key string, val []byte, rh ReadHandler) error {
	s := strings.SplitN(key, ":", 2)
	if len(s) <= 0 {
		return fmt.Errorf("ElasticScanRedis key %s unvalid", key)
	}
	store_head := s[0]
	acid := s[1]
	index := fmt.Sprintf(indx_template, es.shardId, store_head)
	typ := store_head
	id := fmt.Sprintf("%s:%s", store_head, acid)

	exist, err := es.elasticClient.IndexExists(index).Do()
	if err != nil {
		return fmt.Errorf("ElasticScanRedis IndexExists %v %s", err, key)
	}
	if !exist {
		_, err = es.elasticClient.CreateIndex(index).Do()
		if err != nil {
			return fmt.Errorf("ElasticScanRedis CreateIndex %v %s", err, key)
		}
	}

	var dat map[string]string
	if err := json.Unmarshal(val, &dat); err != nil {
		return fmt.Errorf("ElasticScanRedis unmarshal val %v %s", err, key)
	}

	j, err := accountJson.MkTrueJsonFromRedis(dat, nil)
	if err != nil {
		return fmt.Errorf("ElasticScanRedis MkTrueJsonFromRedis err %v %s", err, key)
	}
	if strings.ToLower(store_head) == "profile" {
		j.Set("rng", "")
	}

	// 清理2层数组
	jm, err := j.Map()
	if err == nil {
		for k, _ := range jm {
			v := j.Get(k)
			vm, err := v.Map()
			if err == nil {
				for sk, _ := range vm {
					sv := v.Get(sk)
					if _, err := sv.Array(); err == nil {
						v.Set(sk, "")
						j.Set(k, v)
					}
				}
			}
		}
	}

	nt := time.Now()
	mst := nt.Format("2006-01-02T15:04:05.000Z07:00")
	nj := simplejson.New()
	nj.Set("@timestamp", mst)
	nj.Set("accountid", acid)
	nj.Set(store_head, j.Interface())

	jb, err := nj.Encode()
	if err != nil {
		return fmt.Errorf("ElasticScanRedis j.Bytes() err %v %s", err, key)
	}

	str_jb := string(jb)
	_, err = es.elasticClient.Index().
		Index(index).
		Type(typ).
		Id(id).
		BodyString(str_jb).
		Do()
	if err != nil {
		return fmt.Errorf("ElasticScanRedis add index data err:%v key:%s index:%v typ:%v id:%v body:%s", err, key, index, typ, id, str_jb)
	}
	fmt.Println("ElasticScanRedis dump success key " + key)
	return nil
}

func (es *ElasticScanRedis) Get(key string) ([]byte, error) {
	return nil, fmt.Errorf("not support")
}
func (es *ElasticScanRedis) Del(key string) error {
	return fmt.Errorf("not support")
}
func (es *ElasticScanRedis) StoreKey(key string) string {
	return key
}

func (es *ElasticScanRedis) RedisKey(key_in_store string) (string, bool) {
	return "", false
}
func (es *ElasticScanRedis) Clone() (IStore, error) {
	nn := *es
	n := &nn
	if err := n.Open(); err != nil {
		return nil, err
	}
	return n, nil
}
