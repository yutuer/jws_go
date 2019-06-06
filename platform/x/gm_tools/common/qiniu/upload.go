package qiniu

import (
	"bytes"
	"github.com/qiniu/api.v6/io"
	"github.com/qiniu/api.v6/rs"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func uptoken(bucketName string) string {
	putPolicy := rs.PutPolicy{
		Scope: bucketName,
		//CallbackUrl: callbackUrl,
		//CallbackBody:callbackBody,
		//ReturnUrl:   returnUrl,
		//ReturnBody:  returnBody,
		//AsyncOps:    asyncOps,
		//EndUser:     endUser,
		//Expires:     expires,
	}
	return putPolicy.Token(nil)
}

func UpLoad(bucketName, key string, data []byte) error {
	var err error
	var ret io.PutRet
	var extra = &io.PutExtra{
	//Params:    params,
	//MimeType:  mieType,
	//Crc32:     crc32,
	//CheckCrc:  CheckCrc,
	}

	rsClient := rs.New(nil)
	rsClient.Delete(nil, bucketName, key)
	// ret       变量用于存取返回的信息，详情见 io.PutRet
	// uptoken   为业务服务器端生成的上传口令
	// key       为文件存储的标识
	// r         为io.Reader类型，用于从其读取数据
	// extra     为上传文件的额外信息,可为空， 详情见 io.PutExtra, 可选
	err = io.Put(nil, &ret, uptoken(bucketName), key, bytes.NewReader(data), extra)

	if err != nil {
		//上传产生错误
		logs.Error("io.Put failed: %s", err.Error())
	}

	return err
}
