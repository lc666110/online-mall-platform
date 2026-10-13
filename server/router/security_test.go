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
	"sync"
	"testing"
	"time"
)

func TestSecurityAndConcurrency(t *testing.T) {
	if os.Getenv("MYSQL_DSN") == "" {
		t.Skip("MYSQL_DSN required")
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
	handler := New(a)
	ctx := context.Background()
	call := func(method, path, token string, body any, key string) (int, map[string]any) {
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Idempotency-Key", key)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		var out map[string]any
		json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}
	createUser := func(role string) (model.User, string) {
		u, e := a.Register(fmt.Sprintf("%s-%s@example.com", role, service.RandomID()), "SecurityPass123!", "测试用户")
		if e != nil {
			t.Fatal(e)
		}
		if role != "user" {
			db.Model(&u).Update("role", role)
		}
		login, e := a.Login(ctx, u.Email, "SecurityPass123!")
		if e != nil {
			t.Fatal(e)
		}
		return u, login.Token
	}
	u1, t1 := createUser("user")
	u2, t2 := createUser("user")
	_, admin := createUser("admin")
	_, staff := createUser("staff")
	cat := model.Category{Name: "安全测试" + service.RandomID()[:8]}
	db.Create(&cat)
	p := model.Product{Name: "并发库存商品", CategoryID: cat.ID, Active: true, Images: "[]", Variants: []model.Variant{{SKU: service.RandomID(), Specification: "限量", Price: 1000, Stock: 1}}}
	if e := db.Create(&p).Error; e != nil {
		t.Fatal(e)
	}
	v := p.Variants[0]
	if e := a.AddCart(ctx, u1.ID, v.ID, 1); e != nil {
		t.Fatal(e)
	}
	if e := a.AddCart(ctx, u2.ID, v.ID, 1); e != nil {
		t.Fatal(e)
	}
	_, e = a.Cart(ctx, u1.ID)
	if e != nil {
		t.Fatal(e)
	}
	db.Model(&model.Variant{}).Where("id = ?", v.ID).Update("price", 1200)
	cart, e := a.Cart(ctx, u1.ID)
	if e != nil {
		t.Fatal(e)
	}
	if len(cart) != 1 || cart[0].Variant.Price != 1200 {
		t.Error("cart cache serves stale product prices")
	}
	statuses := make(chan int, 2)
	var wg sync.WaitGroup
	checkout := map[string]any{"recipient": "安全测试", "phone": "123456", "address": "测试地址"}
	for _, token := range []string{t1, t2} {
		wg.Add(1)
		go func(token string) {
			defer wg.Done()
			s, _ := call("POST", "/api/v1/orders", token, checkout, service.RandomID())
			statuses <- s
		}(token)
	}
	wg.Wait()
	close(statuses)
	success, reject := 0, 0
	for s := range statuses {
		if s == 201 {
			success++
		} else if s == 400 {
			reject++
		} else {
			t.Errorf("unexpected concurrent status %d", s)
		}
	}
	if success != 1 || reject != 1 {
		t.Fatalf("oversell: %d successes, %d rejects", success, reject)
	}
	var order model.Order
	db.Where("user_id IN ?", []uint{u1.ID, u2.ID}).Preload("Items").First(&order)
	owner, other := t1, t2
	if order.UserID == u2.ID {
		owner, other = t2, t1
	}
	s, _ := call("GET", fmt.Sprintf("/api/v1/orders/%d", order.ID), other, nil, "")
	if s != 404 {
		t.Errorf("order disclosure %d", s)
	}
	s, _ = call("PATCH", fmt.Sprintf("/api/v1/orders/%d/status", order.ID), other, map[string]any{"status": "cancelled"}, "")
	if s != 404 {
		t.Errorf("order mutation %d", s)
	}
	s, _ = call("POST", "/api/v1/reviews", owner, map[string]any{"order_item_id": order.Items[0].ID, "rating": 5, "content": "提前评价"}, "")
	if s != 400 {
		t.Errorf("premature review %d", s)
	}
	for _, transition := range []struct{ status, carrier, tracking string }{{"confirmed", "", ""}, {"shipped", "测试快递", "TRACK1"}} {
		s, _ = call("PATCH", fmt.Sprintf("/api/v1/admin/orders/%d/status", order.ID), admin, transitionMap(transition.status, transition.carrier, transition.tracking), "")
		if s != 200 {
			t.Fatalf("transition %s %d", transition.status, s)
		}
	}
	s, _ = call("PATCH", fmt.Sprintf("/api/v1/orders/%d/status", order.ID), owner, map[string]any{"status": "completed"}, "")
	if s != 200 {
		t.Fatal("complete")
	}
	review := map[string]any{"order_item_id": order.Items[0].ID, "rating": 5, "content": "合法评价"}
	s, _ = call("POST", "/api/v1/reviews", owner, review, "")
	if s != 201 {
		t.Fatalf("review %d", s)
	}
	s, _ = call("POST", "/api/v1/reviews", owner, review, "")
	if s != 409 {
		t.Fatalf("duplicate review %d", s)
	}
	s, out := call("POST", fmt.Sprintf("/api/v1/orders/%d/after-sales", order.ID), owner, map[string]any{"reason": "测试退货"}, "")
	if s != 201 {
		t.Fatal(out)
	}
	saleID := uint(out["data"].(map[string]any)["id"].(float64))
	s, _ = call("PATCH", fmt.Sprintf("/api/v1/admin/after-sales/%d", saleID), admin, map[string]any{"status": "approved", "note": "测试"}, "")
	if s != 200 {
		t.Fatalf("approve %d", s)
	}
	s, _ = call("PATCH", fmt.Sprintf("/api/v1/admin/after-sales/%d", saleID), admin, map[string]any{"status": "approved"}, "")
	if s != 400 {
		t.Fatalf("repeat approve %d", s)
	}
	db.First(&v, v.ID)
	if v.Stock != 1 {
		t.Fatalf("double restock %d", v.Stock)
	}
	s, _ = call("GET", "/api/v1/admin/users", staff, nil, "")
	if s != 403 {
		t.Errorf("staff privilege %d", s)
	}
	s, _ = call("PATCH", "/api/v1/me", t1, map[string]any{"name": "试图提权", "role": "admin"}, "")
	if s != 200 {
		t.Fatal("profile")
	}
	db.First(&u1, u1.ID)
	if u1.Role != "user" {
		t.Fatal("mass assignment")
	}
	db.Model(&u1).Update("enabled", false)
	s, _ = call("GET", "/api/v1/me", t1, nil, "")
	if s != 401 {
		t.Errorf("disabled auth %d", s)
	}
	keys, e := r.Keys(ctx, "product:*").Result()
	if e != nil {
		t.Fatal(e)
	}
	_ = keys
	var keys2 []string
	keys2, _, _ = r.Scan(ctx, 0, "visits:*", 100).Result()
	if len(keys2) == 0 {
		t.Error("visit cache absent")
	}
	_ = time.Second
}
func transitionMap(status, carrier, tracking string) map[string]any {
	return map[string]any{"status": status, "carrier": carrier, "tracking": tracking}
}
