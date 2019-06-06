package builderjson

import (
	"fmt"
	"strings"

	"vcs.taiyouxi.net/platform/x/tiprotogen/def"
	"vcs.taiyouxi.net/platform/x/tiprotogen/util"
)

func toParamDef(data [][]string) []dsl.ProtoParam {
	res := make([]dsl.ProtoParam, 0, len(data))
	for _, paramArray := range data {
		if len(paramArray) < 3 {
			util.PanicInfo("param num err by %v", paramArray)
		}

		p := dsl.ProtoParam{
			Name:    paramArray[0],
			Type:    paramArray[1],
			Comment: paramArray[2],
		}

		if len(paramArray) >= 4 {
			p.ShortName = paramArray[3]
		} else {
			p.ShortName = strings.ToLower(p.Name)
		}

		res = append(res, p)
	}
	fmt.Println("res:", res)
	return res
}
