package controller

import (
	"github.com/gin-gonic/gin"
	"online-mall-platform/server/service"
	"strings"
)

func (h Controller) Register(c *gin.Context) {
	var in struct {
		Email    string `json:"email" binding:"required,email,max=191"`
		Password string `json:"password" binding:"required,min=10,max=72"`
		Name     string `json:"name" binding:"required,max=80"`
	}
	if !Bind(c, &in) {
		return
	}
	u, e := h.App.Register(in.Email, in.Password, in.Name)
	Reply(c, 201, u, e)
}
func (h Controller) Login(c *gin.Context) {
	var in struct {
		Email    string `json:"email" binding:"required,email,max=191"`
		Password string `json:"password" binding:"required,max=72"`
	}
	if !Bind(c, &in) {
		return
	}
	r, e := h.App.Login(c.Request.Context(), in.Email, in.Password)
	Reply(c, 200, r, e)
}
func (h Controller) Logout(c *gin.Context) {
	e := h.App.Redis.Del(c.Request.Context(), "session:"+c.GetString("sid")).Err()
	Reply(c, 200, true, e)
}
func (h Controller) Me(c *gin.Context) { Reply(c, 200, User(c), nil) }
func (h Controller) Profile(c *gin.Context) {
	var in struct {
		Name   string `json:"name" binding:"required,max=80"`
		Phone  string `json:"phone" binding:"max=30"`
		Avatar string `json:"avatar" binding:"max=500"`
	}
	if !Bind(c, &in) {
		return
	}
	if in.Avatar != "" && !strings.HasPrefix(in.Avatar, "https://") {
		Reply(c, 0, nil, service.Bad("头像需使用 HTTPS 地址"))
		return
	}
	u := User(c)
	e := h.App.DB.Model(&u).Updates(map[string]any{"name": in.Name, "phone": in.Phone, "avatar": in.Avatar}).Error
	Reply(c, 200, u, e)
}
