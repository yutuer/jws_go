package gen2golang

import (
	"strings"

	"fmt"

	"vcs.taiyouxi.net/platform/x/tiprotogen/def"
	"vcs.taiyouxi.net/platform/x/tiprotogen/util"
)

const (
	paramTypNil          = ""
	paramTypLong         = "int64"
	paramTypString       = "string"
	paramTypLongArray    = "[]int64"
	paramTypStringArray  = "[]string"
	paramTypNetObj       = "[]byte"
	paramTypByte         = "byte"
	paramTypByteArray    = "[]byte"
	paramTypeNetObjArray = "[][]byte"
	paramTypeBool        = "bool"
	paramTypeBoolArray   = "[]bool"
)

type paramTypGolang struct {
	typ string
}

func getParamTypClassFromStr(t string, transfer bool) string {
	ts := strings.Replace(t, " ", "", -1)
	ts = strings.Replace(ts, "	", "", -1)
	//log.Trace("getParamTypClassFromStr %s %s", t, ts)
	switch ts {
	case "long":
		return paramTypLong
	case "string":
		return paramTypString
	case "long[]":
		return paramTypLongArray
	case "string[]":
		return paramTypStringArray
	case "byte":
		return paramTypByte
	case "byte[]":
		return paramTypByteArray
	case "bool":
		return paramTypeBool
	case "bool[]":
		return paramTypeBoolArray
	case "":
		return paramTypNil
	default:

		if strings.Index(ts, "[]") > 0 {
			return paramTypeNetObjArray
		} else {
			return paramTypNetObj
		}

	}
}

func NewParamTypGolang(t string, transfer bool) paramTypGolang {
	return paramTypGolang{
		typ: getParamTypClassFromStr(t, transfer),
	}
}

func (p *paramTypGolang) IsNil() bool {
	return p.typ == paramTypNil
}

func (p *paramTypGolang) IsArray() bool {
	return p.typ == paramTypStringArray ||
		p.typ == paramTypLongArray ||
		p.typ == paramTypByteArray
}

func (p *paramTypGolang) String() string {
	fmt.Println("p.typ:", p.typ)
	return p.typ
}

func (p *paramTypGolang) GetArrayItemType() paramTypGolang {
	switch p.typ {
	case paramTypLongArray:
		return paramTypGolang{
			typ: paramTypLong,
		}
	case paramTypStringArray:
		return paramTypGolang{
			typ: paramTypString,
		}
	case paramTypByteArray:
		return paramTypGolang{
			typ: paramTypByte,
		}
	default:
		return paramTypGolang{}
	}
}

func (g *genner2golang) genParams(
	buf *util.CodeGenData,
	params []dsl.ProtoParam) {
	for _, param := range params {
		fmt.Println("param:", param)
		t := NewParamTypGolang(param.Type, param.Transfer)
		buf.WriteLine(
			paramTemplate,
			param.Name,
			t.String(),
			param.ShortName,
			param.Comment)
	}
}
