package atc_mng

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
)

func RouteProfileGet(r *gin.Engine, url string) {
	r.GET("/api/v1/profiles/", func(c *gin.Context) {
		data, err := httpGet(url + "/api/v1/profiles/")
		if err == nil {
			c.String(200, string(data))
		} else {
			c.String(400, err.Error())
		}
	})
}

func RouteGetClientAll(r *gin.Engine, db_address string) {
	r.GET("/api/v1/clientall/", func(c *gin.Context) {
		data, err := GetAllAtcdClient(db_address)
		if err == nil {
			b, jsonerr := json.Marshal(data)
			if jsonerr != nil {
				c.String(400, jsonerr.Error())
			}
			c.String(200, string(b))
		} else {
			c.String(400, err.Error())
		}
	})

	// http://127.0.0.1/api/v1/clientadd/127.1.1.2/name
	r.GET("/api/v1/clientadd/:ip/*name", func(c *gin.Context) {

		ip := c.Param("ip")
		name := c.Param("name")
		if name[0] == '/' {
			name = name[1:]
		}
		if ip != "" {
			AddAddonIp(ip, name)
		}
	})
}

type AtcSettingOne struct {
	Rate  int `json:"rate"`
	Delay struct {
		Delay       int `json:"delay"`
		Jitter      int `json:"jitter"`
		Correlation int `json:"correlation"`
	} `json:"delay"`
	Loss struct {
		Percentage  int `json:"percentage"`
		Correlation int `json:"correlation"`
	} `json:"loss"`
	Reorder struct {
		Percentage  int `json:"percentage"`
		Correlation int `json:"correlation"`
		Gap         int `json:"gap"`
	} `json:"reorder"`
	Corruption struct {
		Percentage  int `json:"percentage"`
		Correlation int `json:"correlation"`
	} `json:"corruption"`
	Iptables_options []string `json:"iptables_options"`
}

type AtcSetting struct {
	Down AtcSettingOne `json:"down"`
	Up   AtcSettingOne `json:"up"`
}

func RouteProxy(r *gin.Engine, url string) {
	r.GET("/api/v1/shape/:name", func(c *gin.Context) {
		data, err := httpGet(url + "/api/v1/shape/" + c.Param("name") + "/")
		if err != nil {
			c.String(400, err.Error())
		} else {
			c.String(200, string(data))
		}
	})

	r.POST("/api/v1/shape/:name", func(c *gin.Context) {
		s := AtcSetting{}
		err := c.Bind(&s)

		fmt.Printf("data : %v\n", s)

		if err != nil {
			fmt.Printf("err %v\n", err.Error())
		}
		b, _ := json.Marshal(s)
		data, err := httpPost(url+"/api/v1/shape/"+c.Param("name")+"/", "application/json; charset=utf-8", b)

		c.String(200, string(data))
	})

	r.DELETE("/api/v1/shape/:name", func(c *gin.Context) {
		name := c.Param("name")
		fmt.Printf("data : %v\n", name)
		data, err := httpDelete(url + "/api/v1/shape/" + name + "/")
		if err != nil {
			c.String(400, err.Error())
		} else {
			fmt.Printf("data res : %v\n", string(data))
			c.String(200, string(data))
		}
	})
}
