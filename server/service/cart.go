package service

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"online-mall-platform/server/model"
)

// Cart mutations and checkout share user -> cart/SKU lock order.
func lockCartUser(tx *gorm.DB, uid uint) error {
	var user model.User
	return Normalize(tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, uid).Error)
}
func bumpCart(tx *gorm.DB, uid uint) error {
	return tx.Model(&model.User{}).Where("id = ?", uid).Update("cart_version", gorm.Expr("cart_version + 1")).Error
}

func (a *App) AddCart(ctx context.Context, uid, vid uint, quantity int) error {
	if !ValidQuantity(quantity) {
		return Bad("数量须为 1–99")
	}
	e := a.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := lockCartUser(tx, uid); e != nil {
			return e
		}
		var v model.Variant
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&v, vid).Error; e != nil {
			return Normalize(e)
		}
		var p model.Product
		if e := tx.First(&p, v.ProductID).Error; e != nil {
			return e
		}
		if !p.Active {
			return Bad("商品已下架")
		}
		var item model.CartItem
		e := tx.Where("user_id = ? AND variant_id = ?", uid, vid).First(&item).Error
		if e != nil && e != gorm.ErrRecordNotFound {
			return e
		}
		n := quantity
		if e == nil {
			n += item.Quantity
		}
		if !ValidQuantity(n) || n > v.Stock {
			return Bad("库存不足或数量超限")
		}
		if e == gorm.ErrRecordNotFound {
			e = tx.Create(&model.CartItem{UserID: uid, VariantID: vid, Quantity: n}).Error
		} else {
			e = tx.Model(&item).Update("quantity", n).Error
		}
		if e != nil {
			return e
		}
		return bumpCart(tx, uid)
	})
	a.Invalidate(ctx, CartKey(uid))
	return e
}
func (a *App) UpdateCart(ctx context.Context, uid, id uint, n int) error {
	if !ValidQuantity(n) {
		return Bad("数量须为 1–99")
	}
	return a.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := lockCartUser(tx, uid); e != nil {
			return e
		}
		var item model.CartItem
		if e := tx.Where("user_id = ?", uid).First(&item, id).Error; e != nil {
			return Normalize(e)
		}
		var v model.Variant
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&v, item.VariantID).Error; e != nil {
			return Normalize(e)
		}
		if n > v.Stock {
			return Bad("库存不足")
		}
		if e := tx.Model(&item).Update("quantity", n).Error; e != nil {
			return e
		}
		return bumpCart(tx, uid)
	})
}
func (a *App) DeleteCart(ctx context.Context, uid, id uint) error {
	return a.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := lockCartUser(tx, uid); e != nil {
			return e
		}
		r := tx.Where("user_id = ?", uid).Delete(&model.CartItem{}, id)
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected == 0 {
			return NotFound()
		}
		return bumpCart(tx, uid)
	})
}
