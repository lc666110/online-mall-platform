package router

import (
	"github.com/go-sql-driver/mysql"
	"online-mall-platform/server/config"
	"strings"
	"testing"
)

// Refuse accidental writes to an application database when running integration tests.
func testConfig(t *testing.T) (config.Config, error) {
	t.Helper()
	c, e := config.Load()
	if e != nil {
		return c, e
	}
	d, e := mysql.ParseDSN(c.DSN)
	if e != nil {
		return c, e
	}
	if !strings.HasSuffix(d.DBName, "_test") {
		t.Fatalf("integration tests require a dedicated database ending in _test; got %q", d.DBName)
	}
	// Never share ID-based caches/sessions with the local application DB 0.
	c.RedisDB = 15
	return c, nil
}
