package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
)

//BuyInStore ..
func (p *Account) BuyInStore(r servers.Request) *servers.Response {
	req := &struct {
		Req
		StoreID uint32 `codec:"s"`
		BlankID uint32 `codec:"b"`
		Count   uint32 `codec:"c"`
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}
	initReqRsp(
		"PlayerAttr/BuyInStoreResponse",
		r.RawBytes,
		req, resp, p)

	if req.Count > uutil.CHEAT_INT_MAX {
		return rpcErrorWithMsg(resp, 99, "BuyInStore count cheat")
	}

	ok, warnCode, chg := p.StoreProfile.Buy(req.StoreID, req.BlankID, req.Count, p.Account, resp)
	if !ok {
		if warnCode > 0 {
			return rpcWarn(resp, warnCode)
		}
		return rpcError(resp, 1)
	}
	// 差量更新商店
	for storeID, changed := range chg {
		if true == changed {
			resp.addSyncStore(storeID)
		}
	}
	resp.addSyncStoreBlank(req.StoreID, req.BlankID)
	resp.OnChangeSC()
	resp.OnChangeHC()

	p.updateCondition(account.COND_TYP_Buy_In_Store,
		1, int(req.StoreID), "", "", resp)

	resp.mkInfo(p)

	return rpcSuccess(resp)
}

//RequestRefreshStore ..
type RequestRefreshStore struct {
	Req
	StoreID uint32 `codec:"s"`
}

//ResponseRefreshStore ..
type ResponseRefreshStore struct {
	SyncResp
}

//RefreshStore ..
func (p *Account) RefreshStore(r servers.Request) *servers.Response {
	req := &RequestRefreshStore{}
	resp := &ResponseRefreshStore{}

	initReqRsp(
		"PlayerAttr/RefreshStoreResponse",
		r.RawBytes,
		req, resp, p)

	ok, chg := p.StoreProfile.Refresh(req.StoreID, p.Account, resp)

	// 差量更新商店
	for storeID, changed := range chg {
		if true == changed {
			resp.addSyncStore(storeID)
		}
	}

	if !ok {
		resp.mkInfo(p)
		return rpcError(resp, 1)
	}
	resp.OnChangeSC()
	resp.OnChangeHC()

	resp.mkInfo(p)

	return rpcSuccess(resp)
}
