package controller

import (
	"github.com/gin-gonic/gin"
	"online-mall-platform/server/model"
	"online-mall-platform/server/service"
)

func (h Controller) Checkout(c *gin.Context) {
	var in service.Checkout
	if !Bind(c, &in) {
		return
	}
	out, e := h.App.Checkout(c.Request.Context(), User(c).ID, c.GetHeader("Idempotency-Key"), in)
	Reply(c, 201, out, e)
}
func (h Controller) Orders(c *gin.Context) {
	p, s := Pagination(c)
	out, e := h.App.Orders(User(c).ID, c.GetBool("admin"), p, s)
	Reply(c, 200, out, e)
}
func (h Controller) Order(c *gin.Context) {
	id := ID(c)
	if id == 0 {
		return
	}
	out, e := h.App.Order(User(c).ID, id, c.GetBool("admin"))
	Reply(c, 200, out, e)
}
func (h Controller) Transition(c *gin.Context) {
	id := ID(c)
	if id == 0 {
		return
	}
	var in struct {
		Status   string `json:"status" binding:"required,max=20"`
		Carrier  string `json:"carrier" binding:"max=80"`
		Tracking string `json:"tracking" binding:"max=100"`
	}
	if !Bind(c, &in) {
		return
	}
	u := User(c)
	out, e := h.App.Transition(c.Request.Context(), u.ID, id, u.Role, in.Status, in.Carrier, in.Tracking)
	Reply(c, 200, out, e)
}
func (h Controller) Review(c *gin.Context) {
	var in struct {
		OrderItemID uint   `json:"order_item_id" binding:"required"`
		Rating      int    `json:"rating" binding:"required,min=1,max=5"`
		Content     string `json:"content" binding:"required,max=2000"`
	}
	if !Bind(c, &in) {
		return
	}
	out, e := h.App.Review(User(c).ID, in.OrderItemID, in.Rating, in.Content)
	Reply(c, 201, out, e)
}
func (h Controller) AfterSales(c *gin.Context) {
	items := []model.AfterSale{}
	q := h.App.DB.Preload("Order.Items")
	if !c.GetBool("admin") {
		q = q.Where("user_id = ?", User(c).ID)
	}
	p, s := Pagination(c)
	out := service.Page[model.AfterSale]{Items: items, Page: p, PageSize: s}
	e := q.Model(&model.AfterSale{}).Count(&out.Total).Error
	if e == nil {
		e = q.Order("id desc").Limit(s).Offset((p - 1) * s).Find(&out.Items).Error
	}
	Reply(c, 200, out, e)
}
func (h Controller) ApplyAfterSale(c *gin.Context) {
	id := ID(c)
	if id == 0 {
		return
	}
	var in struct {
		Reason string `json:"reason" binding:"required,max=1000"`
	}
	if !Bind(c, &in) {
		return
	}
	out, e := h.App.ApplyAfterSale(User(c).ID, id, in.Reason)
	Reply(c, 201, out, e)
}
func (h Controller) ResolveAfterSale(c *gin.Context) {
	id := ID(c)
	if id == 0 {
		return
	}
	var in struct {
		Status string `json:"status" binding:"required"`
		Note   string `json:"note" binding:"max=1000"`
	}
	if !Bind(c, &in) {
		return
	}
	Reply(c, 200, true, h.App.ResolveAfterSale(c.Request.Context(), id, in.Status, in.Note))
}
