package account

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"sync"
	"time"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	regular  = `^1[3578]\d{9}$`
	regPhone = regexp.MustCompile(regular)
)

const smsContextMsg = "本次验证码为%s，您正在极无双手游中绑定手机号码，祝您游戏愉快！"

const OnePhoneSMSCountPreDay = 5
const OneAccountSMSCountPreDay = 5

var (
	phoneHistory               map[string]int
	phoneHistoryLock           sync.RWMutex
	phoneHistoryLastUpdateTime int64
)

var (
	ErrPhoneFormat              = errors.New("phoneFormatErr")
	ErrPhoneSmsTooMuch          = errors.New("phoneSmsTooMuchErr")
	ErrPhoneHasGetReward        = errors.New("ErrPhoneHasGetReward")
	ErrPhoneSmsTooMuchByAccount = errors.New("ErrPhoneSmsTooMuchByAccount")
	ErrPhoneSmsTooFastByAccount = errors.New("ErrPhoneSmsTooMuchByAccount")
)

func InitPhoneHistory() {
	phoneHistoryLock.Lock()
	defer phoneHistoryLock.Unlock()
	phoneHistory = make(map[string]int, 5120)
	phoneHistoryLastUpdateTime = time.Now().Unix()
}

type PlayerPhoneData struct {
	Phone             string `redis:"phone"`
	PhoneCode         string `redis:"phoneCode"`
	LastGetCodeTime   int64  `redis:"phoneCodeT"`
	LastSendCountTime int64  `redis:"ct"`
	SendCountCurrDay  int    `redis:"c"`
	HasGot            bool   `redis:"has"`
}

func (p *PlayerPhoneData) checkPhoneNum(phone string) bool {
	phoneHistoryLock.Lock()
	defer phoneHistoryLock.Unlock()

	nowT := time.Now().Unix()
	if !gamedata.IsSameDayCommon(nowT, phoneHistoryLastUpdateTime) {
		phoneHistory = make(map[string]int, 10240)
		phoneHistoryLastUpdateTime = nowT
	}

	pc, ok := phoneHistory[phone]
	if ok {
		if pc >= OnePhoneSMSCountPreDay {
			return false
		} else {
			phoneHistory[phone] = pc + 1
		}
	} else {
		phoneHistory[phone] = 1
	}
	return true

}

func (p *PlayerPhoneData) IsHasBindPhone() bool {
	return p.LastGetCodeTime == 0 && p.Phone != ""
}

func (p *PlayerPhoneData) SetHasBindPhone(phone string) {
	p.Phone = phone
	p.LastGetCodeTime = 0
	p.HasGot = true
}

func (p *PlayerPhoneData) IsCanGetCode(nowT int64) error {
	if p.IsHasBindPhone() {
		return ErrPhoneHasGetReward
	}

	// 一段时间内只能申请一个码
	if nowT-p.LastGetCodeTime < 60 {
		return ErrPhoneSmsTooFastByAccount
	}

	if !gamedata.IsSameDayCommon(nowT, p.LastSendCountTime) {
		p.SendCountCurrDay = 0
		p.LastSendCountTime = nowT
	}

	if p.SendCountCurrDay >= OneAccountSMSCountPreDay {
		return ErrPhoneSmsTooMuchByAccount
	}

	return nil
}

func (p *PlayerPhoneData) GetCode(phone string, nowT int64, rd *rand.Rand) error {
	// 6位数字
	if !regPhone.MatchString(phone) {
		return ErrPhoneFormat
	}

	if !p.checkPhoneNum(phone) {
		return ErrPhoneSmsTooMuch
	}

	c := fmt.Sprintf("%06d", rd.Int31n(1000000))

	logs.Trace("playerPhoneData GetCode %s", c)

	if !gamedata.IsSameDayCommon(nowT, p.LastSendCountTime) {
		p.SendCountCurrDay = 0
		p.LastSendCountTime = nowT
	}
	p.SendCountCurrDay += 1

	p.LastGetCodeTime = nowT
	p.PhoneCode = c
	p.Phone = phone

	return util.SendHeroSMS(phone, fmt.Sprintf(smsContextMsg, c))
}
