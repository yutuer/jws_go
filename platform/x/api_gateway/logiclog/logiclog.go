package logiclog

import (
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const BITag = "[BI]"

type payInfo struct {
	SdkChannelUid    string
	SdkOrderNo       string
	SdkPayTime       string
	Money            string
	Success          bool
	SdkStatus        string
	SdkIsTest        string
	SdkNote          string
	GameOrderNo      string
	GameExtrasParams string
	OrderIdx         string
	Tistatus         string
	ClientPayTime    string
	ProductId        string
	ClientVer        string
	VNSDKType        string
	PkgInfo          string
}

func LogPay(typ, uid string, success bool,
	sdkChannel, sdkChannelUid, sdkOrderNo, sdkPayTime, money, sdkStatus, sdkIsTest, sdkNote string,
	gameOrderNo, gameExtrasParams, orderIdx, tistatus, clientPayTime, productId, clientVer string, pkginfo string) {
	logPay(typ, uid, success,
		sdkChannel, sdkChannelUid, sdkOrderNo, sdkPayTime, money, sdkStatus, sdkIsTest, sdkNote,
		gameOrderNo, gameExtrasParams, orderIdx, tistatus, clientPayTime, productId, clientVer, "", pkginfo)
}

func LogVNPay(typ, uid string, success bool,
	sdkChannel, sdkChannelUid, sdkOrderNo, sdkPayTime, money, sdkStatus, sdkIsTest, sdkNote string,
	gameOrderNo, gameExtrasParams, orderIdx, tistatus, clientPayTime, productId, clientVer string, vnSdkTyp string) {
	logPay(typ, uid, success,
		sdkChannel, sdkChannelUid, sdkOrderNo, sdkPayTime, money, sdkStatus, sdkIsTest, sdkNote,
		gameOrderNo, gameExtrasParams, orderIdx, tistatus, clientPayTime, productId, clientVer, vnSdkTyp, "")
}

func LogKoPay(typ, uid string, success bool,
	sdkChannel, sdkChannelUid, sdkOrderNo, sdkPayTime, money, sdkStatus, sdkIsTest, sdkNote string,
	gameOrderNo, gameExtrasParams, orderIdx, tistatus, clientPayTime, productId, clientVer string, pkginfo string) {
	logPay(typ, uid, success,
		sdkChannel, sdkChannelUid, sdkOrderNo, sdkPayTime, money, sdkStatus, sdkIsTest, sdkNote,
		gameOrderNo, gameExtrasParams, orderIdx, tistatus, clientPayTime, productId, clientVer, "", pkginfo)
}

func logPay(typ, uid string, success bool,
	sdkChannel, sdkChannelUid, sdkOrderNo, sdkPayTime, money, sdkStatus, sdkIsTest, sdkNote string,
	gameOrderNo, gameExtrasParams, orderIdx, tistatus, clientPayTime, productId, clientVer string, vnSdkTyp string, pkginfo string) {
	r := payInfo{
		SdkChannelUid:    sdkChannelUid,
		SdkOrderNo:       sdkOrderNo,
		SdkPayTime:       sdkPayTime,
		Money:            money,
		SdkStatus:        sdkStatus,
		SdkIsTest:        sdkIsTest,
		SdkNote:          sdkNote,
		GameOrderNo:      gameOrderNo,
		GameExtrasParams: gameExtrasParams,
		OrderIdx:         orderIdx,
		Tistatus:         tistatus,
		ClientPayTime:    clientPayTime,
		ProductId:        productId,
		ClientVer:        clientVer,
		Success:          success,
		VNSDKType:        vnSdkTyp,
		PkgInfo:          pkginfo,
	}
	logs.Trace("LogPay %s %v", BITag, r)
	logiclog.Error(uid, 0, 0, sdkChannel, typ, r, "", BITag)
}
