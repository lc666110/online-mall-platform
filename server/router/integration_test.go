package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"online-mall-platform/server/database"
	"online-mall-platform/server/model"
	"online-mall-platform/server/service"
	"os"
	"testing"
)

func TestIntegration(t *testing.T) {
	if os.Getenv("MYSQL_DSN") == "" {
		t.Skip("MYSQL_DSN required for real MySQL/Redis integration")
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
	app := &service.App{DB: db, Redis: r, Config: cfg}
	handler := New(app)
	call := func(method, path, token string, body any, key string) (int, map[string]any) {
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		var out map[string]any
		json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}
	email := fmt.Sprintf("test-%d@example.com", os.Getpid())
	db.Where("email = ?", email).Delete(&model.User{})
	s, _ := call("POST", "/api/v1/auth/register", "", map[string]any{"email": email, "password": "TestPass123!", "name": "Tester"}, "")
	if s != 201 {
		t.Fatalf("register %d", s)
	}
	s, _ = call("POST", "/api/v1/auth/register", "", map[string]any{"email": email, "password": "TestPass123!", "name": "Tester"}, "")
	if s != 409 {
		t.Fatalf("duplicate %d", s)
	}
	s, _ = call("POST", "/api/v1/auth/login", "", map[string]any{"email": email, "password": "wrong"}, "")
	if s != 401 {
		t.Fatalf("wrong login %d", s)
	}
	s, out := call("POST", "/api/v1/auth/login", "", map[string]any{"email": email, "password": "TestPass123!"}, "")
	if s != 200 {
		t.Fatal(out)
	}
	token := out["data"].(map[string]any)["token"].(string)
	s, _ = call("GET", "/api/v1/admin/users", token, nil, "")
	if s != 403 {
		t.Fatalf("RBAC %d", s)
	}
	var u model.User
	db.Where("email = ?", email).First(&u)
	cat := model.Category{Name: fmt.Sprintf("test-%d", os.Getpid())}
	db.FirstOrCreate(&cat, model.Category{Name: cat.Name})
	p := model.Product{Name: "Integration Product", CategoryID: cat.ID, Active: true, Images: "[]", Variants: []model.Variant{{SKU: fmt.Sprintf("test-%d", os.Getpid()), Price: 1234, Stock: 2, Specification: "default"}}}
	db.Where("sku = ?", p.Variants[0].SKU).Delete(&model.Variant{})
	if e = db.Create(&p).Error; e != nil {
		t.Fatal(e)
	}
	s, _ = call("POST", "/api/v1/cart", token, map[string]any{"variant_id": p.Variants[0].ID, "quantity": 2}, "")
	if s != 200 {
		t.Fatalf("cart %d", s)
	}
	s, _ = call("POST", "/api/v1/cart", token, map[string]any{"variant_id": p.Variants[0].ID, "quantity": -1}, "")
	if s != 400 {
		t.Fatalf("quantity %d", s)
	}
	s, _ = call("GET", "/api/v1/cart", token, nil, "")
	db.First(&u, u.ID)
	if s != 200 || r.Exists(context.Background(), service.CartKey(u.ID, u.CartVersion)).Val() != 1 {
		t.Fatal("cart cache absent")
	}
	checkout := map[string]any{"recipient": "测试", "phone": "13800138000", "address": "测试地址"}
	key := fmt.Sprintf("checkout-%d", os.Getpid())
	s, out = call("POST", "/api/v1/orders", token, checkout, key)
	if s != 201 {
		t.Fatalf("checkout %d %v", s, out)
	}
	order := out["data"].(map[string]any)
	oid := uint(order["id"].(float64))
	if order["total"].(float64) != 2468 {
		t.Fatal("incorrect total")
	}
	s, out = call("POST", "/api/v1/orders", token, checkout, key)
	if s != 201 || uint(out["data"].(map[string]any)["id"].(float64)) != oid {
		t.Fatal("idempotency failed")
	}
	var v model.Variant
	db.First(&v, p.Variants[0].ID)
	if v.Stock != 0 {
		t.Fatal("stock deduction")
	}
	s, _ = call("PATCH", fmt.Sprintf("/api/v1/orders/%d/status", oid), token, map[string]any{"status": "confirmed"}, "")
	if s != 400 {
		t.Fatalf("user confirmation %d", s)
	}
	s, _ = call("PATCH", fmt.Sprintf("/api/v1/orders/%d/status", oid), token, map[string]any{"status": "cancelled"}, "")
	if s != 200 {
		t.Fatalf("cancel %d", s)
	}
	s, _ = call("PATCH", fmt.Sprintf("/api/v1/orders/%d/status", oid), token, map[string]any{"status": "cancelled"}, "")
	if s != 400 {
		t.Fatalf("repeat cancel %d", s)
	}
	db.First(&v, v.ID)
	if v.Stock != 2 {
		t.Fatal("restock")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("AdminTest123!"), bcrypt.MinCost)
	admin := model.User{Email: fmt.Sprintf("admin-test-%d@example.com", os.Getpid()), Password: string(hash), Role: "admin", Enabled: true}
	db.Create(&admin)
	_, out = call("POST", "/api/v1/auth/login", "", map[string]any{"email": admin.Email, "password": "AdminTest123!"}, "")
	at := out["data"].(map[string]any)["token"].(string)
	s, _ = call("GET", "/api/v1/admin/dashboard", at, nil, "")
	if s != 200 {
		t.Fatalf("dashboard %d", s)
	}
	s, _ = call("POST", "/api/v1/auth/logout", token, nil, "")
	if s != http.StatusOK {
		t.Fatal("logout")
	}
	s, _ = call("GET", "/api/v1/me", token, nil, "")
	if s != 401 {
		t.Fatal("revocation")
	}
}
