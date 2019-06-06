package builderjson

import (
	"encoding/json"
	"errors"
	"fmt"

	"vcs.taiyouxi.net/platform/x/tiprotogen/def"
)

type builderFromJson struct {
}

func New() *builderFromJson {
	return new(builderFromJson)
}

// fileName, 消息数组
type JsonMap map[string][]dsl.ProtoDef

func (b *builderFromJson) Build(data []byte) ([]dsl.ProtoDef, error) {
	fmt.Println(string(data[:1]))
	if string(data[:1]) == "{" {
		return b.buildObject(data)
	} else if string(data[:1]) == "[" {
		return b.buildArray(data)
	} else {
		return nil, errors.New("error json format")
	}
}

func (b *builderFromJson) buildArray(data []byte) ([]dsl.ProtoDef, error) {
	jsonData := make([]protoJson, 0)
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		fmt.Println("", err)
		return nil, err
	}

	fmt.Println("datatata:", string(data))
	fmt.Println("json len", len(jsonData))
	return protoJsonArray(jsonData).ToDef(), nil
}

func (b *builderFromJson) buildObject(data []byte) ([]dsl.ProtoDef, error) {
	jsonData := new(protoJson)
	err := json.Unmarshal(data, jsonData)
	if err != nil {
		return nil, err
	}
	fmt.Println("datatata:", string(data))

	return []dsl.ProtoDef{jsonData.ToDef()}, nil
}
