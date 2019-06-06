package util

import (
	"fmt"
	"testing"
)

func TestRandom(t *testing.T) {
	testList := TestItemList{}
	testList.list = make([]int, 10)
	for i := 0; i < testList.Len(); i++ {
		testList.list[i] = i
	}
	index := RandomItem(&testList)
	fmt.Println(testList.list[index])

	ret := RandomItemByCount(&testList, 10)
	fmt.Println(ret)
}

type TestItemList struct {
	list []int
}

func (t *TestItemList) GetWeight(index int) int {
	return t.list[index]
}

func (t *TestItemList) Len() int {
	return len(t.list)
}

func TestShuffleArray(t *testing.T) {
	for i := 0; i < 100; i++ {
		array := []int{0, 0, 1, 1, 2}
		fmt.Println(ShuffleArray(array))
	}
}

func TestRandomWeightInts(t *testing.T) {
	for i := 0; i < 10; i++ {
		index := RandomWeightInts([]int{55, 5, 20, 10})
		fmt.Println(index)
	}
}
