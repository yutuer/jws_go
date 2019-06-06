package incode

import (
	"testing"
)

func TestGet(t *testing.T) {
	ok, err := CheckIsCodeVerify("a8e47ca5dd12da1b846da4260109b82d", "18C3F552")
	if err != nil {
		t.Errorf("CheckIsCodeVerify Err by %s", err.Error())
	}
	t.Logf("Code : %v", ok)
}

func TestGetErr(t *testing.T) {
	ok, err := CheckIsCodeVerify("a8e47ca5dd12da1b846da4260109b82d", "dsfdsf")
	if err != nil {
		t.Errorf("CheckIsCodeVerify Err by %s", err.Error())
	}
	t.Logf("Code : %v", ok)
}

func TestUSE(t *testing.T) {
	err := UseCode("a8e47ca5dd12da1b846da4260109b82d", "77A044AA")
	if err != nil {
		t.Errorf("UseCode Err by %s", err.Error())
	}
	t.Logf("Code : ok")
}
