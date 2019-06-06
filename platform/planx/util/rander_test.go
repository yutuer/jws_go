package util

/*
import (
	"math/rand"
	"testing"
)

func testRander(r *RandSet, rr *rand.Rand, max int) map[string]int {
	output := map[string]int{}
	for i := 0; i < max; i++ {
		res := r.Rand(rr)
		output[res] = output[res] + 1
	}
	return output
}

func testRanders(t *testing.T, r *RandSetMap, rr *rand.Rand, k string, max int) map[string]int {
	output := map[string]int{}
	for i := 0; i < max; i++ {
		res, ok := r.Rand(k, rr)
		if !ok {
			t.Errorf("rand err!")
		}
		output[res] = output[res] + 1
	}
	return output
}

// 小概率测试，测试空和H出现的频率， H出现概率和空的概率均为万分之一，
func TestRandSetMin(t *testing.T) {
	Rng := Kiss64Rng{}
	Rng.Seed(5927)

	r := rand.New(&Rng)

	rander := &RandSet{}
	rander.Init(10000, 10)
	rander.Add("A", 1000)
	rander.Add("B", 2000)
	rander.Add("C", 1000)
	rander.Add("D", 100)
	rander.Add("E", 2400)
	rander.Add("F", 1600)
	rander.Add("G", 900)
	rander.Add("G", 998)
	rander.Add("H", 1)

	o1000 := testRander(rander, r, 1000)
	o10000 := testRander(rander, r, 10000)
	o100000 := testRander(rander, r, 100000)
	o1000000 := testRander(rander, r, 1000000)

	t.Logf("rander : %v\n", rander)

	t.Logf("o1000 : %v\n", o1000)
	t.Logf("o10000 : %v\n", o10000)
	t.Logf("o100000 : %v\n", o100000)
	t.Logf("o1000000 : %v\n", o1000000)

}

// 常规频率测试
func TestRandSet(t *testing.T) {
	Rng := Kiss64Rng{}
	Rng.Seed(5927)

	r := rand.New(&Rng)
	rander := &RandSet{}
	rander.Init(10000, 10)
	rander.Add("A", 1000)
	rander.Add("B", 2000)
	rander.Add("C", 1000)
	rander.Add("D", 100)
	rander.Add("E", 2400)
	rander.Add("F", 1600)
	rander.Add("G", 900)
	rander.Add("H", 100)
	// "" -> 900

	o1000 := testRander(rander, r, 1000)
	o10000 := testRander(rander, r, 10000)
	o100000 := testRander(rander, r, 100000)
	o1000000 := testRander(rander, r, 1000000)

	t.Logf("rander : %v\n", rander)

	t.Logf("o1000 : %v\n", o1000)
	t.Logf("o10000 : %v\n", o10000)
	t.Logf("o100000 : %v\n", o100000)
	t.Logf("o1000000 : %v\n", o1000000)

}

// 随机池集合测试，测试概率结果是否正确
func TestRandSetMapOne(t *testing.T) {
	Rng := Kiss64Rng{}
	Rng.Seed(5927)

	r := rand.New(&Rng)

	rm := &RandSetMap{}
	rm.Init(10000, 10, 5)
	r1 := rm.NewRandSet()
	r1.Add("R1A", 5000)
	r1.Add("R1B", 2000)
	rm.AddRandSet("R1", r1)

	t.Logf("rm : %v", *rm)
	rm1res := testRanders(t, rm, r, "R1", 1000000)
	t.Logf("rm1res : %v\n", rm1res)
}

// 随机池集合测试，测试内存是否正确
func TestRandSetMap(t *testing.T) {
	rm := &RandSetMap{}
	rm.Init(10000, 4, 5)
	r1 := rm.NewRandSet()
	r1.Add("R1A", 1000)
	r1.Add("R1B", 1000)
	r1.Add("R1C", 2000)
	r1.Add("R1D", 3000)
	rm.AddRandSet("R1", r1)

	r2 := rm.NewRandSet()
	r2.Add("R2A", 5000)
	r2.Add("R2B", 1000)
	r2.Add("R2C", 3000)
	r2.Add("R2D", 1000)
	rm.AddRandSet("R2", r2)

	r3 := rm.NewRandSet()
	r3.Add("R3A", 5000)
	r3.Add("R3B", 2000)
	rm.AddRandSet("R3", r3)

	r4 := rm.NewRandSet()
	r4.Add("R4A", 5000)
	r4.Add("R4B", 2000)
	rm.AddRandSet("R4", r4)

	r5 := rm.NewRandSet()
	r5.Add("R5A", 1000)
	r5.Add("R5B", 1000)
	r5.Add("R5C", 1000)
	r5.Add("R5D", 1000)
	rm.AddRandSet("R5", r5)

	t.Logf("rm : %v", *rm)

	Rng := Kiss64Rng{}
	Rng.Seed(5927)

	r := rand.New(&Rng)

	rm1res := testRanders(t, rm, r, "R1", 1000000)
	t.Logf("rm1res : %v\n", rm1res)

	rm2res := testRanders(t, rm, r, "R2", 1000000)
	t.Logf("rm2res : %v\n", rm2res)

	rm3res := testRanders(t, rm, r, "R3", 1000000)
	t.Logf("rm3res : %v\n", rm3res)

	rm4res := testRanders(t, rm, r, "R4", 1000000)
	t.Logf("rm4res : %v\n", rm4res)

	rm5res := testRanders(t, rm, r, "R5", 1000000)
	t.Logf("rm5res : %v\n", rm5res)
}
*/
