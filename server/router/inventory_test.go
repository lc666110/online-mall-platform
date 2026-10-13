package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"online-mall-platform/server/config"
	"online-mall-platform/server/database"
	"online-mall-platform/server/model"
	"online-mall-platform/server/service"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestInventoryEditsAndCartSerialization(t *testing.T) {
	if os.Getenv("MYSQL_DSN") == "" {
		t.Skip("real services required")
	}
	cfg, e := testConfig(t)
	if e != nil {
		t.Fatal(e)
	}
	db, r, e := database.Open(cfg)
	if e != nil {
		t.Fatal(e)
	}
	defer r.Close()
	a := &service.App{DB: db, Redis: r, Config: cfg}
	ctx := context.Background()
	u, e := a.Register(service.RandomID()+"@example.com", "Inventory123!", "测试")
	if e != nil {
		t.Fatal(e)
	}
	db.Model(&u).Update("role", "admin")
	login, e := a.Login(ctx, u.Email, "Inventory123!")
	if e != nil {
		t.Fatal(e)
	}
	handler := New(a)
	call := func(method, path string, body any) int {
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+login.Token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w.Code
	}
	c := model.Category{Name: service.RandomID()}
	if e := db.Create(&c).Error; e != nil {
		t.Fatal(e)
	}
	p := model.Product{Name: "库存测试", Images: "[]", CategoryID: c.ID, Active: true, Variants: []model.Variant{{SKU: service.RandomID(), Price: 100, Stock: 20}}}
	if e := db.Create(&p).Error; e != nil {
		t.Fatal(e)
	}
	vid := p.Variants[0].ID
	// Simulate checkout between reading an edit form and saving it.
	db.Model(&model.Variant{}).Where("id = ?", vid).Update("stock", 19)
	input := map[string]any{"name": "修改名称", "category_id": c.ID, "active": true, "images": []string{}, "variants": []map[string]any{{"id": vid, "sku": p.Variants[0].SKU, "price": 100, "stock": 20, "expected_stock": 20}}}
	path := fmt.Sprintf("/api/v1/admin/products/%d", p.ID)
	if s := call("PUT", path, input); s != 200 {
		t.Fatalf("metadata edit %d", s)
	}
	var v model.Variant
	db.First(&v, vid)
	if v.Stock != 19 {
		t.Fatal("stale metadata edit overwrote inventory")
	}
	input["variants"].([]map[string]any)[0]["stock"] = 25
	if s := call("PUT", path, input); s != 409 {
		t.Fatalf("stale inventory edit %d", s)
	}
	input["variants"].([]map[string]any)[0]["expected_stock"] = 19
	if s := call("PUT", path, input); s != 200 {
		t.Fatalf("fresh inventory edit %d", s)
	}
	db.First(&v, vid)
	if v.Stock != 25 {
		t.Fatal("inventory update missing")
	}
	if _, e := a.Register(service.RandomID()+"@example.com", strings.Repeat("密", 30), "测试"); e == nil {
		t.Fatal("bcrypt oversized password accepted")
	}
	// Concurrent same-user additions must produce one row with exact quantity.
	var wg sync.WaitGroup
	errs := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- a.AddCart(ctx, u.ID, vid, 1) }()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	cart, e := a.Cart(ctx, u.ID)
	if e != nil || len(cart) != 1 || cart[0].Quantity != 10 {
		t.Fatalf("cart serialization: %v %v", cart, e)
	}
	// A cart write racing checkout must commit before or after it, without a deadlock.
	gate := make(chan struct{})
	errs = make(chan error, 2)
	go func() {
		<-gate
		_, e := a.Checkout(ctx, u.ID, service.RandomID(), service.Checkout{Recipient: "测试", Phone: "123", Address: "测试地址"})
		errs <- e
	}()
	go func() { <-gate; errs <- a.AddCart(ctx, u.ID, vid, 1) }()
	close(gate)
	for i := 0; i < 2; i++ {
		if e := <-errs; e != nil {
			t.Fatal(e)
		}
	}
	var orders []model.Order
	db.Where("user_id = ?", u.ID).Preload("Items").Find(&orders)
	cart, e = a.Cart(ctx, u.ID)
	if e != nil {
		t.Fatal(e)
	}
	qty := 0
	for _, o := range orders {
		for _, i := range o.Items {
			qty += i.Quantity
		}
	}
	for _, i := range cart {
		qty += i.Quantity
	}
	if qty != 11 {
		t.Fatalf("lost or duplicated cart mutation: %d", qty)
	}
}

func TestTrustedProxyClientIP(t *testing.T) {
	for _, tc := range []struct {
		proxies      []string
		remote, want string
	}{{nil, "192.0.2.1:3456", "192.0.2.1"}, {[]string{"172.30.88.0/24"}, "172.30.88.5:3456", "203.0.113.12"}, {[]string{"172.30.88.0/24"}, "192.0.2.1:3456", "192.0.2.1"}} {
		a := &service.App{Config: config.Config{TrustedProxies: tc.proxies}}
		h := New(a)
		h.GET("/ip-test", func(c *gin.Context) { c.String(200, c.ClientIP()) })
		req := httptest.NewRequest("GET", "/ip-test", nil)
		req.RemoteAddr = tc.remote
		req.Header.Set("X-Forwarded-For", "203.0.113.12")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Body.String() != tc.want {
			t.Errorf("client IP %s want %s", w.Body.String(), tc.want)
		}
	}
}
