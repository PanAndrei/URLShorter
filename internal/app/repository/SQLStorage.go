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
           full_url TEXT UNIQUE,
           short_url TEXT
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

// func (d *SQLStorage) SaveURL(u *URL) error {
// 	ctx := context.Background()

// 	if _, err := d.DB.ExecContext(ctx,
// 		"INSERT INTO urls (full_url, short_url, id) VALUES ($1, $2, $3)",
// 		u.FullURL, u.ShortURL, u.ID); err != nil {
// 		var pgErr *pgconn.PgError
// 		if errors.As(err, &pgErr) {
// 			if pgErr.Code == pgerrcode.UniqueViolation {
// 				err = newErrURLAlreadyExists()
// 			}
// 		}

// 		println("db11", err)
// 		return err
// 	}

// 	return nil
// }

func (d *SQLStorage) SaveURL(u *URL) (*URL, error) {
	ctx := context.Background()
	if err := d.createTableIfNotExists(ctx); err != nil {
		return nil, err
	}

	if _, err := d.DB.Exec(
		"INSERT INTO links (full_url, short_url) VALUES ($1,$2)",
		u.FullURL, u.ShortURL); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				err = newErrURLAlreadyExists()
			}
		}
		return nil, err
	}
	return nil, nil

	// ctx := context.Background()

	// if err := d.createTableIfNotExists(ctx); err != nil {
	// 	return nil, err
	// }
	// var existingURL URL
	// err := d.DB.QueryRowContext(ctx,
	// 	`INSERT INTO urls (full_url, short_url, id)
	// 	 VALUES ($1, $2, $3)
	// 	 ON CONFLICT (full_url) DO UPDATE SET id = $3
	// 	 RETURNING full_url, short_url, id`,
	// 	u.FullURL, u.ShortURL, u.ID,
	// ).Scan(&existingURL.FullURL, &existingURL.ShortURL, &existingURL.ID)
	// if err != nil {
	// 	var pgErr *pgconn.PgError
	// 	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
	// 		err = d.DB.QueryRowContext(ctx, "SELECT full_url, short_url, id FROM urls WHERE full_url = $1", u.FullURL).Scan(&existingURL.FullURL, &existingURL.ShortURL, &existingURL.ID)
	// 		if err != nil {
	// 			return nil, fmt.Errorf("error getting existing URL: %w", err)
	// 		}
	// 		return &existingURL, newErrURLAlreadyExists()
	// 	}
	// 	return nil, fmt.Errorf("error saving URL: %w", err)
	// }
	// return &existingURL, nil
}

func (d *SQLStorage) LoadURL(u *URL) (r *URL, err error) {
	ctx := context.Background()
	var loadedURL URL
	query := "SELECT full_url, short_url FROM urls WHERE short_url = $1"
	err = d.DB.QueryRowContext(ctx, query, u.ShortURL).Scan(&loadedURL.FullURL, &loadedURL.ShortURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newErrURLNotFound()
		}
		return nil, fmt.Errorf("queryRowContext: %w", err)
	}
	return &loadedURL, nil
}

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
		"INSERT INTO urls (full_url, short_url) VALUES ($1, $2)")
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
