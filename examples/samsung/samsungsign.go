package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

const (

	//privateKey = "MIICXAIBAAKBgQCSb55TWCUB2YxLgcZ396Jw3d6sC1vqPZcgLU3ZrKEWvYeM6a+JLcGBc7egEwlVmnh+Uh2qGwLZfY/aBc3aPTQtQ+7Gk4nVqhYuhqxNJjsnJ7oHbhXW8xv0fcn/hxOGEAqNYGzxARL193wTS5lnsQe9WU15yfgBrEaRvJRxn/dzrwIDAQABAoGAHnmaQpAai4sB1oj4i+j+ZmzWNYnbpCETPYAQLpftonTpK9tKS8s49T7m3Sp5C9as8uUWBVu/uRGXiXEhySRpohvTbOzyvXh3+AoMAzcFQ4xNYGJ7WyEykMyCUSJ3HF0DLfIB5ypB1NpjEy5Q94+XKYRWtaF5Cch3/Er6pttqPpECQQDrywk790xhZUgKVy2eM2Uha1jXPEbGZhZkT6O5F9cYOeGJ8Ov5m47AeHMBtEd+I6d7F0Zgt/SomDfqKJW9NddXAkEAnvw3HOeS7XsT+68Zp+RsMuyiirEj+cZw+JddT27lJ2LZuet4jb61DHu91/XLqxjrjTwmwcY7DATnBvCjjn9HaQJBAL0ZnPNBljj57/eTCNu6Hh3aKGqdRalxT/3svMIBc5hTOyTUWXtaHPcQ4jL4sOlhkLyv4tpITWc0hIs/Ny+CUYsCQGR+BhQmKFAl+OagsnXWQ7IYs1E5UQUjsurfOmLoNL6lk/wZ4/Sss+H1IXWtbxRyRwnaGKFpPI+HVcSkPK3os+kCQF4AhJjHwKktimI5wMMaw/1WiiJY8eOuWbXPZ+x1Xn0frsvgVfioNZuGzu4XUug7zEbLzp29PnmvZ9jAkaTbHzQ="
	//publicKey = "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDFEg09QvSNkHVNpbqgA18MZ4RK/80DBvjoojnIvU6Jm2QcDJaAqGQBkqLoslCg8QcZ8TtUjooGkvVn7Y4nZYhpyI6SU13xsjQ3a98IaHfyNFQAUipr53yyuUpLu45qjqZM/WloC3TndeNk9y0M4voCVYLnCj0mooN63z9153oPSQIDAQAB"
	priKey = "MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAKz0WssMzD9pwfHlEPy8+NFSnsX+CeZoogRyrzAdBkILTVCukOfJeaqS07GSpVgtSk9PcFk3LqY59znddga6Kf6HA6Tpr19T3Os1U3zNeU79X/nT6haw9T4nwRDptWQdSBZmWDkY9wvA28oB3tYSULxlN/S1CEXMjmtpqNw4asHBAgMBAAECgYBzNFj+A8pROxrrC9Ai6aU7mTMVY0Ao7+1r1RCIlezDNVAMvBrdqkCWtDK6h5oHgDONXLbTVoSGSPo62x9xH7Q0NDOn8/bhuK90pVxKzCCI5v6haAg44uqbpt7fZXTNEsnveXlSeAviEKOwLkvyLeFxwTZe3NQJH8K4OqQ1KzxK+QJBANmXzpVdDZp0nAOR34BQWXHHG5aPIP3//lnYCELJUXNB2/JYTN57dv5LlE5/Ckg0Bgak764A/CX62bKhe/b+FMsCQQDLe4F2qHGy7Sa81xatm66mEkG3u88g9qRARdEvgx9SW+F1xBt2k/bU2YI31hB8IYXzL8KW9NzDfQPihBBUFn4jAkEAzbrmq/pLPlo6mHV3qE5QA2+J+hRh0UYVKsVDKkJGLH98gepS45hArbawBne/NP1bJTUVGKP9w7sl0es01hbteQJATzLO/QQq3N15Cl8dMI07uN+6PG0Y/VeCLpH+DWQXuNKSOmgN2GVW2RmfmWP0Hpxdqn2YW3EKy/vIm02TnWbzyQJAXwujUR9u9s8BZI33kw3gQ7bvWVYt8yyiYzWD2Qrnyg08tN5o+JsjW3fEDWHm70jjZIc+l/5FaZ7H5NOYpnVcpA=="
	pubKey = "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCs9FrLDMw/acHx5RD8vPjRUp7F/gnmaKIEcq8wHQZCC01QrpDnyXmqktOxkqVYLUpPT3BZNy6mOfc53XYGuin+hwOk6a9fU9zrNVN8zXlO/V/50+oWsPU+J8EQ6bVkHUgWZlg5GPcLwNvKAd7WElC8ZTf0tQhFzI5raajcOGrBwQIDAQAB"
)

func main() {

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: []byte(priKey),
	}

	file, err := os.Create("private.pem")

	if err != nil {
		fmt.Println(err)
		return
	}

	err = pem.Encode(file, block)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func GenRsaKey(bits int) error {

	// 生成私钥文件

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)

	if err != nil {

		return err

	}

	derStream := x509.MarshalPKCS1PrivateKey(privateKey)

	block := &pem.Block{

		Type: "RSA PRIVATE KEY",

		Bytes: derStream,
	}

	file, err := os.Create("private.pem")

	if err != nil {

		return err

	}

	err = pem.Encode(file, block)

	if err != nil {

		return err

	}

	// 生成公钥文件

	publicKey := &privateKey.PublicKey

	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)

	if err != nil {

		return err

	}

	block = &pem.Block{

		Type: "PUBLIC KEY",

		Bytes: derPkix,
	}

	file, err = os.Create("public.pem")

	if err != nil {

		return err

	}

	err = pem.Encode(file, block)

	if err != nil {

		return err

	}

	return nil

}
