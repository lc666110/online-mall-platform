package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DSN, RedisAddr, RedisPassword, JWTSecret, Address, AdminEmail, AdminPassword string
	TrustedProxies                                                               []string
	RedisDB                                                                      int
}

func Load() (Config, error) {
	c := Config{DSN: os.Getenv("MYSQL_DSN"), RedisAddr: os.Getenv("REDIS_ADDR"), RedisPassword: os.Getenv("REDIS_PASSWORD"), JWTSecret: os.Getenv("JWT_SECRET"), Address: os.Getenv("LISTEN_ADDR"), AdminEmail: os.Getenv("ADMIN_EMAIL"), AdminPassword: os.Getenv("ADMIN_PASSWORD")}
	if value := os.Getenv("REDIS_DB"); value != "" {
		var e error
		c.RedisDB, e = strconv.Atoi(value)
		if e != nil || c.RedisDB < 0 || c.RedisDB > 15 {
			return c, fmt.Errorf("REDIS_DB must be an integer from 0 to 15")
		}
	}
	for _, proxy := range strings.Split(os.Getenv("TRUSTED_PROXIES"), ",") {
		if proxy = strings.TrimSpace(proxy); proxy != "" {
			c.TrustedProxies = append(c.TrustedProxies, proxy)
		}
	}
	if c.Address == "" {
		c.Address = "127.0.0.1:8080"
	}
	if c.RedisAddr == "" {
		c.RedisAddr = "127.0.0.1:6379"
	}
	if c.DSN == "" || len(c.JWTSecret) < 32 || strings.Contains(c.JWTSecret, "change-me") {
		return c, fmt.Errorf("MYSQL_DSN and a random JWT_SECRET (32+ characters) required")
	}
	return c, nil
}
