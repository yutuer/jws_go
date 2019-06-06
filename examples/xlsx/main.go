package main

import (
	"github.com/tealeg/xlsx"
	"os"
	"time"
	"vcs.taiyouxi.net/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/errors"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func init() {
	SetTimeLocal("Asia/Shanghai")
}

var local *time.Location

func SetTimeLocal(local_str string) {
	l, err := time.LoadLocation(local_str)
	if err != nil {
		panic(err)
	}
	local = l
}

func main() {
	defer logs.Close()
	excelFileName := "./Code.xlsx"
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		panic(err)
	}
	codeSheet, ok := xlFile.Sheet["Code"]
	if !ok {
		logs.Error("Code Sheet Cannot Found!")
		os.Exit(1)
		return
	}
	for ridx, row := range codeSheet.Rows {
		if ridx >= 4 && len(row.Cells) > 5 {
			data := make([]string, 0, 64)
			for cidx, cell := range row.Cells {
				s, err := cell.String()
				if err != nil {
					logs.Error("cell err by %d,%d in %s", ridx, cidx, err.Error())
					os.Exit(1)
					return
				}
				data = append(data, s)
			}
			logs.Info("row %2d -- %v", ridx, data)

			if data[0] != "" {
				c := CodeData{}
				c.FromXlsx(row.Cells[:])
				logs.Info("data %v", c)
			}
		}
	}
}

type CodeData struct {
	BID        int64
	TID        int64
	Max        int64
	IsLimit    bool
	TimeBegin  int64
	TimeEnd    int64
	Title      string
	ItemIDs    []string
	ItemCounts []uint32
}

func toTime(s string, err error) (int64, error) {
	if err != nil {
		return 0, err
	} else {
		t, err := time.ParseInLocation("2006\\-01\\-02", s, local)
		if err != nil {
			return 0, err
		}
		return t.Unix(), nil
	}
}

func (c *CodeData) FromXlsx(cells []*xlsx.Cell) {
	var err error

	if c.BID, err = cells[0].Int64(); err != nil {
		panic(err)
	}
	if c.TID, err = cells[1].Int64(); err != nil {
		panic(err)
	}
	if c.Title, err = cells[2].String(); err != nil {
		panic(err)
	}
	if c.TimeBegin, err = toTime(cells[3].String()); err != nil {
		panic(err)
	}
	if c.TimeEnd, err = toTime(cells[4].String()); err != nil {
		panic(err)
	}
	if class, err := cells[5].Int(); err != nil {
		panic(err)
	} else {
		if class == 2 {
			c.IsLimit = true
		}
	}
	if c.Max, err = cells[6].Int64(); err != nil {
		panic(err)
	}

	c.ItemIDs = make([]string, 0, 16)
	c.ItemCounts = make([]uint32, 0, 16)
	for i := 7; i < 19 && i < len(cells); i += 2 {
		var (
			id    string
			count int
			err   error
		)
		if id, err = cells[i].String(); err != nil {
			panic(err)
		}
		if id == "" {
			continue
		}
		if count, err = cells[i+1].Int(); err != nil {
			panic(err)
		}
		if id != "" {
			if count <= 0 {
				panic(errors.New("count <= 0"))
			}
			c.ItemIDs = append(c.ItemIDs, id)
			c.ItemCounts = append(c.ItemCounts, uint32(count))
		}
	}
}
