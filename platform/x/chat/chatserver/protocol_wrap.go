package chatserver

import (
	"github.com/google/flatbuffers/go"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/chat/chatserver/proto/gen"
)

var (
	protocols map[string]proto_interface
)

const (
	RegReq_Id          = "regreq"
	RegResp_Id         = "regresp"
	Pos_Id             = "posnotify"
	UnReg_Id           = "unregnotify"
	ErrMsg_Id          = "errmsg"
	PingMsg_Id         = "pingmsg"
	BroadCastMsg_Id    = "broadcastmsg"
	RoleEnterNotify_Id = "roleenternotify"
	ReassignNotify_Id  = "reassignnotify"
)

func init() {
	protocols = make(map[string]proto_interface, 10)
	protocols[RegReq_Id] = &RegReqImp{}
	protocols[Pos_Id] = &PosNotifyImp{}
	protocols[UnReg_Id] = &UnRegNotifyImp{}
}

func GenRegResp(roomId string, roles []*gen.RoleInfo, rolesPos []*gen.RolePos) []byte {
	fb := flatbuffers.NewBuilder(0)
	role_offsets := make([]flatbuffers.UOffsetT, 0, len(roles))
	for i := 0; i < len(roles); i++ {
		role_offsets = append(role_offsets, Gen_RoleInfo(fb, roles[i], rolesPos[i]))
	}

	gen.RegRespStartRolesVector(fb, len(roles))
	for i := len(roles) - 1; i >= 0; i-- {
		fb.PrependUOffsetT(role_offsets[i])
	}
	roles_offset := fb.EndVector(len(roles))
	room_id_of := fb.CreateString(roomId)
	gen.RegRespStart(fb)
	gen.RegRespAddRoomId(fb, room_id_of)
	gen.RegRespAddRoles(fb, roles_offset)
	fb.Finish(gen.RegRespEnd(fb))
	return GenProto(RegResp_Id, fb.Bytes[fb.Head():])
}

func Gen_RoleEnterNotify(role *gen.RoleInfo, rolePos *gen.RolePos) []byte {
	fb := flatbuffers.NewBuilder(0)
	role_of := Gen_RoleInfo(fb, role, rolePos)

	gen.RoleEnterNotifyStart(fb)
	gen.RoleEnterNotifyAddRole(fb, role_of)
	fb.Finish(gen.RoleEnterNotifyEnd(fb))
	return GenProto(RoleEnterNotify_Id, fb.Bytes[fb.Head():])
}

func Gen_PosNotify(data *gen.PosNotify) []byte {
	fbs := flatbuffers.NewBuilder(0)

	of1 := fbs.CreateString(string(data.AccountId()))
	of2 := fbs.CreateString(string(data.Name()))
	rolePos := data.Pos(nil)
	gen.RolePosStartPosVector(fbs, rolePos.PosLength())
	for i := rolePos.PosLength() - 1; i >= 0; i-- {
		fbs.PrependFloat32(rolePos.Pos(i))
	}
	pos_offset := fbs.EndVector(rolePos.PosLength())
	gen.RolePosStart(fbs)
	gen.RolePosAddPos(fbs, pos_offset)
	gen.RolePosAddRotate(fbs, rolePos.Rotate())
	pos_r_offset := gen.RolePosEnd(fbs)

	gen.PosNotifyStart(fbs)
	gen.PosNotifyAddAccountId(fbs, of1)
	gen.PosNotifyAddName(fbs, of2)
	gen.PosNotifyAddPos(fbs, pos_r_offset)
	fbs.Finish(gen.PosNotifyEnd(fbs))
	return GenProto(Pos_Id, fbs.Bytes[fbs.Head():])
}

func Gen_UnRegNotify(accountId, name string) []byte {
	fbs := flatbuffers.NewBuilder(0)

	of1 := fbs.CreateString(accountId)
	of2 := fbs.CreateString(name)
	gen.UnRegNotifyStart(fbs)
	gen.UnRegNotifyAddAccountId(fbs, of1)
	gen.UnRegNotifyAddName(fbs, of2)
	fbs.Finish(gen.UnRegNotifyEnd(fbs))
	return GenProto(UnReg_Id, fbs.Bytes[fbs.Head():])
}

func Gen_ReassignNotify() []byte {
	fbs := flatbuffers.NewBuilder(0)
	gen.ReassignNotifyStart(fbs)
	fbs.Finish(gen.ReassignNotifyEnd(fbs))
	return GenProto(ReassignNotify_Id, fbs.Bytes[fbs.Head():])
}

func Gen_SysNotice(typ, msg string) []byte {
	fbs := flatbuffers.NewBuilder(0)
	oftyp := fbs.CreateString(typ)
	ofmsg := fbs.CreateString(msg)

	gen.BroadCastMsgStart(fbs)
	gen.BroadCastMsgAddTyp(fbs, oftyp)
	gen.BroadCastMsgAddMsg(fbs, ofmsg)
	fbs.Finish(gen.BroadCastMsgEnd(fbs))
	return GenProto(BroadCastMsg_Id, fbs.Bytes[fbs.Head():])
}

func Gen_PingMsg(msg string) []byte {
	fbs := flatbuffers.NewBuilder(0)
	ofst := fbs.CreateString(msg)

	gen.PingMsgStart(fbs)
	gen.PingMsgAddPing(fbs, ofst)
	fbs.Finish(gen.PingMsgEnd(fbs))
	return GenProto(PingMsg_Id, fbs.Bytes[fbs.Head():])
}

func Gen_ErrMsg(msg string) []byte {
	fbs := flatbuffers.NewBuilder(0)
	ofst := fbs.CreateString(msg)

	gen.ErrMsgStart(fbs)
	gen.ErrMsgAddErr(fbs, ofst)
	fbs.Finish(gen.ErrMsgEnd(fbs))
	return GenProto(ErrMsg_Id, fbs.Bytes[fbs.Head():])
}

func Gen_RoleInfo(fb *flatbuffers.Builder, role *gen.RoleInfo, rolePos *gen.RolePos) flatbuffers.UOffsetT {
	// equip
	equips_offset := make([]flatbuffers.UOffsetT, 0, role.EquipsLength())
	for i := 0; i < role.EquipsLength(); i++ {
		equips_offset = append(equips_offset, fb.CreateString(string(role.Equips(i))))
	}
	gen.RoleInfoStartEquipsVector(fb, role.EquipsLength())
	for i := role.EquipsLength() - 1; i >= 0; i-- {
		fb.PrependUOffsetT(equips_offset[i])
	}
	equip_offset := fb.EndVector(role.EquipsLength())
	// pos
	gen.RolePosStartPosVector(fb, rolePos.PosLength())
	for i := rolePos.PosLength() - 1; i >= 0; i-- {
		fb.PrependFloat32(rolePos.Pos(i))
	}
	pos_offset := fb.EndVector(rolePos.PosLength())
	gen.RolePosStart(fb)
	gen.RolePosAddPos(fb, pos_offset)
	gen.RolePosAddRotate(fb, rolePos.Rotate())
	pos_r_offset := gen.RolePosEnd(fb)

	of1 := fb.CreateString(string(role.AccountId()))
	of2 := fb.CreateString(string(role.Name()))
	of3 := fb.CreateString(string(role.Guild()))
	of4 := fb.CreateString(string(role.Title()))
	gen.RoleInfoStart(fb)
	gen.RoleInfoAddAccountId(fb, of1)
	gen.RoleInfoAddName(fb, of2)
	gen.RoleInfoAddLevel(fb, role.Level())
	gen.RoleInfoAddVip(fb, role.Vip())
	gen.RoleInfoAddGs(fb, role.Gs())
	gen.RoleInfoAddRole(fb, role.Role())
	gen.RoleInfoAddEquips(fb, equip_offset)
	gen.RoleInfoAddGuild(fb, of3)
	gen.RoleInfoAddTitle(fb, of4)
	gen.RoleInfoAddTitleTimeout(fb, role.TitleTimeout())
	gen.RoleInfoAddCharStarLevel(fb, role.CharStarLevel())
	gen.RoleInfoAddWeaponStarLevel(fb, role.WeaponStarLevel())
	gen.RoleInfoAddRolePos(fb, pos_r_offset)
	gen.RoleInfoAddSwing(fb, role.Swing())
	return gen.RoleInfoEnd(fb)
}

// 发送协议时调用
func GenProto(proto_id string, bs []byte) []byte {
	b := flatbuffers.NewBuilder(0)
	b.Reset()

	id_pos := b.CreateString(proto_id)
	pos := b.CreateByteVector(bs)

	gen.ProtoWrapStart(b)
	gen.ProtoWrapAddId(b, id_pos)
	gen.ProtoWrapAddSubProto(b, pos)
	position := gen.ProtoWrapEnd(b)

	b.Finish(position)

	return b.Bytes[b.Head():]
}

// 接受协议时调用
func RecProto(player *Player, bs []byte) {
	id, b := ReadProtoWrap(bs)
	temp, ok := protocols[id]
	if !ok {
		logs.Error("rec protocol not register: %s", id)
		return
	}
	proto := temp.Clone()
	proto.decode(b)
	proto.process(player)
}

func ReadProtoWrap(buf []byte) (id string, b []byte) {
	p := gen.GetRootAsProtoWrap(buf, 0)
	b = make([]byte, p.SubProtoLength())
	for i := 0; i < p.SubProtoLength(); i++ {
		b[i] = byte(p.SubProto(i))
	}
	return string(p.Id()), b
}

type proto_interface interface {
	Clone() proto_interface
	decode(bs []byte)
	process(player *Player)
}

// RegRequest
type RegReqImp struct {
	data *gen.RegReq
}

func (req *RegReqImp) Clone() proto_interface {
	return &RegReqImp{}
}

func (req *RegReqImp) decode(bs []byte) {
	req.data = gen.GetRootAsRegReq(bs, 0)
}

func (req *RegReqImp) process(player *Player) {
	Rec_RegReq(player, req.data)
}

// pos
type PosNotifyImp struct {
	data *gen.PosNotify
}

func (req *PosNotifyImp) Clone() proto_interface {
	return &PosNotifyImp{}
}

func (req *PosNotifyImp) decode(bs []byte) {
	req.data = gen.GetRootAsPosNotify(bs, 0)
}

func (req *PosNotifyImp) process(player *Player) {
	Rec_PosNotify(player, req.data)
}

// unreg
type UnRegNotifyImp struct {
	data *gen.UnRegNotify
}

func (req *UnRegNotifyImp) Clone() proto_interface {
	return &UnRegNotifyImp{}
}

func (req *UnRegNotifyImp) decode(bs []byte) {
	req.data = gen.GetRootAsUnRegNotify(bs, 0)
}

func (req *UnRegNotifyImp) process(player *Player) {
	Rec_UnRegNotify(player, string(req.data.AccountId()), string(req.data.Name()))
}
