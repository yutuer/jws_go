package main

import (
	"fmt"
	"testing"
)

func TestGenSpecificConfigCode(t *testing.T) {
	genCode := genSpecificConfigCode("ACTIVITYCONFIG", "uint32", "ActivityTimes")
	fmt.Println(genCode)

	genCode = genSpecificConfigCode("LEVEL_INFO", "string", "LevelID")
	fmt.Println(genCode)
}
