package roleinfo

import "strings"

//传入一个2006-01-02，返回二零零六年一月二日
func getChineseDate(info string) string {
	Datedata := strings.Split(info, "-")
	year := ""
	for i := 0; i < 4; i++ {
		year = year + numberToWord(Substr(Datedata[0], i, 1))
	}
	year = year + "年"
	if Substr(Datedata[1], 0, 1) == "0" {
		Datedata[1] = Substr(Datedata[1], 1, 1)
	}
	month := numberToWord(Datedata[1]) + "月"
	if Substr(Datedata[2], 0, 1) == "0" {
		Datedata[1] = Substr(Datedata[1], 1, 1)
	}
	day := numberToWord(Datedata[2]) + "日"

	return year + month + day

}

//传入一个数字number，返回一个汉字
func numberToWord(info string) string {
	switch info {
	case "0":
		return "零"
	case "1":
		return "一"
	case "2":
		return "二"
	case "3":
		return "三"
	case "4":
		return "四"
	case "5":
		return "五"
	case "6":
		return "六"
	case "7":
		return "七"
	case "8":
		return "八"
	case "9":
		return "九"
	case "10":
		return "十"
	case "11":
		return "十一"
	case "12":
		return "十二"
	case "13":
		return "十三"
	case "14":
		return "十四"
	case "15":
		return "十五"
	case "16":
		return "十六"
	case "17":
		return "十七"
	case "18":
		return "十八"
	case "19":
		return "十九"
	case "20":
		return "二十"
	case "21":
		return "二十一"
	case "22":
		return "二十二"
	case "23":
		return "二十三"
	case "24":
		return "二十四"
	case "25":
		return "二十五"
	case "26":
		return "二十六"
	case "27":
		return "二十七"
	case "28":
		return "二十八"
	case "29":
		return "二十九"
	case "30":
		return "三十"
	case "31":
		return "三十一"
	default:
		return ""
	}
}

//轮子：截取字符串，输入参数为str：字符串本身，index:不定长度参数，传入一个时从index[0]截取到最后，传入两个时从index[0]开始截，截取index[1]位
func Substr(str string, index ...int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0
	if len(index) == 1 { //没有第二个参数时，将其置为最大
		index = append(index, len(rs)-index[0])
	}

	if index[0] < 0 {
		index[0] = rl - 1 + index[0]
	}
	end = index[0] + index[1]

	if index[0] > end {
		index[0], end = end, index[0]
	}

	if index[0] < 0 {
		index[0] = 0
	}
	if index[0] > rl {
		index[0] = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[index[0]:end])

}
