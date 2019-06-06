package command

import (
	"io/ioutil"

	"fmt"
	"os"

	"vcs.taiyouxi.net/platform/x/tiprotogen/builderjson"
	"vcs.taiyouxi.net/platform/x/tiprotogen/def"
	"vcs.taiyouxi.net/platform/x/tiprotogen/gen2csharp"
	"vcs.taiyouxi.net/platform/x/tiprotogen/gen2golang"
	"vcs.taiyouxi.net/platform/x/tiprotogen/log"
	"vcs.taiyouxi.net/platform/x/tiprotogen/util"
)

func LoadFromFile(path string) ([]dsl.ProtoDef, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		util.PanicInfo("LoadFormFile %s err By %s",
			path, err.Error())
	}

	b := builderjson.New()
	return b.Build(bytes)
}

func GenCSharpCode(path string, data *dsl.ProtoDef) {
	g := gen2csharp.New()
	bytes := g.Gen(data)
	os.MkdirAll(path, os.ModePerm)
	ioutil.WriteFile(path+data.GetFileName()+".cs", bytes, 0666)

}

func GenMultipleCSharpCode(path string, dataArray []dsl.ProtoDef) {
	fmt.Println("dataArray:", dataArray)
	for i := range dataArray {
		GenCSharpCode(path, &dataArray[i])
	}
}

func GenGolangCode(path string, data *dsl.ProtoDef) {
	g := gen2golang.New()
	bytes := g.Gen(data)
	pathAll := path + data.GetFileName() + ".go"
	//info, _ := os.Stat(pathAll)
	//if info != nil {
	//	log.Err("This has an exit file!!!, gen stop %s", pathAll)
	//	return
	//}
	os.MkdirAll(path, os.ModePerm)
	ioutil.WriteFile(pathAll, bytes, 0666)
}

func GenGolangPathCode(path string, data []*dsl.ProtoDef) {
	bytes := gen2golang.GenPathFile(data)
	pathAll := path + ".go"
	os.MkdirAll(path, os.ModePerm)
	ioutil.WriteFile(pathAll, bytes, 0666)
}

func GenGolangCodes(outPath string, fileName string, dataArray []dsl.ProtoDef) {
	g := gen2golang.New()
	bytes := make([]byte, 0)
	headBytes := g.GenMultiFileHeader()
	bytes = append(bytes, headBytes...)
	for i := range dataArray {
		fmt.Println("gen golang file", dataArray[i].Name)
		fmt.Println("data object:", dataArray[i].Object)
		bytes = append(bytes, g.GenMultiFileBody(&dataArray[i])...)
	}
	pathAll := outPath + fileName + ".go"
	os.MkdirAll(outPath, os.ModePerm)
	ioutil.WriteFile(pathAll, bytes, 0666)
}

func GenGolangHandlers(outPath string, fileName string, dataArray []dsl.ProtoDef) {
	g := gen2golang.New()
	bytes := make([]byte, 0)
	headBytes := g.GenHandlerFileHeader()
	bytes = append(bytes, headBytes...)
	for i := range dataArray {
		fmt.Println("gen golang file", dataArray[i].Name)
		bytes = append(bytes, g.GenHandler(&dataArray[i])...)
	}
	pathAll := outPath + fileName + ".go"
	os.MkdirAll(outPath, os.ModePerm)
	ioutil.WriteFile(pathAll, bytes, 0666)
}

func AppendGolangPathCode(path string, data []*dsl.ProtoDef) {
	regFile, err := ioutil.ReadFile(path + "path_reg_by_gens.go")
	if err != nil {
		log.Err("read path_reg_by_gens error, %v", err)
		return
	}
	regFile = regFile[:len(regFile)-1]
	bytes := gen2golang.AppendRegPathFile(data)
	fmt.Println("path len", len(bytes))
	regFile = append(regFile, bytes...)
	regFile = append(regFile, byte(125))
	ioutil.WriteFile(path+"path_reg_by_gens.go", regFile, 0666)
}
