package util

import (
	"fmt"

	"crypto"
	"encoding/base64"

	"vcs.taiyouxi.net/platform/x/api_gateway/util/rsa"
)

const SamsungChannel = "5000"

var cipher rsa.Cipher

func Init(privKey, publicKey string) {
	client, err := rsa.NewDefault(fmt.Sprintf(`-----BEGIN RSA PRIVATE KEY-----
%s
-----END RSA PRIVATE KEY-----`, privKey), fmt.Sprintf(`-----BEGIN PUBLIC KEY-----
%s
-----END PUBLIC KEY-----`, publicKey))
	if err != nil {
		panic(fmt.Errorf("init rsa.Cipher err %v", err))
	}
	cipher = client
}

func Sign(content []byte) (string, error) {
	signBytes, err := cipher.Sign(content, crypto.MD5)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signBytes), nil
}

func Verify(content, sign []byte) error {
	err := cipher.Verify(content, sign, crypto.MD5)
	if err != nil {
		return err
	}
	return nil
}
