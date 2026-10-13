package service

import (
	"context"
	"online-mall-platform/server/model"
	"time"
)

type Page[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

func (a *App) Products(query string, category uint, page, size int, admin bool) (Page[model.Product], error) {
	out := Page[model.Product]{Items: []model.Product{}, Page: page, PageSize: size}
	q := a.DB.Model(&model.Product{})
	if !admin {
		q = q.Where("active = ?", true)
	}
	if query != "" {
		q = q.Where("name LIKE ?", "%"+query+"%")
	}
	if category > 0 {
		q = q.Where("category_id = ?", category)
	}
	if e := q.Count(&out.Total).Error; e != nil {
		return out, e
	}
	e := q.Preload("Category").Preload("Variants").Order("id desc").Offset((page - 1) * size).Limit(size).Find(&out.Items).Error
	return out, e
}
func (a *App) Product(ctx context.Context, id uint) (model.Product, error) {
	var p model.Product
	// MySQL gates visibility and selects a committed cache revision. Late cache
	// refills can only populate an obsolete key, never undo a committed edit.
	var current model.Product
	if e := a.DB.WithContext(ctx).Select("id", "active", "cache_version", "category_id").First(&current, id).Error; e != nil {
		return p, Normalize(e)
	}
	if !current.Active {
		return p, NotFound()
	}
	key := ProductKey(id, current.CacheVersion)
	if a.Cache(ctx, key, &p) {
		p.Category = model.Category{}
		if e := a.DB.WithContext(ctx).Where("product_id = ?", id).Find(&p.Variants).Error; e != nil {
			return p, e
		}
		if e := a.DB.WithContext(ctx).First(&p.Category, current.CategoryID).Error; e != nil {
			return p, e
		}
		return p, nil
	}
	e := a.DB.Preload("Category").Preload("Variants").Where("active = ?", true).First(&p, id).Error
	if e == nil {
		a.Store(ctx, key, p, 5*time.Minute)
	}
	return p, Normalize(e)
}
func (a *App) Cart(ctx context.Context, uid uint) ([]model.CartItem, error) {
	items := []model.CartItem{}
	var user model.User
	if e := a.DB.WithContext(ctx).Select("id", "cart_version").First(&user, uid).Error; e != nil {
		return nil, Normalize(e)
	}
	key := CartKey(uid, user.CartVersion)
	if a.Cache(ctx, key, &items) {
		// Cache cart membership and quantities, but always hydrate current SKU prices/stock.
		ids := make([]uint, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.VariantID)
		}
		var variants []model.Variant
		if len(ids) > 0 {
			if e := a.DB.Where("id IN ?", ids).Find(&variants).Error; e != nil {
				return nil, e
			}
		}
		byID := map[uint]model.Variant{}
		for _, v := range variants {
			byID[v.ID] = v
		}
		for i := range items {
			items[i].Variant = byID[items[i].VariantID]
		}
		return items, nil
	}
	e := a.DB.Preload("Variant").Where("user_id = ?", uid).Order("id").Find(&items).Error
	if e == nil {
		a.Store(ctx, key, items, 10*time.Minute)
	}
	return items, e
}
