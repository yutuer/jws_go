package main

import (
	"io"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/examples/proto/pbgen"
)

func main() {
	//file, err := os.Open("tnt_deploy_goods_info.data")
	file, err := os.Open("item.data")
	if err != nil {
		log.Fatalf("failed: %s\n", err)
		return
	}

	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		log.Fatalf("failed: %s\n", err)
		return
	}

	buffer := make([]byte, fi.Size())
	_, err = io.ReadFull(file, buffer) //read all content
	if err != nil {
		log.Fatalf("failed: %s\n", err)
		return
	}

	//Goods := &tnt_deploy.GOODS_INFO_ARRAY{}
	//err = proto.Unmarshal(buffer, Goods)
	//if err != nil {
	//log.Fatal("unmarshaling error: ", err)
	//}
	//// Now test and newTest contain the same data.
	//for i, item := range Goods.Items {
	//log.Printf("%d, %v", i, string(item.GetName()))
	//}

	Items := &ProtobufGen.Item_ARRAY{}
	err = proto.Unmarshal(buffer, Items)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}
	for i, item := range Items.Items {
		log.Printf("%d, %v", i, item.GetNameIDS())
	}

}
