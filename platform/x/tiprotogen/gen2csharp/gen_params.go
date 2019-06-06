package gen2csharp

import (
	"strings"

	"fmt"

	"vcs.taiyouxi.net/platform/x/tiprotogen/def"
	"vcs.taiyouxi.net/platform/x/tiprotogen/util"
)

const (
	paramTypNil = iota
	paramTypLong
	paramTypString
	paramTypByte
	paramTypLongArray
	paramTypStringArray
	paramTypByteArray
	paramTypNetObj
	paramTypeNetObjArray
	paramTypeBool
	paramTypeBoolArray
)

type paramTypCSharp struct {
	typ int
	str string
}

func getParamTypClassFromStr(t string) int {
	ts := strings.Replace(t, " ", "", -1)
	ts = strings.Replace(ts, "	", "", -1)
	ts = strings.ToLower(ts)
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

func getStrFromParamTypClass(t int, str string) string {
	switch t {
	case paramTypLong:
		return "long"
	case paramTypString:
		return "string"
	case paramTypLongArray:
		return "long[]"
	case paramTypStringArray:
		return "string[]"
	case paramTypNetObj:
		return str
	case paramTypByte:
		return "byte"
	case paramTypByteArray:
		return "byte[]"
	case paramTypeNetObjArray:
		return str
	case paramTypeBool:
		return "bool"
	case paramTypeBoolArray:
		return "bool[]"
	default:
		return ""
	}
}

func NewParamTypCSharp(t string) paramTypCSharp {
	return paramTypCSharp{
		typ: getParamTypClassFromStr(t),
		str: t,
	}
}

func (p *paramTypCSharp) IsNil() bool {
	return p.typ == paramTypNil
}

func (p *paramTypCSharp) IsArray() bool {
	return p.typ == paramTypStringArray ||
		p.typ == paramTypLongArray ||
		p.typ == paramTypByteArray ||
		p.typ == paramTypeBoolArray
}

func (p *paramTypCSharp) IsObjArray() bool {
	return p.typ == paramTypeNetObjArray
}

func (p *paramTypCSharp) IsObj() bool {
	return p.typ == paramTypNetObj
}

func (p *paramTypCSharp) GetArrayItemType() paramTypCSharp {
	switch p.typ {
	case paramTypLongArray:
		return paramTypCSharp{
			typ: paramTypLong,
		}
	case paramTypStringArray:
		return paramTypCSharp{
			typ: paramTypString,
		}
	case paramTypByteArray:
		return paramTypCSharp{
			typ: paramTypByte,
		}
	case paramTypeBoolArray:
		return paramTypCSharp{
			typ: paramTypeBool,
		}
	case paramTypeNetObjArray:
		return paramTypCSharp{
			typ: paramTypNetObj,
			str: p.TypeName(),
		}
	default:
		return paramTypCSharp{}
	}
}

func (p *paramTypCSharp) TypeName() string {
	return getStrFromParamTypClass(p.typ, p.str)
}

func (g *genner2Csharp) genReqParams(
	buf *util.CodeGenData,
	params []dsl.ProtoParam) {
	for _, param := range params {
		buf.WriteLine(
			paramTemplate,
			param.Type,
			param.ShortName,
			param.Comment)
	}
}

func (g *genner2Csharp) genRspParams(
	buf *util.CodeGenData,
	params []dsl.ProtoParam) {
	for _, param := range params {
		t := NewParamTypCSharp(param.Type)
		paramTyp := param.Type
		if t.IsArray() || t.IsObjArray() {
			paramTyp = "List<object>"
		} else if t.IsObj() {
			paramTyp = "byte[]"
		}

		buf.WriteLine(
			paramTemplate,
			paramTyp,
			param.ShortName,
			param.Comment)
	}
}

func (g *genner2Csharp) genParamCalls(params []dsl.ProtoParam) string {
	res := make([]byte, 0, 256)
	if params == nil || len(params) == 0 {
		return ""
	}

	res = append(res, ", "...)

	for _, p := range params {
		res = append(res, p.Type...)
		res = append(res, " "...)
		res = append(res, p.Name...)
		res = append(res, ", "...)
	}
	res = res[:len(res)-2]
	return string(res)
}

func (g *genner2Csharp) genParamsInit(
	buf *util.CodeGenData,
	params []dsl.ProtoParam) {
	for _, param := range params {
		buf.WriteLine(
			paramInit,
			param.ShortName,
			param.Name)
	}
}

func (g *genner2Csharp) genParamsDataDef(
	buf *util.CodeGenData,
	params []dsl.ProtoParam) {
	for _, param := range params {
		t := NewParamTypCSharp(param.Type)
		if t.IsNil() {
			util.PanicInfo(
				"param type nil %v",
				param)
		}
		buf.WriteLine(
			msgDataDefParam,
			t.TypeName(),
			param.Name,
			param.Comment)
	}
}

func (g *genner2Csharp) genParamsDataDefNoTab(
	buf *util.CodeGenData,
	params []dsl.ProtoParam) {
	for _, param := range params {
		t := NewParamTypCSharp(param.Type)
		if t.IsNil() {
			util.PanicInfo(
				"param type nil %v",
				param)
		}
		buf.WriteLine(
			msgDataDefParamWithoutTab,
			t.TypeName(),
			param.Name,
			param.Comment)
	}
}

func (g *genner2Csharp) genParamsDataGet(
	buf *util.CodeGenData,
	params []dsl.ProtoParam) {
	for _, param := range params {
		fmt.Println("cs params:", params)
		t := NewParamTypCSharp(param.Type)
		if t.IsNil() {
			util.PanicInfo(
				"param type nil %v",
				param)
		} else if t.IsArray() {
			tItem := t.GetArrayItemType()
			buf.WriteLine(
				msgDataGetArrayParam,
				param.Name,
				"MkArrayFromObjList",
				tItem.TypeName(),
				param.ShortName)
		} else if t.IsObjArray() {
			tItem := t.GetArrayItemType()
			typeName := tItem.TypeName()[:len(tItem.TypeName())-2]
			buf.WriteLine(
				msgNetDataGetArrayParam,
				typeName,
				param.Name,
				"MkArrayFromNet",
				typeName,
				param.ShortName,
				param.Name,
				typeName,
				param.Name,
				param.Name,
				param.Name,
				param.Name)
		} else if t.IsObj() {
			buf.WriteLine(
				msgDataGetObjParam,
				param.Type,
				param.Type,
				param.ShortName,
				param.Name)
		} else {
			buf.WriteLine(
				msgDataGetParam,
				param.Name,
				param.ShortName)
		}

	}
}
