package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type SQLStorage struct {
	DB   *sql.DB
	cnfg string
}

var (
	ErrURLAlreadyExists = errors.New("url already exists")
)

func newErrURLAlreadyExists() error {
	return ErrURLNotFound
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
			id TEXT,
			CONSTRAINT unique_full_url UNIQUE (full_url)  
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

func (d *SQLStorage) SaveURL(u *URL) (string, error) {
	ctx := context.Background()

	var shortURL string
	query := `
		INSERT INTO urls (full_url, short_url, id) 
		VALUES ($1, $2, $3)
		ON CONFLICT (full_url) DO UPDATE SET id = $3
		RETURNING short_url
	`

	err := d.DB.QueryRowContext(ctx, query, u.FullURL, u.ShortURL, u.ID).Scan(&shortURL)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			var existingURL URL
			err = d.DB.QueryRowContext(ctx, "SELECT short_url FROM urls WHERE full_url = $1", u.FullURL).Scan(&existingURL.ShortURL)
			if err != nil {
				return "", fmt.Errorf("error getting existing short URL: %w", err)
			}

			return existingURL.ShortURL, newErrURLAlreadyExists()

		}
		return "", fmt.Errorf("error saving URL: %w", err)
	}
	return shortURL, nil
}

func (d *SQLStorage) LoadURL(u *URL) (r *URL, err error) {
	ctx := context.Background()
	var loadedURL URL
	query := "SELECT full_url, short_url, id FROM urls WHERE short_url = $1"
	err = d.DB.QueryRowContext(ctx, query, u.ShortURL).Scan(&loadedURL.FullURL, &loadedURL.ShortURL, &loadedURL.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newErrURLNotFound()
		}
		return nil, fmt.Errorf("queryRowContext: %w", err)
	}
	return &loadedURL, nil
}

// func (d *SQLStorage) IsUniqueShort(shortURL string) bool {
// 	ctx := context.Background()
// 	var count int
// 	query := "SELECT COUNT(*) FROM urls WHERE short_url = $1"
// 	err := d.DB.QueryRowContext(ctx, query, shortURL).Scan(&count)
// 	if err != nil {
// 		return false
// 	}
// 	return count == 0
// }

func (d *SQLStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return d.DB.PingContext(ctx)
}

func (d *SQLStorage) BatchURLS(urls []*URL) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		"INSERT INTO urls (full_url, short_url, id) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, url := range urls {
		_, err := stmt.Exec(url.FullURL, url.ShortURL, url.ID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
