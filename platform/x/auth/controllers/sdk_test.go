package controllers

import (
	"testing"
	"vcs.taiyouxi.net/platform/x/auth/models/sdk"
	"fmt"
)

func TestCheck6waves(t *testing.T){
	err := sdk.Check6waves("abcde")
	fmt.Println(err)
}