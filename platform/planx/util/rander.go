package util

import (
	"math/rand"
	"sort"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RandSet struct {
	power_max uint32
	res       []string
	power     []uint32
	threshold []float64
}

/*
	随机选择类
	根据一个随机池随机出一个结果
	随机池由 {结果, 权重}组成，
	权重有一个最大值，如果大于所有权重之和，表示有一定概率出现空结果

	存储时将所有权重处理在[0, 1)中间
	如

	- "A" 2000
	- "B" 3000
	- "C" 4000

	权重最大值是 10000， 则处理成：

	+---+----+---+---+
	0  0.2  0.5 0.9  1
	  A    B   C   ""

	随机时生成一个[0, 1)的浮点数
	看在哪个区间中

	存储时 res数组存储结果， threshold存储每个区间的最大阙值

	需要注意精度 如果刚好没有空得情况下，最后一个必然是1
	但是计算有误差会出现0.99999999999的情况
	这里强制最小的概率为百万分之一
*/

var MIN_POWER_FLOAT float64 = 0.000001

func (r *RandSet) Init(size int) {
	r.res = make([]string, 0, size)
	r.threshold = make([]float64, 0, size)
	r.power = make([]uint32, 0, size)
}
func (r *RandSet) Add(res string, power uint32) bool {
	if power > 0 {
		r.res = append(r.res, res)
		r.power = append(r.power, power)
	}
	return true
}
func (r *RandSet) addImp(res string, power uint32) bool {
	if r.power_max == 0 {
		return false
	}
	last_threshold := float64(0.0)
	if len(r.threshold) > 0 {
		last_threshold = r.threshold[len(r.threshold)-1]
	}
	new_threshold := last_threshold + (float64(power) / float64(r.power_max))
	if new_threshold >= float64(1.0)+MIN_POWER_FLOAT {
		return false
	}

	if (float64(1.0) - new_threshold) < MIN_POWER_FLOAT {
		new_threshold = float64(1.0)
	}

	r.threshold = append(r.threshold, new_threshold)
	return true
}

func (r *RandSet) getPowerMax() uint32 {
	var p uint32
	for _, v := range r.power {
		p += v
	}
	return p
}

func (r *RandSet) Rand(rd *rand.Rand) string {
	var s float64
	if rd != nil {
		s = rd.Float64()
	} else {
		s = rand.Float64()
	}
	res_idx := sort.Search(len(r.threshold),
		func(i int) bool { return r.threshold[i] > s })

	if res_idx >= len(r.res) {
		return ""
	}
	return r.res[res_idx]
}

func (r *RandSet) Make() bool {
	r.power_max = r.getPowerMax()
	if r.power_max == 0 {
		return false
	}

	re := true

	for idx, n := range r.res {
		re = re && r.addImp(n, r.power[idx])
	}

	if len(r.res) == 0 {
		return false
	}

	return re
}

type RandIntSet struct {
	power_max uint32
	res       []int
	power     []uint32
	threshold []float64
}

func (r *RandIntSet) Init(size int) {
	r.res = make([]int, 0, size)
	r.power = make([]uint32, 0, size)
	r.threshold = make([]float64, 0, size)
}

func (r *RandIntSet) getPowerMax() uint32 {
	var p uint32
	for _, v := range r.power {
		p += v
	}
	return p
}

func (r *RandIntSet) Add(res int, power uint32) bool {
	if power > 0 {
		r.res = append(r.res, res)
		r.power = append(r.power, power)
	}
	return true
}

func (r *RandIntSet) addImp(res int, power uint32) bool {
	if r.power_max == 0 {
		return false
	}
	last_threshold := float64(0.0)
	if len(r.threshold) > 0 {
		last_threshold = r.threshold[len(r.threshold)-1]
	}
	new_threshold := last_threshold + (float64(power) / float64(r.power_max))
	if new_threshold >= float64(1.0)+MIN_POWER_FLOAT {
		return false
	}

	if (float64(1.0) - new_threshold) < MIN_POWER_FLOAT {
		new_threshold = float64(1.0)
	}

	r.threshold = append(r.threshold, new_threshold)
	return true
}

func (r *RandIntSet) Rand(rd *rand.Rand) int {
	s := rd.Float64()
	logs.Trace("Rand Res %v", s)
	res_idx := sort.Search(len(r.threshold),
		func(i int) bool { return r.threshold[i] > s })

	if res_idx >= len(r.res) {
		logs.Debug("rand failed %v", r)
		return -1
	}
	return r.res[res_idx]
}

func (r *RandIntSet) Make() bool {
	r.power_max = r.getPowerMax()
	if r.power_max == 0 {
		return false
	}

	re := true

	for idx, n := range r.res {
		re = re && r.addImp(n, r.power[idx])
	}

	if len(r.res) == 0 {
		return false
	}

	return re
}

type RandUInt32Set struct {
	power_max uint32
	res       []uint32
	threshold []float64
	power     []uint32
}

func (r *RandUInt32Set) Init(size int) {
	r.res = make([]uint32, 0, size)
	r.power = make([]uint32, 0, size)
	r.threshold = make([]float64, 0, size)
}

func (r *RandUInt32Set) Add(res uint32, power uint32) bool {
	if power > 0 {
		r.res = append(r.res, res)
		r.power = append(r.power, power)
	}
	return true
}

func (r *RandUInt32Set) addImp(res uint32, power uint32) bool {
	if r.power_max == 0 {
		return false
	}

	last_threshold := float64(0.0)
	if len(r.threshold) > 0 {
		last_threshold = r.threshold[len(r.threshold)-1]
	}
	new_threshold := last_threshold + (float64(power) / float64(r.power_max))
	if new_threshold >= float64(1.0)+MIN_POWER_FLOAT {
		return false
	}

	if (float64(1.0) - new_threshold) < MIN_POWER_FLOAT {
		new_threshold = float64(1.0)
	}

	r.threshold = append(r.threshold, new_threshold)
	return true
}

func (r *RandUInt32Set) getPowerMax() uint32 {
	var p uint32
	for _, v := range r.power {
		p += v
	}
	return p
}

func (r *RandUInt32Set) Rand(rd *rand.Rand) uint32 {
	s := rd.Float64()
	logs.Trace("Rand Res %v", s)
	res_idx := sort.Search(len(r.threshold),
		func(i int) bool { return r.threshold[i] > s })

	if res_idx >= len(r.res) {
		return 0
	}
	return r.res[res_idx]
}

func (r *RandUInt32Set) Make() bool {
	r.power_max = r.getPowerMax()
	if r.power_max == 0 {
		return false
	}

	re := true

	for idx, n := range r.res {
		re = re && r.addImp(n, r.power[idx])
	}

	if len(r.res) == 0 {
		return false
	}

	return re
}

type RandSetMap struct {
	randers        map[string]RandSet
	power_max      uint32
	res_data       []string
	power          []uint32
	threshold_data []float64
	data_next      int
	pool_max_size  int
}

func (r *RandSetMap) Init(pool_max_size, list_size int) {
	r.randers = make(map[string]RandSet, list_size)
	data_size := pool_max_size * list_size
	r.res_data = make([]string, data_size, data_size)
	r.threshold_data = make([]float64, data_size, data_size)
	r.power = make([]uint32, data_size, data_size)
	r.data_next = 0
	r.pool_max_size = pool_max_size
}

func (r *RandSetMap) NewRandSet() *RandSet {
	data_next := r.data_next
	new_data_next := r.data_next + r.pool_max_size

	res := &RandSet{
		r.power_max,
		r.res_data[data_next:data_next:new_data_next],
		r.power[data_next:data_next:new_data_next],
		r.threshold_data[data_next:data_next:new_data_next],
	}
	r.data_next = new_data_next
	return res
}

func (r *RandSetMap) AddRandSet(k string, nr *RandSet) {
	if !nr.Make() {
		logs.Error("AddRandSet Err %s %v", k, *nr)
		return
	}
	r.randers[k] = *nr
}

func (r *RandSetMap) Rand(k string, rd *rand.Rand) (string, bool) {
	rander, ok := r.randers[k]
	if !ok {
		return "", false
	}

	return rander.Rand(rd), true
}

//
// 随机判断是否触发， 按照 n / N 的概率判定一个事件是否为真
//
func RandIfTrue(rd *rand.Rand, n, N int64) bool {
	if N <= 0 {
		return false
	}
	s := rd.Float64()
	return s < (float64(n) / float64(N))

}
