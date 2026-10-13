package controller

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"online-mall-platform/server/model"
	"online-mall-platform/server/service"
	"strconv"
)

func (h Controller) Products(c *gin.Context) {
	page, size := Pagination(c)
	cid, _ := strconv.ParseUint(c.Query("category_id"), 10, 32)
	out, e := h.App.Products(c.Query("q"), uint(cid), page, size, stringsAdmin(c))
	Reply(c, 200, out, e)
}
func stringsAdmin(c *gin.Context) bool { return c.GetBool("admin") }
func (h Controller) Product(c *gin.Context) {
	id := ID(c)
	if id == 0 {
		return
	}
	out, e := h.App.Product(c.Request.Context(), id)
	Reply(c, 200, out, e)
}
func (h Controller) Categories(c *gin.Context) {
	items := []model.Category{}
	e := h.App.DB.Order("id").Find(&items).Error
	Reply(c, 200, items, e)
}
func (h Controller) Reviews(c *gin.Context) {
	page, size := Pagination(c)
	id := ID(c)
	if id == 0 {
		return
	}
	out := service.Page[model.Review]{Items: []model.Review{}, Page: page, PageSize: size}
	q := h.App.DB.Model(&model.Review{}).Where("product_id = ?", id)
	e := q.Count(&out.Total).Error
	if e == nil {
		e = q.Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "name", "avatar") }).Order("id desc").Limit(size).Offset((page - 1) * size).Find(&out.Items).Error
	}
	Reply(c, 200, out, e)
}
