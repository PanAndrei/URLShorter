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

var (
	ErrURLAlreadyExists = errors.New("url already exists")
)

func newErrURLAlreadyExists() error {
	return ErrURLAlreadyExists
}

type SQLStorage struct {
	DB   *sql.DB
	cnfg string
}

func NewDB(cnfg string) (*SQLStorage, error) {
	db, err := sql.Open("pgx", cnfg)
	if err != nil {
		fmt.Printf("sql.Open error: %v\n", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		fmt.Printf("db.Ping error: %v\n", err)
		return nil, err
	}

	storage := &SQLStorage{
		DB:   db,
		cnfg: cnfg,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := storage.createTableIfNotExists(ctx); err != nil {
		fmt.Printf("createTableIfNotExists error: %v\n", err)
		return nil, err
	}

	return storage, nil
}

func (d *SQLStorage) createTableIfNotExists(ctx context.Context) error {
	_, err := d.DB.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS urls (
           full_url TEXT UNIQUE,
           short_url TEXT,
		   user_id TEXT
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

func (d *SQLStorage) SaveURL(ctx context.Context, u *URL) (*URL, error) {
	if err := d.createTableIfNotExists(ctx); err != nil {
		return nil, err
	}

	if _, err := d.DB.Exec(
		"INSERT INTO urls (full_url, short_url, user_id) VALUES ($1,$2,$3)",
		u.FullURL, u.ShortURL, u.UUID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				err = newErrURLAlreadyExists()
			}
		}
		return nil, err
	}
	return u, nil
}

func (d *SQLStorage) LoadURL(ctx context.Context, u *URL) (*URL, error) {
	var loadedURL URL
	query := "SELECT full_url, short_url FROM urls WHERE short_url = $1 OR full_url = $2"
	err := d.DB.QueryRowContext(ctx, query, u.ShortURL, u.FullURL).Scan(&loadedURL.FullURL, &loadedURL.ShortURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newErrURLNotFound()
		}
		return nil, fmt.Errorf("queryRowContext: %w", err)
	}
	return &loadedURL, nil
}

func (d *SQLStorage) Ping(ctx context.Context) error {
	return d.DB.PingContext(ctx)
}

func (d *SQLStorage) BatchURLS(ctx context.Context, urls []*URL) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		"INSERT INTO urls (full_url, short_url, user_id) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, url := range urls {
		_, err := stmt.Exec(url.FullURL, url.ShortURL)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (d *SQLStorage) GetByUID(ctx context.Context, id string) ([]*URL, error) {
	var urls []*URL
	query := "SELECT full_url, short_url, user_id FROM urls WHERE user_id = $1"
	rows, err := d.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("queryContext: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var url URL
		if err := rows.Scan(&url.FullURL, &url.ShortURL, &url.UUID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		urls = append(urls, &url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return urls, nil
}
