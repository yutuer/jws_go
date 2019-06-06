package util

import (
	"crypto/md5"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

//TODO:
// - [] self signed ssl http://stackoverflow.com/questions/12122159/golang-how-to-do-a-https-request-with-bad-certificate
// - [] 因为混服,需要让客户端通知服务器,这个是vivo帐号
// - [] 支付流程不太一样! ,需要找gamex传输一些参数,然后返回交易代码.quick流程用这个机会返回cpOrderNumber
// 		1.notify url (最长200)由服务器配置,
// 		1. 同时返回cpid, appid
// 		1. cpOrderNumber 最长64位.字母、数字和下划线组成
// 		1. extInfo 最长64位. extInfo不能为空
// - [] 支付trade需要注意: 判断version: 1.0.0, signMethod: MD5
// - [] 具体流程需要参考相关文档的处理逻辑伪代码和错误类型描述
// - [] 11 次回掉未能达成后,考虑使用主动query来处理? 需要单独加特殊功能来实现。
// - [] 这11次的间隔是多少?

const VivoChannel = "17"
const SIGNATURE = "signature"
const SIGN_METHOD = "signMethod"

type Para map[string]string

/**
 * 拼接请求字符串
 * @param req 	请求要素
 * @param key	vivo分配给商户的密钥
 * @return 请求字符串
 */
func buildReq(req Para, key string) string {
	// 除去数组中的空值和签名参数
	filteredReq := paraFilter(req)
	// 根据参数获取vivo签名
	signature := GetVivoSign(filteredReq, key)
	// 签名结果与签名方式加入请求提交参数组中
	filteredReq[SIGNATURE] = signature
	filteredReq[SIGN_METHOD] = "MD5"

	//value需要URL编码
	//NOTE: java vivo example:value = URLEncoder.encode(value, "utf-8");
	kvs := make([]string, 0, len(filteredReq)+1)
	for k, v := range filteredReq {
		kvs = append(kvs, k+"="+url.QueryEscape(v))
	}
	prestr := strings.Join(kvs, "&")
	return prestr
}

/**
 * 除去请求要素中的空值和签名参数
 * @param para 请求要素
 * @return 去掉空值与签名参数后的请求要素
 */
func paraFilter(para Para) Para {
	result := make(Para)

	if para == nil || len(para) <= 0 {
		return result
	}

	for k, v := range para {
		if v == "" ||
			strings.EqualFold(k, SIGNATURE) ||
			strings.EqualFold(k, SIGN_METHOD) {
			continue
		}
		result[k] = v
	}

	return result
}

/**
 * 异步通知消息验签
 * @param para	异步通知消息
 * @param key	vivo分配给商户的密钥
 * @return 验签结果
 */
func VerifySignature(para Para, key string) bool {
	// 除去数组中的空值和签名参数
	filteredReq := paraFilter(para)
	// 根据参数获取vivo签名
	signature := GetVivoSign(filteredReq, key)
	// 获取参数中的签名值
	respSignature := para[SIGNATURE]
	fmt.Println("服务器签名：" + signature + " | 请求消息中的签名：" + respSignature)
	// 对比签名值
	if "" != respSignature && respSignature == signature {
		return true
	} else {
		return false
	}
}

/**
 * 获取vivo签名
 * @param para	参与签名的要素<key,value>
 * @param key	vivo分配给商户的密钥
 * @return 签名结果
 */
func GetVivoSign(para Para, key string) string {
	// 除去数组中的空值和签名参数
	filteredReq := paraFilter(para)

	//得到待签名字符串 需要对map进行sort
	keys := make([]string, 0, len(filteredReq))
	for k, _ := range filteredReq {
		keys = append(keys, k)
	}
	sort.StringSlice(keys).Sort()
	kvs := make([]string, 0, len(filteredReq)+1)
	for _, k := range keys {
		kvs = append(kvs, k+"="+filteredReq[k])
	}
	kvs = append(kvs, md5Summary(key))
	prestr := strings.Join(kvs, "&")
	//fmt.Println(prestr)
	return md5Summary(prestr)
}

/**
 * 对传入的参数进行MD5摘要
 * @param str	需进行MD5摘要的数据
 * @return		MD5摘要值
 */
func md5Summary(str string) string {
	//NOTE: java vivo example make it str convert to UTF8
	//but go is utf8 inside already. we do not make convert it here.
	//TODO: what will happen if http params with Chinese characters?
	bb := md5.Sum([]byte(str))
	return fmt.Sprintf("%x", bb)
}

func main() {
	para := make(Para)
	para["version"] = "1.0.0"
	para["signMethod"] = "MD5"
	para["signature"] = "XXXXXXX"

	para["cpId"] = "123456"
	para["appId"] = "123"
	para["cpOrderNumber"] = "123456878"
	para["notifyUrl"] = "http://127.0.0.1:8080/vivopay/vivoPay/testNotify"
	para["orderTime"] = "20150408162237"
	para["orderAmount"] = "1000"
	para["orderTitle"] = "手机"
	para["orderDesc"] = "vivo Xplay"
	para["extInfo"] = "Payment3.0.1"

	key := "123456789123456"

	sign := GetVivoSign(para, key)
	fmt.Println("vivo签名值：", sign)

	verify := VerifySignature(para, key)
	fmt.Println("验签成功", verify)

	fmt.Println(md5Summary("abcdefg"))

}
