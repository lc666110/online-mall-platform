package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"online-mall-platform/server/model"
	"strings"
	"time"
)

type Claims struct {
	UID uint `json:"uid"`
	jwt.RegisteredClaims
}
type LoginResult struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

func RandomID() string {
	var b [24]byte
	if _, e := rand.Read(b[:]); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b[:])
}
func (a *App) Register(email, password, name string) (model.User, error) {
	u := model.User{Email: strings.ToLower(strings.TrimSpace(email)), Name: name, Role: "user", Enabled: true}
	if len(password) > 72 {
		return u, Bad("密码 UTF-8 编码不能超过 72 字节")
	}
	hash, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if e != nil {
		return u, e
	}
	u.Password = string(hash)
	e = a.DB.Create(&u).Error
	if e != nil {
		var n int64
		a.DB.Model(&model.User{}).Where("email = ?", u.Email).Count(&n)
		if n > 0 {
			return u, &Error{409, "邮箱已注册"}
		}
	}
	return u, e
}
func (a *App) Login(ctx context.Context, email, password string) (LoginResult, error) {
	var u model.User
	err := a.DB.Where("email = ?", strings.ToLower(strings.TrimSpace(email))).First(&u).Error
	if err != nil || !u.Enabled || bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
		return LoginResult{}, &Error{401, "邮箱或密码错误"}
	}
	sid := RandomID()
	expiry := time.Now().Add(24 * time.Hour)
	claims := Claims{UID: u.ID, RegisteredClaims: jwt.RegisteredClaims{ID: sid, Issuer: "online-mall-platform", ExpiresAt: jwt.NewNumericDate(expiry), IssuedAt: jwt.NewNumericDate(time.Now())}}
	token, e := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(a.Config.JWTSecret))
	if e != nil {
		return LoginResult{}, e
	}
	if e = a.Redis.Set(ctx, "session:"+sid, fmt.Sprint(u.ID), 24*time.Hour).Err(); e != nil {
		return LoginResult{}, &Error{503, "登录服务暂不可用"}
	}
	return LoginResult{token, u}, nil
}
func (a *App) Authenticate(ctx context.Context, raw string) (model.User, string, error) {
	var u model.User
	c := &Claims{}
	_, e := jwt.ParseWithClaims(raw, c, func(t *jwt.Token) (any, error) { return []byte(a.Config.JWTSecret), nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer("online-mall-platform"), jwt.WithExpirationRequired())
	if e != nil {
		return u, "", &Error{401, "请重新登录"}
	}
	value, e := a.Redis.Get(ctx, "session:"+c.ID).Result()
	if e != nil || value != fmt.Sprint(c.UID) {
		return u, "", &Error{401, "登录已失效"}
	}
	if e = a.DB.First(&u, c.UID).Error; e != nil || !u.Enabled {
		return u, "", &Error{401, "账号不可用"}
	}
	return u, c.ID, nil
}
func (a *App) SeedAdmin() error {
	if a.Config.AdminEmail == "" && a.Config.AdminPassword == "" {
		return nil
	}
	if len(a.Config.AdminPassword) < 12 {
		return fmt.Errorf("ADMIN_PASSWORD must contain 12+ characters")
	}
	var u model.User
	e := a.DB.Where("email = ?", a.Config.AdminEmail).First(&u).Error
	if e == nil {
		return nil
	}
	if e != gorm.ErrRecordNotFound {
		return e
	}
	u, e = a.Register(a.Config.AdminEmail, a.Config.AdminPassword, "管理员")
	if e != nil {
		return e
	}
	return a.DB.Model(&u).Update("role", "admin").Error
}
