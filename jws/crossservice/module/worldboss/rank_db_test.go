package worldboss

import (
	"strings"
	"testing"

	"vcs.taiyouxi.net/jws/crossservice/util/csdb"
	"vcs.taiyouxi.net/platform/planx/util/uuid"

	"github.com/stretchr/testify/assert"
)

// 使用全局rankDB对象
var rdb *RankDB = newRankDB(debugGetResource())

// debugDelDamageRank 删除全局db中对应的key
func debugDelDamageRank(tag string) {
	rankKey := rdb.tableNameRank(tag)
	conn := csdb.GetDBConn(rdb.group)
	defer conn.Close()
	conn.Do("DEL", rankKey)
}

// debugGenRankElems 生成n个元素，Acid随机，Damage递减，Pos递增
func debugGenRankElems(n int) []DamageRankElem {
	dRankElems := make([]DamageRankElem, n)
	for i := 0; i < n; i++ {
		dRankElems[i] = DamageRankElem{
			Acid:   uuid.NewV4().String(),
			Sid:    testSid,
			Damage: uint64(n - i),
			Pos:    uint32(i + 1)}
	}

	return dRankElems
}

func TestNewRankDB(t *testing.T) {
	newRdb := newRankDB(debugGetResource())

	assert.Equal(t, testGroupId, newRdb.group)
}

func TestPushRankMember(t *testing.T) {
	damageRankElems := debugGenRankElems(2)
	tag := "rank_db_test_push"
	tableName := rdb.tableNameRank(tag)

	// 正常push
	t.Run("Normal Push", func(t *testing.T) {
		err0 := rdb.pushRankMember(damageRankElems, tag)
		assert.Nil(t, err0)

		conn := csdb.GetDBConn(rdb.group)
		r, _ := conn.Do("HLEN", tableName)
		assert.Equal(t, int64(2), r)

		code, _ := conn.Do("HEXISTS", tableName, damageRankElems[0].Acid)
		assert.Equal(t, int64(1), code)
	})

	// 空RankElems
	t.Run("Push Empty", func(t *testing.T) {
		emptyDamageRankElems := []DamageRankElem{}
		err := rdb.pushRankMember(emptyDamageRankElems, tag)
		assert.NotNil(t, err)
	})

	// Redis不可用
	t.Run("Redis Unavailable", func(t *testing.T) {
		unavailableDB := newRankDB(newResources(unavailableGroupId, &WorldBoss{}))
		err1 := unavailableDB.pushRankMember(damageRankElems, tag)
		assert.NotNil(t, err1)
		assert.True(t, strings.Contains(err1.Error(), "refused"))
	})

	// groupId对应的Redis不存在
	t.Run("groupID invalid", func(t *testing.T) {
		notExistDB := newRankDB(newResources(notExistGroupId, &WorldBoss{}))
		err2 := notExistDB.pushRankMember(damageRankElems, tag)
		assert.NotNil(t, err2)
		assert.Contains(t, err2.Error(), "GetDBConn")
	})

	defer debugDelDamageRank(tag)
}

func TestRemoveRank(t *testing.T) {
	dRankElems := debugGenRankElems(1)
	tag := "rank_db_test_remove"
	tableName := rdb.tableNameRank(tag)
	rdb.pushRankMember(dRankElems, tag)

	// Tag错误 -- 并不会出现错误，也不理会Redis的反馈
	t.Run("Incorrect Tag", func(t *testing.T) {
		tag := "No such tag"
		err := rdb.removeRank(tag)
		assert.Nil(t, err)

		// 不会删除数据
		conn := csdb.GetDBConn(rdb.group)
		code, _ := conn.Do("HEXISTS", tableName, dRankElems[0].Acid)
		assert.Equal(t, int64(1), code)
	})

	// 正常Remove
	t.Run("Normal Remove", func(t *testing.T) {
		err := rdb.removeRank(tag)
		assert.Nil(t, err)

		// 正确删除数据
		conn := csdb.GetDBConn(rdb.group)
		code, _ := conn.Do("HEXISTS", tableName, dRankElems[0].Acid)
		assert.Equal(t, int64(0), code)
	})

	// Redis不可用
	t.Run("Redis Unavailable", func(t *testing.T) {
		unavailableDB := newRankDB(newResources(unavailableGroupId, &WorldBoss{}))
		err1 := unavailableDB.pushRankMember(dRankElems, tag)
		assert.NotNil(t, err1)
		assert.True(t, strings.Contains(err1.Error(), "refused"))
	})

	// groupId对应的Redis不存在
	t.Run("groupID invalid", func(t *testing.T) {
		notExistDB := newRankDB(newResources(notExistGroupId, &WorldBoss{}))
		err2 := notExistDB.pushRankMember(dRankElems, tag)
		assert.NotNil(t, err2)
		assert.Contains(t, err2.Error(), "GetDBConn")
	})
}

func TestGetAllRankMember(t *testing.T) {
	damageRankElems := debugGenRankElems(1000)
	tag := "rank_db_test_get_all"
	rdb.pushRankMember(damageRankElems, tag)

	t.Run("Correct", func(t *testing.T) {
		data, err := rdb.getAllRankMember(tag)
		assert.Nil(t, err)
		assert.IsType(t, make(map[string]DamageRankElem), data)
		assert.Equal(t, 1000, len(data))

		_, ok := data[damageRankElems[0].Acid]
		assert.True(t, ok)
	})

	t.Run("Invalid tag", func(t *testing.T) {
		tag := "Invalid tag"
		data, err := rdb.getAllRankMember(tag)
		assert.Nil(t, err)
		assert.Empty(t, data)
	})

	// Redis不可用
	t.Run("Redis Unavailable", func(t *testing.T) {
		unavailableDB := newRankDB(newResources(unavailableGroupId, &WorldBoss{}))
		_, err1 := unavailableDB.getAllRankMember(tag)
		assert.NotNil(t, err1)
		assert.True(t, strings.Contains(err1.Error(), "refused"))
	})

	// groupId对应的Redis不存在
	t.Run("groupID invalid", func(t *testing.T) {
		notExistDB := newRankDB(newResources(notExistGroupId, &WorldBoss{}))
		_, err2 := notExistDB.getAllRankMember(tag)
		assert.NotNil(t, err2)
		assert.Contains(t, err2.Error(), "GetDBConn")
	})

	defer rdb.removeRank(tag)
}
