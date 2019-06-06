package main

import (
	"encoding/csv"
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gocarina/gocsv"
	"gopkg.in/mgo.v2"
)

type PayItem struct {
	AccountName      string `bson:"account_name,omitempty"`
	Channel          string `bson:"channel,omitempty"`
	ChannelUID       string `bson:"channel_uid,omitempty"`
	ExtraParams      string `bson:"extras_params,omitempty"`
	GoodIndex        string `bson:"good_idx"`
	GoodName         string `bson:"good_name"`
	HcBuy            int64  `bson:"hc_buy"`
	HcGive           int64  `bson:"hc_give"`
	IsTest           string `bson:"is_test,omitempty"`
	Mobile           string `bson:"mobile"`
	MoneyAmount      string `bson:"money_amount"`
	OrderNo          string `bson:"order_no"`
	PayTime          string `bson:"pay_time"`
	PayTimestamp     string `bson:"pay_time_s"`
	Platform         string `bson:"platform"`
	ProductID        string `bson:"product_id"`
	ReceiveTimestamp int64  `bson:"receiveTimestamp"`
	RoleName         string `bson:"role_name"`
	SN               int64  `bson:"sn"`
	Status           string `bson:"status,omitempty"`
	TiStatus         string `bson:"tistatus,omitempty"`
	Uid              string `bson:"uid,omitempty"`
	Version          string `bson:"ver,omitempty"`
}

func main() {
	flag.Parse()
	l := log.New(os.Stderr, "", 0)
	session, err := mgo.Dial(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer session.Close()

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

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Eventual, true)

	c := session.DB("").C("Pay")
	cursor := c.Find(nil).Iter()

	var total int64
	var m PayItem
	for cursor.Next(&m) {
		n := m
		payChan <- n
		total++
	}

	if cursor.Err() != nil {
		cursor.Close()
	}

	l.Println("all:", total)
	close(payChan)
	wg.Wait()
	time.Sleep(time.Second * 1)
}
