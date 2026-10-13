package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"online-mall-platform/server/model"
	"online-mall-platform/server/service"
	"strconv"
)

type Controller struct{ App *service.App }

func Reply(c *gin.Context, status int, data any, e error) {
	if e != nil {
		e = service.Normalize(e)
		var ae *service.Error
		if errors.As(e, &ae) {
			c.JSON(ae.Status, gin.H{"error": ae.Message})
			return
		}
		c.Error(e)
		c.JSON(500, gin.H{"error": "服务暂不可用，请稍后重试"})
		return
	}
	c.JSON(status, gin.H{"data": data})
}
func Bind(c *gin.Context, v any) bool {
	if e := c.ShouldBindJSON(v); e != nil {
		Reply(c, 0, nil, service.Bad("请求字段无效"))
		return false
	}
	return true
}
func ID(c *gin.Context) uint {
	n, e := strconv.ParseUint(c.Param("id"), 10, 32)
	if e != nil || n == 0 {
		Reply(c, 0, nil, service.Bad("ID 无效"))
		c.Abort()
		return 0
	}
	return uint(n)
}
func User(c *gin.Context) model.User { return c.MustGet("user").(model.User) }
func Pagination(c *gin.Context) (int, int) {
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	s, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))
	if p < 1 {
		p = 1
	}
	if p > 100000 {
		p = 100000
	}
	if s < 1 || s > 100 {
		s = 12
	}
	return p, s
}
