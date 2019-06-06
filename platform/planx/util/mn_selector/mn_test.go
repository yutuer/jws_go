package mnSelector

import (
	"math/rand"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func initRander() *rand.Rand {
	return rand.New(rand.NewSource(44332))
}

func TestMN(t *testing.T) {
	rd := initRander()

	Convey("Base", t, func() {
		const (
			C int = 1
			S int = 10
		)
		mn := MNSelectorState{}
		mn.Init(int64(C), int64(S))
		var seletedNum int
		for i := 0; i < S; i++ {
			s := mn.Selector(rd)
			if s {
				seletedNum++
			}
		}
		So(seletedNum, ShouldEqual, C)
	})

	Convey("Rand", t, func() {
		const (
			C int = 1
			S int = 10
			M int = 50000
		)
		mn := MNSelectorState{}
		mn.Init(int64(C), int64(S))

		seleted := make([]int, S, S)
		for j := 0; j < M; j++ {
			for i := 0; i < S; i++ {
				s := mn.Selector(rd)
				if s {
					seleted[i]++
				}
			}
		}
		Println("Over : ", seleted)
	})

	Convey("SelectM", t, func() {
		const (
			C int = 5
			S int = 10
		)
		mn := MNSelectorState{}
		mn.Init(int64(C), int64(S))
		var seletedNum int
		for i := 0; i < S; i++ {
			s := mn.Selector(rd)
			if s {
				seletedNum++
			}
		}
		So(seletedNum, ShouldEqual, C)
	})

	Convey("SelectM10", t, func() {
		const (
			C int = 10
			S int = 10
		)
		mn := MNSelectorState{}
		mn.Init(int64(C), int64(S))
		var seletedNum int
		for i := 0; i < S; i++ {
			s := mn.Selector(rd)
			if s {
				seletedNum++
			}
		}
		So(seletedNum, ShouldEqual, C)
	})

	Convey("MNChange", t, func() {
		const (
			C int = 1
			S int = 10
		)
		mn := MNSelectorState{}
		mn.Init(int64(C), int64(S))
		var seletedNum int
		for i := 0; i < S; i++ {
			s := mn.Selector(rd)
			if s {
				seletedNum++
			}
		}
		So(seletedNum, ShouldEqual, C)

		isNowNewTurn := mn.IsNowNeedNewTurn()
		So(isNowNewTurn, ShouldBeTrue)
		mn.Reset(55, 100)
		var seletedNum2 int
		for i := 0; i < 100; i++ {
			s := mn.Selector(rd)
			if s {
				seletedNum2++
			}
		}
		So(seletedNum2, ShouldEqual, 55)

	})

	Convey("SelectNone", t, func() {
		const (
			C int = 1
			S int = 0
		)
		mn := MNSelectorState{}
		mn.Init(int64(C), int64(S))
		var seletedNum int
		for i := 0; i < 10; i++ {
			s := mn.Selector(rd)
			if s {
				seletedNum++
			}
		}
		So(seletedNum, ShouldEqual, 0)

	})
}
