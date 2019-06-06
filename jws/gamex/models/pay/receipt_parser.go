package pay

import (
	"encoding/base64"
	"regexp"
)

/*
{
	"original-purchase-date-pst" = "2016-01-04 00:08:15 America/Los_Angeles";
	"unique-identifier" = "40810849679aea9cbecd8c43dffe0db0a664c8e2";
	"original-transaction-id" = "1000000187273883";
	"bvrs" = "722";
	"transaction-id" = "1000000187273883";
	"quantity" = "1";
	"original-purchase-date-ms" = "1451894895779";
	"unique-vendor-identifier" = "D218920E-DDE0-4B9B-B3FA-61881C6B0ABE";
	"product-id" = "com.taiyouxi.ifsg.10";
	"item-id" = "1061946652";
	"bid" = "com.taiyouxi.ifsg";
	"purchase-date-ms" = "1451894895779";
	"purchase-date" = "2016-01-04 08:08:15 Etc/GMT";
	"purchase-date-pst" = "2016-01-04 00:08:15 America/Los_Angeles";
	"original-purchase-date" = "2016-01-04 08:08:15 Etc/GMT";
}
*/

type (
	ReceiptData struct {
		Signature    string           `json:"signature"`
		PurchaseInfo PurchaseInfoData `json:"purchase-info"`
		Environment  string           `json:"environment"`
		Pod          string           `json:"pod"`
		Status       string           `json:"signing-status"`
	}

	PurchaseInfoData struct {
		OriginalPurchaseDatePst string `json:"original-purchase-date-pst"`
		UniqueIdentifier        string `json:"unique-identifier"`
		OriginalTransactionId   string `json:"original-transaction-id"`
		Bvrs                    string `json:"bvrs"`
		TransactionId           string `json:"transaction-id"`
		Quantity                string `json:"quantity"`
		OriginalPurchaseDateMs  string `json:"original-purchase-date-ms"`
		UniqueVendorIdentifier  string `json:"unique-vendor-identifier"`
		ProductID               string `json:"product-id"`
		ItemID                  string `json:"item-id"`
		BID                     string `json:"bid"`
		PurchaseDateMs          string `json:"purchase-date-ms"`
		PurchaseDate            string `json:"purchase-date"`
		PurchaseDatePst         string `json:"purchase-date-pst"`
		OriginalPurchaseDate    string `json:"original-purchase-date"`
	}
)

func ParseReceiptData(data string) (*ReceiptData, error) {
	res := ReceiptData{}
	str, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	receipts := parseReceiptDataToMap(string(str))

	if s, ok := receipts["environment"]; ok {
		res.Environment = s
	}

	if s, ok := receipts["pod"]; ok {
		res.Pod = s
	}

	if s, ok := receipts["signing-status"]; ok {
		res.Status = s
	}

	if s, ok := receipts["signature"]; ok {
		res.Signature = s
	}

	if s, ok := receipts["purchase-info"]; ok {
		str, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return &res, err
		}
		purchases := parseReceiptDataToMap(string(str))
		if s, ok := purchases["original-purchase-date-pst"]; ok {
			res.PurchaseInfo.OriginalPurchaseDatePst = s
		}
		if s, ok := purchases["unique-identifier"]; ok {
			res.PurchaseInfo.UniqueIdentifier = s
		}
		if s, ok := purchases["original-transaction-id"]; ok {
			res.PurchaseInfo.OriginalTransactionId = s
		}
		if s, ok := purchases["bvrs"]; ok {
			res.PurchaseInfo.Bvrs = s
		}
		if s, ok := purchases["transaction-id"]; ok {
			res.PurchaseInfo.TransactionId = s
		}
		if s, ok := purchases["quantity"]; ok {
			res.PurchaseInfo.Quantity = s
		}
		if s, ok := purchases["original-purchase-date-ms"]; ok {
			res.PurchaseInfo.OriginalPurchaseDateMs = s
		}
		if s, ok := purchases["unique-vendor-identifier"]; ok {
			res.PurchaseInfo.UniqueVendorIdentifier = s
		}
		if s, ok := purchases["product-id"]; ok {
			res.PurchaseInfo.ProductID = s
		}
		if s, ok := purchases["item-id"]; ok {
			res.PurchaseInfo.ItemID = s
		}
		if s, ok := purchases["bid"]; ok {
			res.PurchaseInfo.BID = s
		}
		if s, ok := purchases["purchase-date-ms"]; ok {
			res.PurchaseInfo.PurchaseDateMs = s
		}
		if s, ok := purchases["purchase-date"]; ok {
			res.PurchaseInfo.PurchaseDate = s
		}
		if s, ok := purchases["purchase-date-pst"]; ok {
			res.PurchaseInfo.PurchaseDatePst = s
		}
		if s, ok := purchases["original-purchase-date"]; ok {
			res.PurchaseInfo.OriginalPurchaseDate = s
		}

	}

	return &res, nil

}

var regReceiptData = regexp.MustCompile(`.*\"(.*)\".*=.*\"(.*)\".*;`)

func parseReceiptDataToMap(data string) map[string]string {
	res := make(map[string]string, 16)
	ss := regReceiptData.FindAllStringSubmatch(data, -1)
	for _, s := range ss {
		if len(s) >= 3 {
			res[s[1]] = s[2]
		}
	}
	return res
}
