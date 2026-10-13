package router

import (
	"context"
	"github.com/gin-gonic/gin"
	"online-mall-platform/server/controller"
	"online-mall-platform/server/middleware"
	"online-mall-platform/server/service"
	"time"
)

func New(a *service.App) *gin.Engine {
	r := gin.New()
	if e := r.SetTrustedProxies(a.Config.TrustedProxies); e != nil {
		panic(e)
	}
	r.Use(gin.Logger(), gin.Recovery(), middleware.Security())
	h := controller.Controller{App: a}
	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		db, e := a.DB.DB()
		if e != nil || db.PingContext(ctx) != nil || a.Redis.Ping(ctx).Err() != nil {
			c.JSON(503, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})
	api := r.Group("/api/v1")
	api.Use(middleware.Visit(a))
	auth := api.Group("/auth")
	auth.Use(middleware.RateLimit(a, "auth", 30, 60))
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	api.GET("/products", h.Products)
	api.GET("/products/:id", h.Product)
	api.GET("/products/:id/reviews", h.Reviews)
	api.GET("/categories", h.Categories)
	user := api.Group("")
	user.Use(middleware.Auth(a))
	user.POST("/auth/logout", h.Logout)
	user.GET("/me", h.Me)
	user.PATCH("/me", h.Profile)
	user.GET("/cart", h.Cart)
	user.POST("/cart", h.AddCart)
	user.PATCH("/cart/:id", h.UpdateCart)
	user.DELETE("/cart/:id", h.DeleteCart)
	user.GET("/orders", h.Orders)
	user.POST("/orders", h.Checkout)
	user.GET("/orders/:id", h.Order)
	user.PATCH("/orders/:id/status", h.Transition)
	user.POST("/orders/:id/after-sales", h.ApplyAfterSale)
	user.GET("/after-sales", h.AfterSales)
	user.POST("/reviews", h.Review)
	admin := user.Group("/admin")
	admin.Use(middleware.Admin("admin", "staff"))
	admin.GET("/dashboard", h.Dashboard)
	admin.GET("/products", h.Products)
	admin.POST("/products", h.SaveProduct)
	admin.PUT("/products/:id", h.SaveProduct)
	admin.PATCH("/products/:id/status", h.ProductStatus)
	admin.DELETE("/products/:id", h.DeleteProduct)
	admin.GET("/categories", h.Categories)
	admin.POST("/categories", h.SaveCategory)
	admin.PUT("/categories/:id", h.SaveCategory)
	admin.DELETE("/categories/:id", h.DeleteCategory)
	admin.GET("/orders", h.Orders)
	admin.GET("/orders/:id", h.Order)
	admin.PATCH("/orders/:id/status", h.Transition)
	admin.GET("/after-sales", h.AfterSales)
	admin.PATCH("/after-sales/:id", h.ResolveAfterSale)
	users := admin.Group("/users")
	users.Use(middleware.Admin("admin"))
	users.GET("", h.Users)
	users.PATCH("/:id", h.UpdateUser)
	r.NoRoute(func(c *gin.Context) { c.JSON(404, gin.H{"error": "接口不存在"}) })
	return r
}
