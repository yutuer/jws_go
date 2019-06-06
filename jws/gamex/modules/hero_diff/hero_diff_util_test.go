package hero_diff

import (
	"fmt"
	"testing"
)

func TestGetRandomStageSeq(t *testing.T) {
	slice := getRandomStageSeq([]int{1, 2, 3, 4}, []int{2, 3, 4, 5})
	if len(slice) != 14 {
		t.Fail()
	}
	fmt.Println(slice)
}

func TestGetRandomStage(t *testing.T) {
	slice, max := getRandomStage([]int{4, 3, 4, 5, 6, 7, 3}, 3)
	if max == 3 {
		t.Fail()
	}
	fmt.Println(slice)
}
