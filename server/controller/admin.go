package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"online-mall-platform/server/model"
	"online-mall-platform/server/service"
	"strings"
)

type ProductInput struct {
	CategoryID  uint     `json:"category_id" binding:"required"`
	Name        string   `json:"name" binding:"required,max=160"`
	Description string   `json:"description" binding:"max=20000"`
	Images      []string `json:"images" binding:"max=10,dive,max=500"`
	Active      bool     `json:"active"`
	Variants    []struct {
		ID            uint   `json:"id"`
		SKU           string `json:"sku" binding:"required,max=80"`
		Specification string `json:"specification" binding:"max=160"`
		Price         int64  `json:"price" binding:"required,min=1,max=100000000"`
		Stock         int    `json:"stock" binding:"min=0,max=1000000"`
		ExpectedStock *int   `json:"expected_stock" binding:"omitempty,min=0,max=1000000"`
	} `json:"variants" binding:"required,min=1,max=50,dive"`
}

func (h Controller) SaveProduct(c *gin.Context) {
	var in ProductInput
	if !Bind(c, &in) {
		return
	}
	id := uint(0)
	if c.Request.Method == "PUT" {
		id = ID(c)
		if id == 0 {
			return
		}
	}
	for _, img := range in.Images {
		if !strings.HasPrefix(img, "https://") && !strings.HasPrefix(img, "/images/") {
			Reply(c, 0, nil, service.Bad("图片需使用 HTTPS 或站内图片地址"))
			return
		}
	}
	images, _ := json.Marshal(in.Images)
	p := model.Product{}
	e := h.App.DB.Transaction(func(tx *gorm.DB) error {
		var category model.Category
		if e := tx.First(&category, in.CategoryID).Error; e != nil {
			return service.Bad("分类不存在")
		}
		if id > 0 {
			if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&p, id).Error; e != nil {
				return service.Normalize(e)
			}
		}
		p.CategoryID = in.CategoryID
		p.Name = in.Name
		p.Description = in.Description
		p.Images = string(images)
		p.Active = in.Active
		if id == 0 {
			if e := tx.Create(&p).Error; e != nil {
				return e
			}
		} else {
			p.CacheVersion++
			if e := tx.Model(&p).Select("CategoryID", "Name", "Description", "Images", "Active", "CacheVersion").Updates(&p).Error; e != nil {
				return e
			}
		}
		for _, v := range in.Variants {
			if v.ID == 0 {
				if e := tx.Create(&model.Variant{ProductID: p.ID, SKU: v.SKU, Specification: v.Specification, Price: v.Price, Stock: v.Stock}).Error; e != nil {
					return &service.Error{Status: 409, Message: "SKU 冲突或保存失败"}
				}
			} else {
				if v.ExpectedStock == nil {
					return service.Bad("编辑已有 SKU 需提供 expected_stock")
				}
				var existing model.Variant
				if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("product_id = ?", p.ID).First(&existing, v.ID).Error; e != nil {
					return service.Bad("SKU 不属于此商品")
				}
				changes := map[string]any{"sku": v.SKU, "specification": v.Specification, "price": v.Price}
				// Metadata-only edits preserve concurrent sales. Explicit inventory
				// changes must match the value observed when the form was opened.
				if v.Stock != *v.ExpectedStock {
					if existing.Stock != *v.ExpectedStock {
						return &service.Error{Status: 409, Message: "库存已变化，请刷新后重新调整"}
					}
					changes["stock"] = v.Stock
				}
				r := tx.Model(&model.Variant{}).Where("id = ? AND product_id = ?", v.ID, p.ID).Updates(changes)
				if r.Error != nil {
					return &service.Error{Status: 409, Message: "SKU 冲突或保存失败"}
				}
				if r.RowsAffected == 0 {
					var n int64
					tx.Model(&model.Variant{}).Where("id = ? AND product_id = ?", v.ID, p.ID).Count(&n)
					if n == 0 {
						return service.Bad("SKU 不属于此商品")
					}
				}
			}
		}
		return nil
	})
	if e == nil {
		h.App.Invalidate(c.Request.Context(), service.ProductKey(p.ID))
		e = h.App.DB.Preload("Category").Preload("Variants").First(&p, p.ID).Error
	}
	Reply(c, 200, p, e)
}
func (h Controller) ProductStatus(c *gin.Context) {
	id := ID(c)
	if id == 0 {
		return
	}
	var in struct {
		Active *bool `json:"active" binding:"required"`
	}
	if !Bind(c, &in) {
		return
	}
	r := h.App.DB.Model(&model.Product{}).Where("id = ?", id).Updates(map[string]any{"active": *in.Active, "cache_version": gorm.Expr("cache_version + 1")})
	if r.Error == nil && r.RowsAffected == 0 {
		Reply(c, 0, nil, service.NotFound())
		return
	}
	Reply(c, 200, true, r.Error)
}
func (h Controller) DeleteProduct(c *gin.Context) {
	id := ID(c)
	if id == 0 {
		return
	}
	e := h.App.DB.Transaction(func(tx *gorm.DB) error {
		var p model.Product
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&p, id).Error; e != nil {
			return service.Normalize(e)
		}
		var n int64
		tx.Model(&model.OrderItem{}).Where("product_id = ?", id).Count(&n)
		if n > 0 {
			return &service.Error{Status: 409, Message: "已有订单的商品请下架，保留交易记录"}
		}
		var variants []model.Variant
		tx.Where("product_id = ?", id).Find(&variants)
		ids := []uint{}
		for _, v := range variants {
			ids = append(ids, v.ID)
		}
		tx.Model(&model.CartItem{}).Where("variant_id IN ?", ids).Count(&n)
		if n > 0 {
			return &service.Error{Status: 409, Message: "购物车仍引用此商品，请先下架"}
		}
		if e := tx.Where("product_id = ?", id).Delete(&model.Variant{}).Error; e != nil {
			return e
		}
		return tx.Delete(&p).Error
	})
	h.App.Invalidate(c.Request.Context(), service.ProductKey(id))
	Reply(c, 200, true, e)
}
func (h Controller) SaveCategory(c *gin.Context) {
	var in struct {
		Name string `json:"name" binding:"required,max=80"`
	}
	if !Bind(c, &in) {
		return
	}
	cat := model.Category{Name: in.Name}
	var e error
	if c.Request.Method == "POST" {
		e = h.App.DB.Create(&cat).Error
	} else {
		cat.ID = ID(c)
		if cat.ID == 0 {
			return
		}
		var old model.Category
		if e = h.App.DB.First(&old, cat.ID).Error; e == nil {
			e = h.App.DB.Model(&old).Update("name", in.Name).Error
		}
	}
	if e != nil && e != gorm.ErrRecordNotFound {
		e = &service.Error{Status: 409, Message: "分类名称冲突"}
	}
	Reply(c, 200, cat, e)
}
func (h Controller) DeleteCategory(c *gin.Context) {
	id := ID(c)
	if id == 0 {
		return
	}
	var n int64
	h.App.DB.Model(&model.Product{}).Where("category_id = ?", id).Count(&n)
	if n > 0 {
		Reply(c, 0, nil, &service.Error{Status: 409, Message: "分类仍有商品"})
		return
	}
	r := h.App.DB.Delete(&model.Category{}, id)
	if r.RowsAffected == 0 && r.Error == nil {
		Reply(c, 0, nil, service.NotFound())
		return
	}
	Reply(c, 200, true, r.Error)
}
func (h Controller) Users(c *gin.Context) {
	p, s := Pagination(c)
	out := service.Page[model.User]{Items: []model.User{}, Page: p, PageSize: s}
	q := h.App.DB.Model(&model.User{})
	e := q.Count(&out.Total).Error
	if e == nil {
		e = q.Order("id desc").Limit(s).Offset((p - 1) * s).Find(&out.Items).Error
	}
	Reply(c, 200, out, e)
}
func (h Controller) UpdateUser(c *gin.Context) {
	id := ID(c)
	if id == 0 {
		return
	}
	var in struct {
		Name    string `json:"name" binding:"required,max=80"`
		Phone   string `json:"phone" binding:"max=30"`
		Role    string `json:"role" binding:"required,oneof=user staff admin"`
		Enabled bool   `json:"enabled"`
	}
	if !Bind(c, &in) {
		return
	}
	if id == User(c).ID && (in.Role != "admin" || !in.Enabled) {
		Reply(c, 0, nil, service.Bad("不能移除自己的管理员权限"))
		return
	}
	var u model.User
	e := h.App.DB.First(&u, id).Error
	if e == nil {
		e = h.App.DB.Model(&u).Updates(map[string]any{"name": in.Name, "phone": in.Phone, "role": in.Role, "enabled": in.Enabled}).Error
	}
	Reply(c, 200, u, e)
}

type Trend struct {
	Date   string `json:"date"`
	Sales  int64  `json:"sales"`
	Orders int64  `json:"orders"`
}
type Growth struct {
	Date  string `json:"date"`
	Users int64  `json:"users"`
}
type Ranking struct {
	ProductID uint   `json:"product_id"`
	Name      string `json:"name"`
	Quantity  int64  `json:"quantity"`
	Sales     int64  `json:"sales"`
}
type Dashboard struct {
	Sales    int64     `json:"sales"`
	Orders   int64     `json:"orders"`
	Users    int64     `json:"users"`
	Products int64     `json:"products"`
	Trend    []Trend   `json:"trend"`
	Growth   []Growth  `json:"growth"`
	Ranking  []Ranking `json:"ranking"`
}

func (h Controller) Dashboard(c *gin.Context) {
	d := Dashboard{Trend: []Trend{}, Growth: []Growth{}, Ranking: []Ranking{}}
	db := h.App.DB
	queries := []func() error{
		func() error {
			return db.Model(&model.Order{}).Where("status IN ?", []string{"confirmed", "shipped", "completed"}).Select("COALESCE(SUM(total),0)").Scan(&d.Sales).Error
		},
		func() error { return db.Model(&model.Order{}).Count(&d.Orders).Error }, func() error { return db.Model(&model.User{}).Count(&d.Users).Error }, func() error { return db.Model(&model.Product{}).Count(&d.Products).Error },
		func() error {
			return db.Raw("SELECT DATE(created_at) date, COUNT(*) orders, COALESCE(SUM(CASE WHEN status IN ('confirmed','shipped','completed') THEN total ELSE 0 END),0) sales FROM orders WHERE created_at >= UTC_TIMESTAMP() - INTERVAL 30 DAY GROUP BY DATE(created_at) ORDER BY date").Scan(&d.Trend).Error
		},
		func() error {
			return db.Raw("SELECT DATE(created_at) date, COUNT(*) users FROM users WHERE created_at >= UTC_TIMESTAMP() - INTERVAL 30 DAY GROUP BY DATE(created_at) ORDER BY date").Scan(&d.Growth).Error
		},
		func() error {
			return db.Raw("SELECT oi.product_id, MAX(oi.name) name, SUM(oi.quantity) quantity, SUM(oi.price*oi.quantity) sales FROM order_items oi JOIN orders o ON o.id=oi.order_id WHERE o.status IN ('confirmed','shipped','completed') GROUP BY oi.product_id ORDER BY quantity DESC LIMIT 10").Scan(&d.Ranking).Error
		}}
	for _, q := range queries {
		if e := q(); e != nil {
			Reply(c, 0, nil, e)
			return
		}
	}
	Reply(c, 200, d, nil)
}
