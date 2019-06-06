package main

import (
	"testing"

	"fmt"
	"vcs.taiyouxi.net/platform/x/gainer_beholder/tools"
)

func TestGetHotPackageData(t *testing.T) {
	for i, v := range tools.Hot_package {
		fmt.Printf("hotpackage 第 %v 个 是 %v\n", i, v)
	}
	fmt.Println("over")
}
