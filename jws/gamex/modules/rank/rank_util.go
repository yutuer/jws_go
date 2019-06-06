package rank

import (
	"reflect"

	"strconv"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type httpDel interface {
	delRank(acID string) error
}

func (r *RankModule) handleHTTP(c *gin.Context) {
	name := c.PostForm("rank_name")
	param := c.PostForm("param")
	acid := c.PostForm("acid")
	logs.Debug("get rak_name: %v, param: %v, acid: %v", name, param, acid)
	values := reflect.ValueOf(r).Elem()
	field := values.FieldByName(name)
	if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
		index, err := strconv.Atoi(param)
		if err != nil {
			logs.Error("parameter error by %v", err)
			c.String(200, "Failed")
			return
		}
		if index >= field.Len() {
			logs.Error("parameter: %s error", param)
			c.String(200, "Failed")
			return
		}
		field = field.Index(index)
	}
	if field.Kind() != reflect.Ptr && field.Kind() != reflect.Slice {
		field = field.Addr()
	}
	if v, ok := field.Interface().(httpDel); ok {
		if err := v.delRank(acid); err != nil {
			c.String(200, "Failed")
		} else {
			c.String(200, "Success")
		}

	} else {
		logs.Error("convert failed")
		c.String(200, "Failed")
		return
	}
}
