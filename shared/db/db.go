package db

import (
	"context"
	"database/sql"
	"embed"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
}

// RunMigrations runs all pending database migrations
func RunMigrations(databaseURL string) error {
	// Open a standard sql.DB for goose (it doesn't support pgxpool)
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	// Register pgx driver for goose
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	log.Println("Running database migrations...")
	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	log.Println("Migrations completed successfully")
	return nil
}

