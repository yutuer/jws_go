package secure

import (
	"encoding/base64"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEncoding(t *testing.T) {
	Convey("测试加密模块和客户端表现一致", t, func() {
		// TBD 需要更多测测试用例来测试完整性
		username := "username"
		stdb64 := base64.StdEncoding.EncodeToString([]byte(username))
		urlb64 := base64.URLEncoding.EncodeToString([]byte(username))
		planxb64 := DefaultEncode.Encode64ForNet([]byte(username))
		So(stdb64, ShouldEqual, "dXNlcm5hbWU=")
		So(urlb64, ShouldEqual, "dXNlcm5hbWU=")
		So(planxb64, ShouldEqual, "3qgYF8sxvw6=")

		std, _ := base64.StdEncoding.DecodeString("dXNlcm5hbWU=")
		url, _ := base64.URLEncoding.DecodeString("dXNlcm5hbWU=")
		planxs, _ := DefaultEncode.Decode64FromNet("3qgYF8sxvw6=")
		So(string(std), ShouldEqual, username)
		So(string(url), ShouldEqual, username)
		So(string(planxs), ShouldEqual, username)

	})
}
