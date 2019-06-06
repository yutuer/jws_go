package controllers

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

// PlanXController ...
type PlanXController struct {
}

func (p *PlanXController) GetString(c *gin.Context, name string) string {
	return c.DefaultQuery(name, "")
}

func (p *PlanXController) GetInt(c *gin.Context, name string) (int, error) {
	s := c.Query(name)
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}

func (p *PlanXController) GetInt64(c *gin.Context, name string) (int64, error) {
	s := c.Query(name)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseInt(s, 10, 64)
}
