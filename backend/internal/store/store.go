package store

import (
	"context"
	"database/sql"
	"time"

	dbschema "automatictools/backend/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func Migrate(db *sql.DB) error {
	migrations := []string{
		dbschema.InitialPostgresSchema,
		dbschema.AddUserContactsPostgres,
	}
	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}
	return nil
}
