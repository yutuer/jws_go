package serverinfo

const (
	HERO_GY = iota
	HERO_ZF
	HERO_SX
	HERO_LB
	HERO_DC
	HERO_GJ
	HERO_XC
	HERO_ZL
	HERO_XHD
	HERO_ZY
	HERO_XH
	HERO_ZYU
	HERO_ZJ
	HERO_DQ
	HERO_MC
	HERO_XHY
	HERO_ZGL
	HERO_DZ
	HERO_ZH
	HERO_CC
	HERO_XQ
	HERO_ZFB
	HERO_GYB
	HERO_TSC
	HERO_SMY
	HERO_SSX
	HERO_SC
	HERO_YS
	HERO_MH
	HERO_ZR
	HERO_DW
	HERO_ZNJ
	HERO_LM
	HERO_HT
	HERO_SJ
	HERO_HYY
	HERO_ZC
	HERO_SQ
	HERO_YJ
	HERO_HZ
)

func GetHeroName(id int) string {
	switch id {
	case HERO_GY:
		return "关平"
	case HERO_ZF:
		return "张苞"
	case HERO_SX:
		return "刘舞婵"
	case HERO_LB:
		return "吕布"
	case HERO_DC:
		return "貂蝉"
	case HERO_GJ:
		return "郭嘉"
	case HERO_XC:
		return "许褚"
	case HERO_ZL:
		return "张辽"
	case HERO_XHD:
		return "夏侯惇"
	case HERO_ZY:
		return "赵云"
	case HERO_XH:
		return "徐晃"
	case HERO_ZYU:
		return "周瑜"
	case HERO_ZJ:
		return "张角"
	case HERO_DQ:
		return "大乔"
	case HERO_MC:
		return "马超"
	case HERO_XHY:
		return "夏侯渊"
	case HERO_ZGL:
		return "诸葛亮"
	case HERO_DZ:
		return "董卓"
	case HERO_ZH:
		return "张郃"
	case HERO_CC:
		return "曹操"
	case HERO_XQ:
		return "小乔"
	case HERO_ZFB:
		return "张飞"
	case HERO_GYB:
		return "关羽"
	case HERO_TSC:
		return "太史慈"
	case HERO_SMY:
		return "司马懿"
	case HERO_SSX:
		return "孙尚香"
	case HERO_SC:
		return "孙策"
	case HERO_YS:
		return "袁绍"
	case HERO_MH:
		return "孟获"
	case HERO_ZR:
		return "祝融"
	case HERO_DW:
		return "典韦"
	case HERO_ZNJ:
		return "甄姬"
	case HERO_LM:
		return "吕蒙"
	case HERO_HT:
		return "华佗"
	case HERO_SJ:
		return "孙坚"
	case HERO_HYY:
		return "黄月英"
	case HERO_ZC:
		return "左慈"
	case HERO_SQ:
		return "孙权"
	case HERO_YJ:
		return "于吉"
	case HERO_HZ:
		return "黄忠"
	default:
		return ""
	}
}
