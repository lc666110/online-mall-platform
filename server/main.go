package main

import (
	"context"
	"log"
	"net/http"
	"online-mall-platform/server/config"
	"online-mall-platform/server/database"
	"online-mall-platform/server/router"
	"online-mall-platform/server/service"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, e := config.Load()
	if e != nil {
		log.Fatal(e)
	}
	db, r, e := database.Open(cfg)
	if e != nil {
		log.Fatal(e)
	}
	defer r.Close()
	sql, e := db.DB()
	if e != nil {
		log.Fatal(e)
	}
	defer sql.Close()
	app := &service.App{DB: db, Redis: r, Config: cfg}
	if e = app.SeedAdmin(); e != nil {
		log.Fatal(e)
	}
	server := &http.Server{Addr: cfg.Address, Handler: router.New(app), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	go func() {
		log.Printf("API listening on %s", cfg.Address)
		if e := server.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			log.Fatal(e)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if e := server.Shutdown(ctx); e != nil {
		log.Print(e)
	}
}
