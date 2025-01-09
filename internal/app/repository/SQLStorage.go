package repository

import (
	"database/sql"

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
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	d.DB = db
	return nil
}

func (d *SQLStorage) Close() {
	d.DB.Close()
}
