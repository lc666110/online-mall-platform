package service

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"online-mall-platform/server/model"
	"time"
)

type Checkout struct {
	Recipient string `json:"recipient" binding:"required,max=80"`
	Phone     string `json:"phone" binding:"required,max=30"`
	Address   string `json:"address" binding:"required,max=500"`
}

func (a *App) Checkout(ctx context.Context, uid uint, key string, in Checkout) (model.Order, error) {
	var order model.Order
	if len(key) < 8 || len(key) > 80 {
		return order, Bad("Idempotency-Key 须为 8–80 个字符")
	}
	var productIDs []uint
	e := a.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Serialize checkouts for the same user so retries observe the committed order.
		var user model.User
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, uid).Error; e != nil {
			return e
		}
		e := tx.Where("user_id = ? AND idempotency_key = ?", uid, key).Preload("Items").First(&order).Error
		if e == nil {
			return nil
		}
		if e != gorm.ErrRecordNotFound {
			return e
		}
		var cart []model.CartItem
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", uid).Order("variant_id").Find(&cart).Error; e != nil {
			return e
		}
		if len(cart) == 0 {
			return Bad("购物车为空")
		}
		order = model.Order{Number: fmt.Sprintf("M%d%s", time.Now().UnixMilli(), RandomID()[:10]), UserID: uid, IdempotencyKey: key, Status: "pending", Recipient: in.Recipient, Phone: in.Phone, Address: in.Address}
		for _, item := range cart {
			var v model.Variant
			if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&v, item.VariantID).Error; e != nil {
				return Normalize(e)
			}
			var p model.Product
			if e := tx.First(&p, v.ProductID).Error; e != nil {
				return e
			}
			if !p.Active || !ValidQuantity(item.Quantity) || v.Stock < item.Quantity {
				return Bad("商品已下架或库存不足")
			}
			if v.Price <= 0 || v.Price > 100000000 {
				return Bad("商品价格异常")
			}
			r := tx.Model(&model.Variant{}).Where("id = ? AND stock >= ?", v.ID, item.Quantity).Update("stock", gorm.Expr("stock - ?", item.Quantity))
			if r.Error != nil {
				return r.Error
			}
			if r.RowsAffected != 1 {
				return Bad("库存不足")
			}
			order.Total += v.Price * int64(item.Quantity)
			order.Items = append(order.Items, model.OrderItem{ProductID: p.ID, VariantID: v.ID, Name: p.Name, Specification: v.Specification, Price: v.Price, Quantity: item.Quantity})
			productIDs = append(productIDs, p.ID)
		}
		if e := tx.Create(&order).Error; e != nil {
			return e
		}
		if e := tx.Where("user_id = ?", uid).Delete(&model.CartItem{}).Error; e != nil {
			return e
		}
		return bumpCart(tx, uid)
	})
	if e == nil {
		a.Invalidate(ctx, CartKey(uid))
		for _, id := range productIDs {
			a.Invalidate(ctx, ProductKey(id))
		}
	}
	return order, e
}
func (a *App) Orders(uid uint, admin bool, page, size int) (Page[model.Order], error) {
	out := Page[model.Order]{Items: []model.Order{}, Page: page, PageSize: size}
	q := a.DB.Model(&model.Order{})
	if !admin {
		q = q.Where("user_id = ?", uid)
	}
	if e := q.Count(&out.Total).Error; e != nil {
		return out, e
	}
	e := q.Preload("Items").Order("id desc").Offset((page - 1) * size).Limit(size).Find(&out.Items).Error
	return out, e
}
func (a *App) Order(uid, id uint, admin bool) (model.Order, error) {
	var o model.Order
	q := a.DB.Preload("Items")
	if !admin {
		q = q.Where("user_id = ?", uid)
	}
	e := q.First(&o, id).Error
	return o, Normalize(e)
}
func (a *App) Transition(ctx context.Context, uid, id uint, role, status, carrier, tracking string) (model.Order, error) {
	var order model.Order
	var pids []uint
	e := a.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items")
		if role == "user" {
			q = q.Where("user_id = ?", uid)
		}
		if e := q.First(&order, id).Error; e != nil {
			return Normalize(e)
		}
		if !CanTransition(order.Status, status, role) {
			return Bad("不允许此订单状态变更")
		}
		if status == "shipped" && (carrier == "" || tracking == "") {
			return Bad("发货需填写物流公司和运单号")
		}
		if status == "cancelled" {
			for _, i := range order.Items {
				if e := tx.Model(&model.Variant{}).Where("id = ?", i.VariantID).Update("stock", gorm.Expr("stock + ?", i.Quantity)).Error; e != nil {
					return e
				}
				pids = append(pids, i.ProductID)
			}
		}
		changes := map[string]any{"status": status}
		if status == "shipped" {
			changes["carrier"] = carrier
			changes["tracking"] = tracking
		}
		if e := tx.Model(&order).Updates(changes).Error; e != nil {
			return e
		}
		order.Status = status
		order.Carrier = carrier
		order.Tracking = tracking
		return nil
	})
	if e == nil {
		for _, pid := range pids {
			a.Invalidate(ctx, ProductKey(pid))
		}
	}
	return order, e
}
func (a *App) Review(uid, itemID uint, rating int, content string) (model.Review, error) {
	review := model.Review{}
	if rating < 1 || rating > 5 {
		return review, Bad("评分须为 1–5")
	}
	e := a.DB.Transaction(func(tx *gorm.DB) error {
		var item model.OrderItem
		if e := tx.First(&item, itemID).Error; e != nil {
			return Normalize(e)
		}
		var o model.Order
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", uid).First(&o, item.OrderID).Error; e != nil {
			return Normalize(e)
		}
		if o.Status != "completed" {
			return Bad("完成订单后才能评价")
		}
		var n int64
		if e := tx.Model(&model.Review{}).Where("order_item_id = ?", itemID).Count(&n).Error; e != nil {
			return e
		}
		if n > 0 {
			return &Error{409, "已评价此商品"}
		}
		review = model.Review{UserID: uid, ProductID: item.ProductID, OrderItemID: itemID, Rating: rating, Content: content}
		return tx.Create(&review).Error
	})
	return review, e
}
func (a *App) ApplyAfterSale(uid, oid uint, reason string) (model.AfterSale, error) {
	var sale model.AfterSale
	e := a.DB.Transaction(func(tx *gorm.DB) error {
		var o model.Order
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", uid).First(&o, oid).Error; e != nil {
			return Normalize(e)
		}
		if o.Status != "shipped" && o.Status != "completed" {
			return Bad("仅已发货/完成订单可申请售后")
		}
		var n int64
		tx.Model(&model.AfterSale{}).Where("order_id = ?", oid).Count(&n)
		if n > 0 {
			return &Error{409, "订单已有售后申请"}
		}
		sale = model.AfterSale{OrderID: oid, UserID: uid, Reason: reason, Status: "pending"}
		return tx.Create(&sale).Error
	})
	return sale, e
}
func (a *App) ResolveAfterSale(ctx context.Context, id uint, status, note string) error {
	if status != "approved" && status != "rejected" {
		return Bad("售后状态无效")
	}
	var pids []uint
	e := a.DB.Transaction(func(tx *gorm.DB) error {
		var s model.AfterSale
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&s, id).Error; e != nil {
			return Normalize(e)
		}
		if s.Status != "pending" {
			return Bad("售后已处理")
		}
		var o model.Order
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&o, s.OrderID).Error; e != nil {
			return e
		}
		if status == "approved" {
			for _, i := range o.Items {
				if e := tx.Model(&model.Variant{}).Where("id = ?", i.VariantID).Update("stock", gorm.Expr("stock + ?", i.Quantity)).Error; e != nil {
					return e
				}
				pids = append(pids, i.ProductID)
			}
			if e := tx.Model(&o).Update("status", "refunded").Error; e != nil {
				return e
			}
		}
		return tx.Model(&s).Updates(map[string]any{"status": status, "note": note}).Error
	})
	if e == nil {
		for _, id := range pids {
			a.Invalidate(ctx, ProductKey(id))
		}
	}
	return e
}
