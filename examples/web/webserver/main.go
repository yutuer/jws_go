package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

type GetForm struct {
	Num string `form:"num" binding:"required"`
}

type ReturnValue struct {
	Data []string `json:"data"`
}

func (r *ReturnValue) SetInfo(num int64) {
	r.Data = make([]string, 0, int(num))
	for i := 0; i < int(num); i++ {
		r.Data = append(r.Data, fmt.Sprintf("TestItem%d", i))
	}
}

func (r *ReturnValue) ToJson() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func main() {
	r := gin.Default()

	r.Static("/01-html", "../01-html")
	r.Static("/02-css", "../02-css")
	r.Static("/03-dom", "../03-dom")
	r.Static("/04-react", "../04-react")
	r.Static("/05-ajax", "../05-ajax")
	r.Static("/06-uselib", "../06-uselib")
	r.Static("/07-json", "../07-json")
	r.Static("/08-websocket", "../08-websocket")
	r.Static("/09-ant-test", "../09-ant-test")

	r.GET("/api/v1/get", func(c *gin.Context) {
		f := GetForm{}
		c.Bind(&f)
		fnum, err := strconv.ParseInt(f.Num, 10, 64)
		if err != nil {
			c.String(401, "num err!")
			return
		}
		fmt.Printf("d %d", fnum)
		r := ReturnValue{}
		r.SetInfo(fnum)
		c.String(200, r.ToJson())
	})

	go h.run()

	r.GET("/ws", ServeChatWs)

	r.Run(":7788")
}
