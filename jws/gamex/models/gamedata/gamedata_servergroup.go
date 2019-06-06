package gamedata

//
//import (
//	"fmt"
//
//	"github.com/golang/protobuf/proto"
//	"vcs.taiyouxi.net/jws/gamex/protogen"
//)
//
//var (
//	gdServerGroup         map[uint32]uint32   // shard->group
//	gdServerGroupActivity map[uint32][]uint32 // group->activityid
//	gdServerGroupSbatch   map[uint32]ProtobufGen.SEVERGROUP
//)
//
//func GetShardActivities(shard uint32) []uint32 {
//	groupId, ok := gdServerGroup[shard]
//	if !ok {
//		return []uint32{}
//	}
//	ret, ok := gdServerGroupActivity[groupId]
//	if !ok {
//		return []uint32{}
//	}
//	return ret
//}
//
//func GetShardGroup(shard uint32) uint32 {
//	return gdServerGroup[shard]
//}
//
//func mkServerGroupDatas(loadFunc func(dfilepath string, loadfunc func(string))) {
//	loadFunc("severgroup.data", loadServeGroup)
//	loadFunc("sgactivity.data", loadServeGroupActivity)
//}
//func GetSbatch(shard uint32) []ProtobufGen.SEVERGROUP {
//	return gdServerGroupSbatch[shard]
//}
//
//
//func loadServeGroup(filepath string) {
//	errcheck := func(err error) {
//		if err != nil {
//			panic(err)
//		}
//	}
//	buffer, err := loadBin(filepath)
//	errcheck(err)
//
//	ar := &ProtobufGen.SEVERGROUP_ARRAY{}
//	err = proto.Unmarshal(buffer, ar)
//	errcheck(err)
//
//	as := ar.GetItems()
//	gdServerGroup = make(map[uint32]uint32, len(as))
//	gdServerGroupSbatch = make(map[uint32]ProtobufGen.SEVERGROUP, len(as))
//	for _, r := range as {
//		gdServerGroup[r.GetSID()] = r.GetGroupID()
//		gdServerGroupSbatch[r.GroupID] = r
//	}
//}
//
//func loadServeGroupActivity(filepath string) {
//	errcheck := func(err error) {
//		if err != nil {
//			panic(err)
//		}
//	}
//	buffer, err := loadBin(filepath)
//	errcheck(err)
//
//	ar := &ProtobufGen.SGACTIVITY_ARRAY{}
//	err = proto.Unmarshal(buffer, ar)
//	errcheck(err)
//
//	as := ar.GetItems()
//	gdServerGroupActivity = make(map[uint32][]uint32, len(as))
//	for _, r := range as {
//		if r.GetHotActivityID() == nil {
//			panic(fmt.Errorf("ServeGroupActivity HotActivityID is nil, %d", r.GetGroupID()))
//		}
//		gdServerGroupActivity[r.GetGroupID()] = r.GetHotActivityID()
//	}
//}
