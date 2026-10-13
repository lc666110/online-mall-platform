package controller

import (
	"github.com/gin-gonic/gin"
)

func (h Controller) Cart(c *gin.Context) {
	items, e := h.App.Cart(c.Request.Context(), User(c).ID)
	Reply(c, 200, items, e)
}
func (h Controller) AddCart(c *gin.Context) {
	var in struct {
		VariantID uint `json:"variant_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required"`
	}
	if !Bind(c, &in) {
		return
	}
	e := h.App.AddCart(c.Request.Context(), User(c).ID, in.VariantID, in.Quantity)
	Reply(c, 200, true, e)
}
func (h Controller) UpdateCart(c *gin.Context) {
	id := ID(c)
	if id == 0 {
		return
	}
	var in struct {
		Quantity int `json:"quantity" binding:"required"`
	}
	if !Bind(c, &in) {
		return
	}
	Reply(c, 200, true, h.App.UpdateCart(c.Request.Context(), User(c).ID, id, in.Quantity))
}
func (h Controller) DeleteCart(c *gin.Context) {
	id := ID(c)
	if id == 0 {
		return
	}
	Reply(c, 200, true, h.App.DeleteCart(c.Request.Context(), User(c).ID, id))
}
