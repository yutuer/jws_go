package atc_mng

import (
	"testing"
)

func TestProfile(t *testing.T) {
	httpGet("http://10.0.1.222/api/v1/profiles/")
}

func TestToken(t *testing.T) {
	httpGet("http://10.0.1.222/api/v1/token/")
}
