package rsa_test

import (
	"crypto"

	"fmt"
	"testing"

	"encoding/base64"
	"vcs.taiyouxi.net/examples/samsung/rsa_test/rsa"
)

var cipher rsa.Cipher

func init() {
	// demo
	//	client, err := rsa.NewDefault(`-----BEGIN RSA PRIVATE KEY-----
	//MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAKz0WssMzD9pwfHlEPy8+NFSnsX+CeZoogRyrzAdBkILTVCukOfJeaqS07GSpVgtSk9PcFk3LqY59znddga6Kf6HA6Tpr19T3Os1U3zNeU79X/nT6haw9T4nwRDptWQdSBZmWDkY9wvA28oB3tYSULxlN/S1CEXMjmtpqNw4asHBAgMBAAECgYBzNFj+A8pROxrrC9Ai6aU7mTMVY0Ao7+1r1RCIlezDNVAMvBrdqkCWtDK6h5oHgDONXLbTVoSGSPo62x9xH7Q0NDOn8/bhuK90pVxKzCCI5v6haAg44uqbpt7fZXTNEsnveXlSeAviEKOwLkvyLeFxwTZe3NQJH8K4OqQ1KzxK+QJBANmXzpVdDZp0nAOR34BQWXHHG5aPIP3//lnYCELJUXNB2/JYTN57dv5LlE5/Ckg0Bgak764A/CX62bKhe/b+FMsCQQDLe4F2qHGy7Sa81xatm66mEkG3u88g9qRARdEvgx9SW+F1xBt2k/bU2YI31hB8IYXzL8KW9NzDfQPihBBUFn4jAkEAzbrmq/pLPlo6mHV3qE5QA2+J+hRh0UYVKsVDKkJGLH98gepS45hArbawBne/NP1bJTUVGKP9w7sl0es01hbteQJATzLO/QQq3N15Cl8dMI07uN+6PG0Y/VeCLpH+DWQXuNKSOmgN2GVW2RmfmWP0Hpxdqn2YW3EKy/vIm02TnWbzyQJAXwujUR9u9s8BZI33kw3gQ7bvWVYt8yyiYzWD2Qrnyg08tN5o+JsjW3fEDWHm70jjZIc+l/5FaZ7H5NOYpnVcpA==
	//-----END RSA PRIVATE KEY-----
	//`, `-----BEGIN PUBLIC KEY-----
	//MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCs9FrLDMw/acHx5RD8vPjRUp7F/gnmaKIEcq8wHQZCC01QrpDnyXmqktOxkqVYLUpPT3BZNy6mOfc53XYGuin+hwOk6a9fU9zrNVN8zXlO/V/50+oWsPU+J8EQ6bVkHUgWZlg5GPcLwNvKAd7WElC8ZTf0tQhFzI5raajcOGrBwQIDAQAB
	//-----END PUBLIC KEY-----`)

	// demo2
	//	client, err := rsa.NewDefault(`-----BEGIN RSA PRIVATE KEY-----
	//MIICWwIBAAKBgQCui7zzy+Ce1fhHMaWfYMJ4/DOvTKBLEPq1mUw9o9SdN44gcZAuoP8RgJ6TtJnmjF7vDtImydd5tgC30at5X43AZsUxSqGTX4gGQE5OSzjSJt1zmlb/k5s64hMrYMmCxJ2bOW0cQTbfUfxDySH8MiyCu67KijCAHHiWzofTEQECewIDAQABAoGAKEAvNaVZSieblodbYzKEBjRakt0/xa/HsOMGEtzZ5dtu2gp2LlqQF3AqoXMvXlwWdFhdm/ZFy1puNfWS7m1bmZZKwLHryQfchhSgWZtrsE9BJyGkd12RPc5/4ljkK6xv9AhkNf3mDdhEEIsX/8seE81eAjLttlfyaFzLQtN9KOkCQQDbL9vZjp2RnDY3xm02WyvmHyKBFK12Q0ixt8Gt5tW+EZPkAU9MKJ9AHhk+80jTtISUwZWpZaydnfNWyfDZ0FlXAkEAy9x+tfUDQo3aR2a5vGewvTiCNV4myLBT5dSY9mmnVzeUWHL1l5ojKpVdhg6aYcYmpVQ8vSdtaNq8JQA03ZDVfQJAW5M6QkIAeRaP3Guts1kSFToK2209L7zawU1pwPNBeAC2Djux2rraFhq9J3zTf4fbIJ9knPqazNtyEF+cnhQbTwJAEJn3M6gtSMk2gmQKMh6blP06FVCChgtd+bRzdHWsK/0Zto4+E8d4n6okQJuF1PqHASW4AItqbISLl9PJelWmHQJAGGcGSe2gopJvARs9g4Ki+6fWKepBqcHy3hSFfiuqcWw+t5+TgQknER1XgViQQ7eGJaZPkuJEIX1MNTZmTOhmzQ==
	//-----END RSA PRIVATE KEY-----
	//`, `-----BEGIN PUBLIC KEY-----
	//"MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCFPPKaX8CpITuPmYegzRXZi7135YMj6RpKOdXsS/I5795I6jBZiq77v8YZ2iKwVPjdvVOEsgV6ZeZXatjliDw0DupO50pIJhZccjAd7lxeec5uJIQTxDSrnRYKQoqQVufGvAjFcteBryjMF8E3Z5X8Q0q5HH3kbRk/sxd3XmZvDQIDAQAB
	//-----END PUBLIC KEY-----`)

	// game
	// pkcs8
	// MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAJJvnlNYJQHZjEuBxnf3onDd3qwLW+o9lyAtTdmsoRa9h4zpr4ktwYFzt6ATCVWaeH5SHaobAtl9j9oFzdo9NC1D7saTidWqFi6GrE0mOycnugduFdbzG/R9yf+HE4YQCo1gbPEBEvX3fBNLmWexB71ZTXnJ+AGsRpG8lHGf93OvAgMBAAECgYAeeZpCkBqLiwHWiPiL6P5mbNY1idukIRM9gBAul+2idOkr20pLyzj1PubdKnkL1qzy5RYFW7+5EZeJcSHJJGmiG9Ns7PK9eHf4CgwDNwVDjE1gYntbITKQzIJRInccXQMt8gHnKkHU2mMTLlD3j5cphFa1oXkJyHf8Svqm22o+kQJBAOvLCTv3TGFlSApXLZ4zZSFrWNc8RsZmFmRPo7kX1xg54Ynw6/mbjsB4cwG0R34jp3sXRmC39KiYN+oolb0111cCQQCe/Dcc55LtexP7rxmn5Gwy7KKKsSP5xnD4l11PbuUnYtm563iNvrUMe73X9curGOuNPCbBxjsMBOcG8KOOf0dpAkEAvRmc80GWOPnv95MI27oeHdooap1FqXFP/ey8wgFzmFM7JNRZe1oc9xDiMviw6WGQvK/i2khNZzSEiz83L4JRiwJAZH4GFCYoUCX45qCyddZDshizUTlRBSOy6t86Yug0vqWT/Bnj9Kyz4fUhda1vFHJHCdoYoWk8j4dVxKQ8reiz6QJAXgCEmMfAqS2KYjnAwxrD/VaKIljx465Ztc9n7HVefR+uy+BV+Kg1m4bO7hdS6DvMRsvOnb0+ea9n2MCRpNsfNA==
	// pkcs1
	// MIICXAIBAAKBgQCSb55TWCUB2YxLgcZ396Jw3d6sC1vqPZcgLU3ZrKEWvYeM6a+JLcGBc7egEwlVmnh+Uh2qGwLZfY/aBc3aPTQtQ+7Gk4nVqhYuhqxNJjsnJ7oHbhXW8xv0fcn/hxOGEAqNYGzxARL193wTS5lnsQe9WU15yfgBrEaRvJRxn/dzrwIDAQABAoGAHnmaQpAai4sB1oj4i+j+ZmzWNYnbpCETPYAQLpftonTpK9tKS8s49T7m3Sp5C9as8uUWBVu/uRGXiXEhySRpohvTbOzyvXh3+AoMAzcFQ4xNYGJ7WyEykMyCUSJ3HF0DLfIB5ypB1NpjEy5Q94+XKYRWtaF5Cch3/Er6pttqPpECQQDrywk790xhZUgKVy2eM2Uha1jXPEbGZhZkT6O5F9cYOeGJ8Ov5m47AeHMBtEd+I6d7F0Zgt/SomDfqKJW9NddXAkEAnvw3HOeS7XsT+68Zp+RsMuyiirEj+cZw+JddT27lJ2LZuet4jb61DHu91/XLqxjrjTwmwcY7DATnBvCjjn9HaQJBAL0ZnPNBljj57/eTCNu6Hh3aKGqdRalxT/3svMIBc5hTOyTUWXtaHPcQ4jL4sOlhkLyv4tpITWc0hIs/Ny+CUYsCQGR+BhQmKFAl+OagsnXWQ7IYs1E5UQUjsurfOmLoNL6lk/wZ4/Sss+H1IXWtbxRyRwnaGKFpPI+HVcSkPK3os+kCQF4AhJjHwKktimI5wMMaw/1WiiJY8eOuWbXPZ+x1Xn0frsvgVfioNZuGzu4XUug7zEbLzp29PnmvZ9jAkaTbHzQ=
	client, err := rsa.NewDefault(`-----BEGIN RSA PRIVATE KEY-----
MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAKoV4XnueeNAkvzsgyCAzddxGZXybbQa8ZIwAzQmDurTNzbSctHU6kkJuapsgdyD2Aq39k8nieFUYonYot3v0pqy5OzVeZDZ0mu/GiwCx9nMkKJtjlgdSaRji8ffZeUvZ/JGY+3C1pTB4v15+0ky5SqXkHoGqFgNyTXCJayAnPZDAgMBAAECgYAulGd3mRPQZLLciXkvwZad1d+H7SiWFnrp6jQ2Z+XV8ZpBbUj8pi6zafJq9eRqm8Dizpap/s4H47BIyAdyeGdYfCsEar0VaACSJcTZS5hnnzVdb5gX6pdbvun/Fi1HNZFW+XCByyNPGQuLjczg4dvFY/2bNYt6LpPLHrJsChuogQJBANV+GBdwGIqnnjOQGH+UWbxyzm3QrILxBqGltgS/8UB0obHYEZMIL6UUjDfW40J7xPXhYsNqwU7KO+Q8xs/XEfECQQDL80n2cRPSegUU/v4cbFY/g5pN01uTSCbEnMjEDL5u+Jzvf1G48peZmiw0RZ9LzWxA0jC4gTOOs0xx8TAjV1dzAkBOfq4c7/oWAMsJ6lEXl1PnFc8QUUkcW8I0bNkfpfLt3/QTj33msXvTFlr3rOqh5x/jx5qofvfUIEclA7OVd14BAkAUjGeYT95KZ4bZjbN2k6fA8HZ8ft4MIcneJ1nG/u206pGNQ8utEawaisEHZzhcf873XPYRsNrL9t6t4DoUZXlnAkEAgZa3aD/rvNPNVp64+KEFBh3XyTodrE5jmLgNBQUsV7VQNT9/qkPJrAniMys4gCeWZZCc3ayVnc/qmHO3mS+UCg==
-----END RSA PRIVATE KEY-----`, `-----BEGIN RSA PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCn3RIAam9MMoEMtFz82DfByfMoWIcMqDtUILkzWdyXK7/oAL2bzc1BdU9BnEuI2ZoiMVWYkDoGksVxaODfbAiTjACiqsubiRT+s0TouLYiUMA3I/ncS6UEwqR0hrWIAJe0Bv8VmHcgvhu/SbMv2Uloc/eEKl4sxzMb+fMfc/80UQIDAQAB
-----END RSA PUBLIC KEY-----`)

	if err != nil {
		fmt.Println(err)
	}

	cipher = client
}

//func Test_DefaultClient(t *testing.T) {
//
//	cp, err := cipher.Encrypt([]byte("测试加密解密"))
//	if err != nil {
//		t.Error(err)
//	}
//	cpStr := base64.URLEncoding.EncodeToString(cp)
//
//	fmt.Println(cpStr)
//
//	ppBy, err := base64.URLEncoding.DecodeString(cpStr)
//	if err != nil {
//		t.Error(err)
//	}
//	pp, err := cipher.Decrypt(ppBy)
//
//	fmt.Println(string(pp))
//}

func Test_Sign_DefaultClient(t *testing.T) {

	src := `{"appid":"5000204106","appuserid":"1:17:a73e88d9-ec6a-40f6-85b5-77c","cporderid":"1:17:a73e88d9-ec6a-40f6-85b5-77c3f72b3275:17:101","cpprivate":"1478487431:1:1.11.0","currency":"RMB","feetype":0,"money":1.00,"paytype":403,"result":0,"transid":"32041611071057135261","transtime":"2016-11-07 10:57:26","transtype":0,"waresid":12}`
	//src := "a"

	//signBytes, err := cipher.Sign([]byte(src), crypto.MD5)
	//if err != nil {
	//	t.Error(err)
	//}
	//sign := base64.StdEncoding.EncodeToString(signBytes)
	sign := "PCCeSBMHaiMaToOkWS/MFUqio9r80Ix1pHFW382ntQQbZ50M+HjXAGfByegcY6BhdEdStb3Yl6CkWa6Fe5wjbGCvfv2faPY/bMG05mDFd75pvU4c6FkX571rIPKEhlRlAt/MlquHMPqFFwKkspL3LmM8r7C+Ld/HUJ+9x/Bdfqs="
	fmt.Println(sign)

	signB, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		t.Error(err)
		return
	}
	errV := cipher.Verify([]byte(src), signB, crypto.MD5)
	if errV != nil {
		t.Error(errV)
		return
	}
	fmt.Println("verify success")
}
