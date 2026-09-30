package app

import (
	"database/sql"
	"fmt"

	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/logger"
	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/migrations"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func (a *App) initDB() error {
	db, err := sql.Open("postgres", a.cfg.Database.DSN())
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	a.db = db
	logger.Info("Connected to database")

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	logger.Info("Database migrations applied successfully")

	return nil
}
