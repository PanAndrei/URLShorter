package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type SQLStorage struct {
	DB   *sql.DB
	cnfg string
}

func NewDB(cnfg string) *SQLStorage {
	db, err := sql.Open("pgx", cnfg)
	if err != nil {
		fmt.Printf("sql.Open error: %v\n", err)
		return nil
	}

	if err := db.Ping(); err != nil {
		fmt.Printf("db.Ping error: %v\n", err)
		return nil
	}

	storage := &SQLStorage{
		DB:   db,
		cnfg: cnfg,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := storage.createTableIfNotExists(ctx); err != nil {
		fmt.Printf("createTableIfNotExists error: %v\n", err)
		return nil
	}

	return storage
}

func (d *SQLStorage) createTableIfNotExists(ctx context.Context) error {
	_, err := d.DB.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS urls (
           full_url TEXT,
           short_url TEXT,
           id TEXT
        );
    `)
	if err != nil {
		return fmt.Errorf("error creating table: %w", err)
	}
	return nil
}

func (d *SQLStorage) Close() {
	if d.DB != nil {
		d.DB.Close()
	}
}

func (d *SQLStorage) SaveURL(u *URL) {
	ctx := context.Background()

	_, err := d.DB.ExecContext(ctx,
		"INSERT INTO urls (full_url, short_url, id) VALUES ($1, $2, $3)",
		u.FullURL, u.ShortURL, u.ID)

	if err != nil {
		fmt.Printf("Error saving URL: %v\n", err)
	}
}

func (d *SQLStorage) LoadURL(u *URL) (r *URL, err error) {
	ctx := context.Background()
	var loadedURL URL
	query := "SELECT full_url, short_url, id FROM urls WHERE full_url = $1 OR short_url = $2"
	err = d.DB.QueryRowContext(ctx, query, u.FullURL, u.ShortURL).Scan(&loadedURL.FullURL, &loadedURL.ShortURL, &loadedURL.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newErrURLNotFound()
		}
		return nil, fmt.Errorf("queryRowContext: %w", err)
	}
	return &loadedURL, nil
}

func (d *SQLStorage) IsUniqueShort(shortURL string) bool {
	ctx := context.Background()
	var count int
	query := "SELECT COUNT(*) FROM urls WHERE short_url = $1"
	err := d.DB.QueryRowContext(ctx, query, shortURL).Scan(&count)
	if err != nil {
		return false
	}
	return count == 0
}

func (d *SQLStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return d.DB.PingContext(ctx)
}

func (d *SQLStorage) BatchURLS(urls []*URL) error {
	ctx := context.Background()
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction error: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			fmt.Printf("transaction rollback error: %v\n", err)
		}
	}()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO urls (full_url, short_url, id) VALUES ($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("prepare statement error: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			fmt.Printf("statement close error: %v\n", err)
		}
	}()

	for _, url := range urls {
		_, err := stmt.ExecContext(ctx, url.FullURL, url.ShortURL, url.ID)
		if err != nil {
			return fmt.Errorf("statement exec context error: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction error: %w", err)
	}
	return nil
}
