package model

import "time"

type Base struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type User struct {
	Base
	CartVersion uint64 `json:"-" gorm:"not null;default:0"`
	Email       string `json:"email" gorm:"size:191;uniqueIndex;not null"`
	Password    string `json:"-" gorm:"size:100;not null"`
	Name        string `json:"name" gorm:"size:80"`
	Avatar      string `json:"avatar" gorm:"size:500"`
	Phone       string `json:"phone" gorm:"size:30"`
	Role        string `json:"role" gorm:"size:20;default:user;index"`
	Enabled     bool   `json:"enabled" gorm:"not null"`
}
type Category struct {
	Base
	Name string `json:"name" gorm:"size:80;uniqueIndex;not null"`
}
type Product struct {
	Base
	CacheVersion uint64    `json:"-" gorm:"not null;default:0"`
	CategoryID   uint      `json:"category_id" gorm:"not null;index"`
	Category     Category  `json:"category" gorm:"constraint:OnDelete:RESTRICT"`
	Name         string    `json:"name" gorm:"size:160;not null;index"`
	Description  string    `json:"description" gorm:"type:text"`
	Images       string    `json:"images" gorm:"type:text"`
	Active       bool      `json:"active" gorm:"not null;index"`
	Variants     []Variant `json:"variants" gorm:"foreignKey:ProductID"`
}
type Variant struct {
	Base
	ProductID     uint   `json:"product_id" gorm:"not null;index"`
	SKU           string `json:"sku" gorm:"size:80;uniqueIndex;not null"`
	Specification string `json:"specification" gorm:"size:160"`
	Price         int64  `json:"price" gorm:"not null;check:price > 0"`
	Stock         int    `json:"stock" gorm:"not null;check:stock >= 0"`
}

func (Variant) TableName() string { return "product_variants" }

type CartItem struct {
	Base
	UserID    uint    `json:"user_id" gorm:"uniqueIndex:cart_user_variant;not null"`
	User      User    `json:"-"`
	VariantID uint    `json:"variant_id" gorm:"uniqueIndex:cart_user_variant;not null"`
	Variant   Variant `json:"variant"`
	Quantity  int     `json:"quantity" gorm:"not null;check:quantity > 0"`
}
type Order struct {
	Base
	Number         string      `json:"number" gorm:"size:64;uniqueIndex;not null"`
	UserID         uint        `json:"user_id" gorm:"uniqueIndex:order_user_key;index;not null"`
	User           User        `json:"-"`
	IdempotencyKey string      `json:"-" gorm:"size:80;uniqueIndex:order_user_key;not null"`
	Status         string      `json:"status" gorm:"size:20;index;not null"`
	Total          int64       `json:"total" gorm:"not null"`
	Recipient      string      `json:"recipient" gorm:"size:80;not null"`
	Phone          string      `json:"phone" gorm:"size:30;not null"`
	Address        string      `json:"address" gorm:"size:500;not null"`
	Carrier        string      `json:"carrier" gorm:"size:80"`
	Tracking       string      `json:"tracking" gorm:"size:100"`
	Items          []OrderItem `json:"items"`
}
type OrderItem struct {
	Base
	OrderID       uint    `json:"order_id" gorm:"index;not null"`
	ProductID     uint    `json:"product_id" gorm:"index;not null"`
	Product       Product `json:"-"`
	VariantID     uint    `json:"variant_id" gorm:"index;not null"`
	Variant       Variant `json:"-"`
	Name          string  `json:"name" gorm:"size:160"`
	Specification string  `json:"specification" gorm:"size:160"`
	Price         int64   `json:"price"`
	Quantity      int     `json:"quantity"`
}
type Review struct {
	Base
	UserID      uint      `json:"user_id" gorm:"index;not null"`
	User        User      `json:"user"`
	ProductID   uint      `json:"product_id" gorm:"index;not null"`
	Product     Product   `json:"-"`
	OrderItemID uint      `json:"order_item_id" gorm:"uniqueIndex;not null"`
	OrderItem   OrderItem `json:"-"`
	Rating      int       `json:"rating" gorm:"check:rating >= 1 AND rating <= 5"`
	Content     string    `json:"content" gorm:"size:2000"`
}
type AfterSale struct {
	Base
	OrderID uint   `json:"order_id" gorm:"uniqueIndex;not null"`
	Order   Order  `json:"order"`
	UserID  uint   `json:"user_id" gorm:"index;not null"`
	User    User   `json:"-"`
	Reason  string `json:"reason" gorm:"size:1000"`
	Status  string `json:"status" gorm:"size:20;index"`
	Note    string `json:"note" gorm:"size:1000"`
}
