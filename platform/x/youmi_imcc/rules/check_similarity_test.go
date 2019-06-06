package rules

import (
	"fmt"
	"testing"
)

func TestCheckSimularity(t *testing.T) {
	similar := getSimilarity("差距夥計左在天水 高戰力來支援fadfadf",
		"差距夥計左在天水")
	fmt.Println(similar)
	if similar > 0.8 {
		t.FailNow()
	}
}
