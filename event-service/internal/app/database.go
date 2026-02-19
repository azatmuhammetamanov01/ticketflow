package app

import (
	"database/sql"

	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/logger"
	_ "github.com/lib/pq"
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
	return nil
}
