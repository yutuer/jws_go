package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tealeg/xlsx"
)

func DebugSendChecklist(cls chan<- *CheckList, count int) {
	f := getCheckListFile("Checklist.xlsx")

	rowChan := make(chan *xlsx.Row)
	rowIdx := make(chan int)
	done := make(chan bool)

	go sendChecklistSource(f, rowIdx, rowChan)
	go sendChecklist(rowIdx, rowChan, cls, done)

	close(cls)
	<-done
}

func TestCell2Slice(t *testing.T) {
	expected := []string{"Test1", "Test2", "Test3"}

	assert.Equal(t, expected, Cell2Slice("Test1,Test2,Test3"))
	assert.Equal(t, expected, Cell2Slice("Test3, Test2, Test1 "))
	assert.Equal(t, expected, Cell2Slice(" Test1 , Test3 , Test2 "))
	assert.Equal(t, expected, Cell2Slice(" 		Test2, Test3, Test1\n"))
}

func TestGetCheckListFile(t *testing.T) {
	f := getCheckListFile("Checklist.xlsx")

	assert.True(t, len(f.Sheets) == 2)
	//t.Logf(":", f.Sheets)
}

func TestSendChecklistSource(t *testing.T) {
	f := getCheckListFile("Checklist.xlsx")
	rowChan := make(chan *xlsx.Row)
	rowIdx := make(chan int)

	go sendChecklistSource(f, rowIdx, rowChan)

	rowDict := make(map[int]*xlsx.Row)

	for {
		r, ok := <-rowChan
		if ok {
			i := <-rowIdx
			rowDict[i] = r
		} else {
			break
		}
	}

	assert.NotNil(t, rowDict[4])
	assert.Equal(t, rowDict[4].Cells[COL_CHECK_TYPE].Value, "0")
}

func TestGenChecklist(t *testing.T) {
	f := getCheckListFile("Checklist.xlsx")
	rowChan := make(chan *xlsx.Row)
	rowIdx := make(chan int)
	cls := make(chan *CheckList)
	done := make(chan bool)

	go sendChecklistSource(f, rowIdx, rowChan)
	go sendChecklist(rowIdx, rowChan, cls, done)

	for {
		cl := <-cls
		if cl == nil {
			break
		} else {
			assert.True(t, 10 >= cl.CheckType && cl.CheckType >= 0)
			// t.Logf("The check type is %d", cl.CheckType)
		}
	}

	<-done
}

func TestCheckList_SetDataTarget(t *testing.T) {

}

func TestCheckList_SetFileSource(t *testing.T) {

}
