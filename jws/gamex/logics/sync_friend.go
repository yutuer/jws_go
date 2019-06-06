package logics

import (
	"vcs.taiyouxi.net/jws/gamex/modules/friend"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
)

//每隔一段时间就会把好有信息同步给客户端  唉，都是眼泪，
func (s *SyncResp) mkFriendInfo(p *Account) {
	nowT := p.GetProfileNowTime()
	f := &p.Friend
	if s.update_black_list_sync || nowT-f.LastUpdateBlackListTime > friend.UpdateBlackListInterval {
		s.SyncBlackListNeed = true
		needNew := nowT-f.LastUpdateBlackListTime > friend.UpdateBlackListInterval
		if needNew {
			newBlackList := friend.GetModule(p.AccountID.ShardId).UpdateFriendList(f.GetBlackList(), false)
			f.SetBlackList(newBlackList)
			f.LastUpdateBlackListTime = nowT
		}
		s.SyncBlackList = make([][]byte, 0, len(f.GetBlackList()))
		for _, item := range f.GetBlackList() {
			if player_msg.GetModule(p.AccountID.ShardId).IsOnline(item.AcID) {
				item.LastActTime = 0
			}
			s.SyncBlackList = append(s.SyncBlackList, encode(item))
		}
	}

	if s.update_friend_list_sync || nowT-f.LastUpdateFriendListTime > friend.UpdateFriendListInterval {
		s.SyncFriendListNeed = true
		needNew := nowT-f.LastUpdateFriendListTime > friend.UpdateFriendListInterval
		if needNew {
			newFriendList := friend.GetModule(p.AccountID.ShardId).UpdateFriendList(f.GetFriendList(), false)
			f.SetFriendList(newFriendList)
			f.LastUpdateFriendListTime = nowT
		}
		s.SyncFriendList = make([][]byte, 0, len(f.GetFriendList()))
		for _, item := range f.GetFriendList() {
			if player_msg.GetModule(p.AccountID.ShardId).IsOnline(item.AcID) {
				item.LastActTime = 0
			}
			s.SyncFriendList = append(s.SyncFriendList, encode(item))
		}
	}
	f.RefreshGiftTimes(nowT)
	if s.SyncFaceBookNeed {
		s.SyncIsFaceBook = p.Profile.IsFaceBook
	}
	if s.SyncTwitterNeed {
		s.SyncTwitterIsShare = p.Profile.IsTwitterShared
	}
	if s.SyncLineNeed {
		s.SyncLineIsShare = p.Profile.IsLineShared
	}

	if s.SyncFriendGiftListNeed {
		s.SyncFriendGift = p.Friend.GetFriendGiftInfo()
	}
}
