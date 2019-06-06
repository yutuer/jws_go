package redeemCode

import "math/rand"

// 1      2        3        4   5     6       7    8     9
// 9      4        99999    9   9999  9       99   7     9
// 1      4        00000    0   0000  0       00   7     0
// 循环序号 固定随机  序号     随机 批号  随机     组号  固定  随机

func mkRandNum(start, end int64) int64 {
	return rand.Int63n(end-start) + start
}

const (
	constNumInGiftCodeNum1         = 4
	constNumInGiftCodeNum2_Normal  = 7
	constNumInGiftCodeNum2_NoLimit = 1
	nums                           = "DUVWOPQRSTXYZFGHIJKALBECMN"
	maxCount                       = 100000
)

var cc int64

func revNum(i, m int64) int64 {
	return m - i - 1
}

func mkGiftCodeNum(batchID, typID, count int64, isNoLimit bool) int64 {
	var res int64

	// 1 循环序号 1-9
	res = cc%9 + 1
	cc++

	// 2 固定
	res = res*10 + mkRandNum(1, 5)

	// 3 序号
	res = res*100000 + revNum(count, 100000)

	// 4 随机
	res = res*10 + mkRandNum(1, 10)

	// 5 批号
	res = res*10000 + batchID

	// 6 随机
	res = res*10 + mkRandNum(1, 10)

	// 7 组号
	res = res*100 + revNum(typID, 100)

	// 8 固定
	if isNoLimit {
		res = res*10 + constNumInGiftCodeNum2_NoLimit
	} else {
		res = res*10 + constNumInGiftCodeNum2_Normal
	}
	// 9 随机
	res = res*10 + mkRandNum(1, 10)

	return res
}

func parseGiftCodeNum(num int64) (batchID, typID, count int64, isNoLimit, ok bool) {
	num = num / 10
	n8 := num % 10
	num = num / 10
	n7 := revNum(num%100, 100)
	num = num / 100
	num = num / 10
	n5 := num % 10000
	num = num / 10000
	num = num / 10
	n3 := revNum(num%100000, 100000)
	num = num / 100000
	n2 := num % 10
	num = num / 10

	batchID = n5
	typID = n7
	count = n3
	if n8 == constNumInGiftCodeNum2_NoLimit {
		ok = true
		isNoLimit = true
	} else if n8 == constNumInGiftCodeNum2_Normal {
		ok = true
		isNoLimit = false
	} else {
		ok = false
	}
	ok = ok && (n2 <= constNumInGiftCodeNum1)
	return
}

func toNumInStr(i int64) string {
	numLen := int64(len(nums))

	res := ""

	j := i
	for j > 0 {
		n := j % numLen
		j = j / numLen

		res = string(nums[n]) + res
	}

	return res
}

func fromStrToNum(s string) int64 {
	numLen := int64(len(nums))

	var res int64

	chars := make(map[string]int64, numLen)

	for num, c := range nums {
		chars[string(c)] = int64(num)
	}
	//fmt.Printf("n: %v\n", chars)

	for _, char := range s {
		n, ok := chars[string(char)]
		if !ok {
			return -1
		}
		//fmt.Printf("%s 2: %2d %d\n",string(char),  n, res)
		res = res*numLen + n
	}

	return res
}

//Gen 批量生成兑换码
func Gen(batchID, typID, count int64) []string {
	if count >= maxCount {
		return nil
	}
	res := make([]string, 0, 10000)
	var c int64
	for ; c < count; c++ {
		id := mkGiftCodeNum(batchID, typID, c, false)
		res = append(res, toNumInStr(id))
	}
	return res
}

//Parse 解析兑换码
func Parse(code string) (batchID, typID, count int64, noLimit, ok bool) {
	n := fromStrToNum(code)
	ok = false
	b, t, c, is_no_limit, is_ok := parseGiftCodeNum(n)
	if !is_ok {
		return
	}
	if b <= 0 || t <= 0 || c < 0 || c >= maxCount {
		return
	}
	batchID = b
	typID = t
	count = c
	ok = true
	noLimit = is_no_limit
	return
}
