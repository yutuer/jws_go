package logics

import "testing"

var testsData = []struct {
	src  []int64
	dest []int64
}{
	{
		[]int64{1, 2, 3, 1, 2, 3, 1, 2, 3},
		[]int64{1, -1, -1, 2, -1, -1, 3, -1, -1},
	},
	{
		[]int64{1, 2, 3, 3, 2, 1, 2, 1, 3},
		[]int64{1, -1, -1, 2, -1, -1, 3, -1, -1},
	},
	{
		[]int64{1, 2, -1, 1, 2, -1, 1, 2, -1},
		[]int64{1, -1, -1, 2, -1, -1, 1, -1, -1},
	},
	{
		[]int64{1, -1, -1, 1, -1, -1, 1, -1, -1},
		[]int64{1, -1, -1, 1, -1, -1, 1, -1, -1},
	},
	{
		[]int64{1, 2, 3, 4, 5, 6, 7, 8, 9},
		[]int64{1, 2, 3, 4, 5, 6, 7, 8, 9},
	},
	{
		nil,
		nil,
	},
	{
		[]int64{1, -1, 3, -1, 5, 6, 7, -1, 9},
		[]int64{1, -1, 3, -1, 5, 6, 7, -1, 9},
	},
}

func TestCheckAndChangeFormation(t *testing.T) {
	for _, data := range testsData {
		case1 := CheckAndChangeFormation(data.src)

		for i := range case1 {
			if case1[i] != data.dest[i] {
				if t.Failed() {
					t.Error("")
					t.Fail()
				}
			}
		}
	}
}
