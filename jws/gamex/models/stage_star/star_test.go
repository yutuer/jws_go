package stage_star

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
是要数二进制里有多少1么……
*/

func GetStarCount2(n int32) int32 {
	var c int32
	for ; n != 0; c++ {
		n &= (n - 1)
	}
	return c
}

func TestGetStarCount(t *testing.T) {
	for i := int32(0); i < 8; i++ {
		assert.Equal(t, GetStarCount(i), GetStarCount2(i))
	}
}

// 取星数多的，如果一样，取后面的
func TestAddStar(t *testing.T) {
	assert.EqualValues(t, Star000, AddStar(Star000, Star000))
	assert.EqualValues(t, Star001, AddStar(Star100, Star001))
	assert.EqualValues(t, Star011, AddStar(Star011, Star100))
	assert.EqualValues(t, Star111, AddStar(Star111, Star010))
	assert.EqualValues(t, Star111, AddStar(Star111, Star111))
	assert.EqualValues(t, Star011, AddStar(Star110, Star011))
	assert.EqualValues(t, Star010, AddStar(Star010, Star000))
}

/*
func TestAdd(t *testing.T) {
	logs.Info("%b %v", c, AddStar(Star000, Star001))
	logs.Info("%b %v", AddStar(Star100, Star001), AddStar(Star100, Star001))
	logs.Info("%b %v", AddStar(Star111, Star001), AddStar(Star111, Star001))
	logs.Info("%b %v", AddStar(Star111, Star101), AddStar(Star111, Star101))
	logs.Info("%b %v", AddStar(Star110, Star011), AddStar(Star110, Star011))
	logs.Info("%b %v", AddStar(Star010, Star101), AddStar(Star010, Star101))
	logs.Info("%b %v", AddStar(int32(int64(2)), Star101), AddStar(Star010, Star101))
	logs.Info("%b %v", AddStar(int32(int64(2)), Star101), AddStar(Star010, Star101))
	logs.Info("%b %v", AddStar(int32(int64(2)), Star101), AddStar(Star010, Star101))
	logs.Flush()
}
*/
