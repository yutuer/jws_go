package room

import "sort"

const (
	roomNumAllocMax = 128
)

//TODO by YZH FenghuoRoom room_num模式的改造
func (r *module) allocRoomNum() int {
	if r.roomNums == nil || len(r.roomNums) == 0 {
		max := roomNumAllocMax + r.roomNumAllocCurrMax
		r.roomNums = make([]int, roomNumAllocMax, roomNumAllocMax)
		for i := 0; i < roomNumAllocMax; i++ {
			r.roomNums[i] = max - i
		}
		r.roomNumAllocCurrMax = max
	}

	resNum := r.roomNums[len(r.roomNums)-1]
	r.roomNums = r.roomNums[:len(r.roomNums)-1]

	r.addToRoomNumArray(resNum)
	return resNum
}

func (r *module) deallocRoomNum(num int) {
	if num <= 0 {
		return
	}

	r.roomNums = append(r.roomNums, num)

	if len(r.roomNums) > roomNumAllocMax*8 {
		r.roomNums = r.roomNums[:roomNumAllocMax]
	}
	r.delFromRoomNumArray(num)
}

func (r *module) addToRoomNumArray(num int) {
	r.roomNumArray = append(r.roomNumArray, num)
	sort.Ints(r.roomNumArray[:])
}

func (r *module) delFromRoomNumArray(num int) {
	arrayLen := len(r.roomNumArray)
	if arrayLen <= 1 {
		r.roomNumArray = r.roomNumArray[0:0]
		return
	}
	idx := sort.SearchInts(r.roomNumArray, num)
	if idx >= 0 && idx < arrayLen {
		if idx != arrayLen-1 {
			r.roomNumArray[idx] = r.roomNumArray[arrayLen-1]
		}
		r.roomNumArray = r.roomNumArray[:arrayLen-1]
		sort.Ints(r.roomNumArray)
	}
}
