package util

import "math/rand"

// 生成一个0到n-1的数组, 并将其随机洗牌
func Shuffle1ToN(n int) []int {
	return ShuffleN(0, n)
}

// 生成一个0到n-1的数组, 并将其随机洗牌
func Shuffle1ToNSelf(n int, rd *rand.Rand) []int {
	return ShuffleNByRand(0, n, rd)
}

func ShuffleArrayByRand(res []int, rd *rand.Rand) []int {
	n := len(res)
	for j := n; j > 1; j-- {
		randomIndex := rd.Intn(j)
		res[j-1], res[randomIndex] = res[randomIndex], res[j-1]
	}
	return res[:]
}

func ShuffleNByRand(n1, n2 int, rd *rand.Rand) []int {
	if n1 >= n2 {
		return nil
	}
	res := make([]int, n2-n1)
	for i := 0; i < n2-n1; i++ {
		res[i] = i + n1
	}
	return ShuffleArrayByRand(res, rd)
}

func ShuffleArray(res []int) []int {
	n := len(res)
	for j := n; j > 1; j-- {
		randomIndex := rand.Intn(j)
		res[j-1], res[randomIndex] = res[randomIndex], res[j-1]
	}
	return res[:]
}

// [n1, n2)
func ShuffleN(n1, n2 int) []int {
	if n1 >= n2 {
		return nil
	}
	res := make([]int, n2-n1)
	for i := 0; i < n2-n1; i++ {
		res[i] = i + n1
	}
	return ShuffleArray(res)
}
