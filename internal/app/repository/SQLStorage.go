package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type SQLStorage struct {
	DB   *sql.DB
	cnfg string
}

func NewDB(cnfg string) *SQLStorage {
	return &SQLStorage{
		cnfg: cnfg,
	}
}

func (d *SQLStorage) Open() error {
	db, err := sql.Open("pgx", d.cnfg)
	if err != nil {
		return fmt.Errorf("sql.Open error: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("db.Ping error: %w", err)
	}
	d.DB = db

	err = d.createTableIfNotExists(context.Background())
	if err != nil {
		return fmt.Errorf("createTableIfNotExists: %w", err)
	}

	return nil
}
func (d *SQLStorage) createTableIfNotExists(ctx context.Context) error {
	rows, err := d.DB.QueryContext(ctx, "SELECT 1 FROM information_schema.tables WHERE table_name = 'urls'")
	if err != nil {
		fmt.Printf("Error during check table existance: %v\n", err) // Выводим ошибку при проверке таблицы
		_, err = d.DB.ExecContext(ctx, `
            CREATE TABLE urls (
               full_url TEXT,
               short_url TEXT,
               uuid INTEGER
            );
        `)
		if err != nil {
			return fmt.Errorf("error creating table: %w", err) // Возвращаем ошибку при создании таблицы
		}
		fmt.Println("Table urls created") // Добавили вывод об успешном создании таблицы
	} else {
		defer rows.Close()
		if err := rows.Err(); err != nil {
			return fmt.Errorf("error checking table existence: %w", err)
		}
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
		"INSERT INTO urls (full_url, short_url, uuid) VALUES ($1, $2, $3)",
		u.FullURL, u.ShortURL, u.UUID)

	if err != nil {
		fmt.Printf("Error saving URL: %v\n", err)
	}
}

func (d *SQLStorage) LoadURL(u *URL) (r *URL, err error) {
	ctx := context.Background()
	var loadedURL URL
	query := "SELECT full_url, short_url, uuid FROM urls WHERE full_url = $1 OR short_url = $2"
	err = d.DB.QueryRowContext(ctx, query, u.FullURL, u.ShortURL).Scan(&loadedURL.FullURL, &loadedURL.ShortURL, &loadedURL.UUID)
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
