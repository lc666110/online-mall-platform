package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"online-mall-platform/server/database"
	"online-mall-platform/server/model"
	"online-mall-platform/server/service"
	"os"
	"testing"
	"time"
)

func TestStaleCacheRefillAndVisibility(t *testing.T) {
	if os.Getenv("MYSQL_DSN") == "" {
		t.Skip("real services required")
	}
	cfg, _ := testConfig(t)
	db, r, e := database.Open(cfg)
	if e != nil {
		t.Fatal(e)
	}
	defer r.Close()
	a := &service.App{DB: db, Redis: r, Config: cfg}
	ctx := context.Background()
	u, e := a.Register(service.RandomID()+"@example.com", "Regression123!", "测试")
	if e != nil {
		t.Fatal(e)
	}
	c := model.Category{Name: service.RandomID()}
	db.Create(&c)
	p := model.Product{Name: "库存回归", CategoryID: c.ID, Active: true, Images: "[]", Variants: []model.Variant{{SKU: service.RandomID(), Price: 100, Stock: 3}}}
	db.Create(&p)
	a.AddCart(ctx, u.ID, p.Variants[0].ID, 1)
	old, _ := a.Cart(ctx, u.ID)
	db.First(&u, u.ID)
	key := service.CartKey(u.ID, u.CartVersion)
	a.DeleteCart(ctx, u.ID, old[0].ID)
	a.Store(ctx, key, old, time.Minute)
	items, e := a.Cart(ctx, u.ID)
	if e != nil || len(items) != 0 {
		t.Error("late refill resurrects deleted cart")
	}
	db.Model(&u).Update("role", "admin")
	login, _ := a.Login(ctx, u.Email, "Regression123!")
	handler := New(a)
	oldProduct, _ := a.Product(ctx, p.ID)
	db.Model(&model.Variant{}).Where("id = ?", p.Variants[0].ID).Update("stock", 0)
	body, _ := json.Marshal(map[string]any{"active": false})
	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/admin/products/%d/status", p.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+login.Token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("visibility endpoint %d", w.Code)
	}
	a.Store(ctx, service.ProductKey(p.ID), oldProduct, time.Minute)
	_, e = a.Product(ctx, p.ID)
	if e == nil {
		t.Error("late refill exposes inactive product")
	}
	var v model.Variant
	db.First(&v, p.Variants[0].ID)
	if v.Stock != 0 {
		t.Error("visibility overwrites stock")
	}
}
