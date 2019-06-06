package RW_xlsx

import (
	"fmt"
	"testing"
)

func TestgetFieldDataFromXlsx(t *testing.T) {
	xlsx_data, err := GetFieldDataFromXlsx("KoSuccessOrders.xlsx", "orders", "pay_description")
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range xlsx_data {
		fmt.Println(v)
	}
}

func TestgetDataFromXlsx(t *testing.T) {
	xlsx_data, err := GetDataFromXlsx("KoSuccessOrders.xlsx", "orders")
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range xlsx_data {
		fmt.Println(v)
	}
}

func TestgetIndexFromField(t *testing.T) {
	xlsx_data, err := GetDataFromXlsx("KoSuccessOrders.xlsx", "orders")
	if err != nil {
		fmt.Println(err)
	}
	index, err := GetIndexFromField(xlsx_data, "pay_description")
	fmt.Println(index)
}

func TestwriteToXlsxFile(t *testing.T) {
	xlsx_data, err := GetDataFromXlsx("KoSuccessOrders.xlsx", "orders")
	if err != nil {
		fmt.Println(err)
	}
	WriteToXlsxFile(xlsx_data, "CopyOfKoSuccessOrders.xlsx", "orders")
}
