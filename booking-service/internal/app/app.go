package app

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/config"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/client"
	grpclib "google.golang.org/grpc"
)

type App struct {
	cfg         *config.Config
	db          *sql.DB
	grpcServer  *grpclib.Server
	httpServer  *http.Server
	eventClient client.EventClient
}

func New(cfg *config.Config) *App {
	return &App{cfg: cfg}
}

func (a *App) Run() error {
	if err := a.initDB(); err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}

	if err := a.initServers(); err != nil {
		return fmt.Errorf("failed to init servers: %w", err)
	}

	return a.start()
}
