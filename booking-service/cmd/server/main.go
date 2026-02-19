package main

import (
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/config"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/app"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/logger"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	if err := logger.Init(cfg.App.Environment); err != nil {
		panic("failed to init logger: " + err.Error())
	}
	defer logger.Sync()

	application := app.New(cfg)
	if err := application.Run(); err != nil {
		logger.Fatal("application failed: " + err.Error())
	}
}
