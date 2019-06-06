package timail

type PackageInfo struct {
	PkgId    int
	SubPkgId int
}

// tag中可能信息
type IAPOrder struct {
	Order_no      string      `json:"on"`
	Game_order    string      `json:"go"`
	Game_order_id string      `json:"goid"`
	Amount        string      `json:"a"`
	Channel       string      `json:"ch"`
	PayTime       string      `json:"t"`       // 客户端带过来的支付时间
	PkgInfo       PackageInfo `json:"pkginfo"` //直购礼包信息
	PayType       string      `json:"payty"`
}

type VerTag struct {
	Ver string `json:"ver"`
	Ch  string `json:"ch"`
}
