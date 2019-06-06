// Code generated by protoc-gen-go.
// source: ProtobufGen_comic.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type COMIC struct {
	// * 漫画ID
	ComicID *string `protobuf:"bytes,1,req,def=" json:"ComicID,omitempty"`
	// * 关卡ID
	PassID *string `protobuf:"bytes,2,req,def=" json:"PassID,omitempty"`
	// * 触发时机，0=进入场景  1=战斗结束
	TriggerTime *int32 `protobuf:"varint,3,req,def=0" json:"TriggerTime,omitempty"`
	// * 是否重复，1=每次进场景都播放，0=只有通关前播放
	Repeatable *int32 `protobuf:"varint,4,req,def=0" json:"Repeatable,omitempty"`
	// * 动画Prefab
	Prefab *string `protobuf:"bytes,5,req,def=" json:"Prefab,omitempty"`
	// * 播放的漫画图片
	Picture *string `protobuf:"bytes,6,req,def=" json:"Picture,omitempty"`
	// * 声音文件
	Voice *string `protobuf:"bytes,7,opt,def=" json:"Voice,omitempty"`
	// * 字幕组ID
	SubtitleGroupID  *string `protobuf:"bytes,8,opt,def=" json:"SubtitleGroupID,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *COMIC) Reset()         { *m = COMIC{} }
func (m *COMIC) String() string { return proto.CompactTextString(m) }
func (*COMIC) ProtoMessage()    {}

const Default_COMIC_TriggerTime int32 = 0
const Default_COMIC_Repeatable int32 = 0

func (m *COMIC) GetComicID() string {
	if m != nil && m.ComicID != nil {
		return *m.ComicID
	}
	return ""
}

func (m *COMIC) GetPassID() string {
	if m != nil && m.PassID != nil {
		return *m.PassID
	}
	return ""
}

func (m *COMIC) GetTriggerTime() int32 {
	if m != nil && m.TriggerTime != nil {
		return *m.TriggerTime
	}
	return Default_COMIC_TriggerTime
}

func (m *COMIC) GetRepeatable() int32 {
	if m != nil && m.Repeatable != nil {
		return *m.Repeatable
	}
	return Default_COMIC_Repeatable
}

func (m *COMIC) GetPrefab() string {
	if m != nil && m.Prefab != nil {
		return *m.Prefab
	}
	return ""
}

func (m *COMIC) GetPicture() string {
	if m != nil && m.Picture != nil {
		return *m.Picture
	}
	return ""
}

func (m *COMIC) GetVoice() string {
	if m != nil && m.Voice != nil {
		return *m.Voice
	}
	return ""
}

func (m *COMIC) GetSubtitleGroupID() string {
	if m != nil && m.SubtitleGroupID != nil {
		return *m.SubtitleGroupID
	}
	return ""
}

type COMIC_ARRAY struct {
	Items            []*COMIC `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *COMIC_ARRAY) Reset()         { *m = COMIC_ARRAY{} }
func (m *COMIC_ARRAY) String() string { return proto.CompactTextString(m) }
func (*COMIC_ARRAY) ProtoMessage()    {}

func (m *COMIC_ARRAY) GetItems() []*COMIC {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
