package store

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShop_ShopRefresh(t *testing.T) {
	shop := make(chan *Shop)
	defer close(shop)

	now := time.Now().UnixNano()

	go func() {
		shop <- &Shop{ShopTyp: 0,
			Goods: []Good{Good{GoodId: "JD_1_12", UseTimes: 5}},
		}
		shop <- &Shop{ShopTyp: 3,
			Goods: []Good{Good{GoodId: "JD_1_12", UseTimes: 5}},
		}
		shop <- &Shop{ShopTyp: 0,
			Goods:    []Good{Good{GoodId: "JD_233_122", UseTimes: 5}},
			LastTime: now,
		}
		shop <- &Shop{Goods: make([]Good, 0, 10)}
		shop <- &Shop{Goods: make([]Good, 0, 10), LastTime: now}
	}()

	// 正常刷新
	s1 := <-shop
	assert.True(t, s1.ShopRefresh(now))
	assert.Equal(t, s1.Goods[0].UseTimes, 0)

	// 正常不刷新
	s1.Goods[0].UseTimes = 15
	assert.False(t, s1.ShopRefresh(now))
	assert.Equal(t, s1.Goods[0].UseTimes, 15)

	// Shop type不存在，只判断时间
	s2 := <-shop
	assert.True(t, s2.ShopRefresh(now))
	assert.Equal(t, s2.Goods[0].UseTimes, 5)

	// Good config不存在，只判断时间
	s3 := <-shop
	assert.False(t, s3.ShopRefresh(now))
	assert.Equal(t, s3.Goods[0].UseTimes, 5)

	// 什么都没有
	assert.True(t, (<-shop).ShopRefresh(now))

	// 只有时间
	assert.False(t, (<-shop).ShopRefresh(now))
}

func TestShop_AddGoodUseTimes(t *testing.T) {
	shop := &Shop{ShopTyp: 0}

	// 新增
	assert.True(t, shop.AddGoodUseTimes("JD_1_12", 1))
	assert.Equal(t, shop.Goods[0].UseTimes, 1)

	// 累加
	assert.True(t, shop.AddGoodUseTimes("JD_1_12", 2))
	assert.Equal(t, shop.Goods[0].UseTimes, 3)

	// 超过10个
	for i := 1; i < 10; i++ {
		s := "JD_1_" + strconv.Itoa(i)
		shop.AddGoodUseTimes(s, 5)
	}
	assert.True(t, shop.AddGoodUseTimes("JD_1_14", 23))
	assert.Equal(t, len(shop.Goods), 11)

	// 不存在
	assert.False(t, shop.AddGoodUseTimes("JD_12_123", 2))
}

func TestShop_GetGoodUseTimes(t *testing.T) {
	shop := &Shop{ShopTyp: 0}
	shop.AddGoodUseTimes("JD_1_12", 4)

	// 没有
	assert.Equal(t, shop.GetGoodUseTimes("JD_12_123"), 0)

	// 有
	assert.Equal(t, shop.GetGoodUseTimes("JD_1_12"), 4)
}
