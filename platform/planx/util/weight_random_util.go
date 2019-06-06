package util

import (
	"math/rand"
)

type WeightedItems interface {
	GetWeight(index int) int
	Len() int
}

func RandomItem(data WeightedItems) int {
	if data.Len() == 0 {
		return 0
	}
	weightArray := make([]int, data.Len())
	weightArray[0] = data.GetWeight(0)
	for i := 1; i < data.Len(); i++ {
		weightArray[i] = weightArray[i-1] + data.GetWeight(i)
	}
	randomWeight := rand.Int31n(int32(weightArray[data.Len()-1]))
	for i := 0; i < data.Len(); i++ {
		if randomWeight < int32(weightArray[i]) {
			return i
		}
	}
	return 0
}

func RandomItemByCount(data WeightedItems, count int) []int {
	if data.Len() == 0 || count <= 0 {
		return nil
	}
	weightArray := make([]int, data.Len())
	weightArray[0] = data.GetWeight(0)
	for i := 1; i < data.Len(); i++ {
		weightArray[i] = weightArray[i-1] + data.GetWeight(i)
	}
	result := make([]int, count)
	for c := 0; c < count; c++ {
		randomWeight := rand.Int31n(int32(weightArray[data.Len()-1]))
		for i := 0; i < data.Len(); i++ {
			if randomWeight < int32(weightArray[i]) {
				result[c] = i
				break
			}
		}
	}
	return result
}

func RandomInt(start, end int32) uint32 {
	if start == end {
		return uint32(start)
	}
	return uint32(rand.Int31n(end-start) + start)
}

func RandomInts(start, end, count int) []int {
	if end-start < count {
		return nil
	}
	tempInts := ShuffleN(start, end)
	return tempInts[:count]
}

type IntSlice []int

func (p IntSlice) Len() int            { return len(p) }
func (p IntSlice) GetWeight(i int) int { return p[i] }

func RandomWeightInts(weightArray []int) int {
	array := IntSlice(weightArray)
	return RandomItem(array)
}
