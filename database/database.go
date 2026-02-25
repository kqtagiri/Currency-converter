package database

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
)

type Database struct {
	Ctx  context.Context
	Conn *pgx.Conn
}

func (db *Database) InsertConn(ctx context.Context, conn *pgx.Conn) {

	db.Ctx = ctx
	db.Conn = conn

}

func Connect(ctx context.Context) (*pgx.Conn, error) {

	Conn_string := os.Getenv("CONN_STRING")
	return pgx.Connect(ctx, Conn_string)

}

func (db Database) CreateTables() error {

	if err := db.CreateTableUsers(); err != nil {
		return err
	}
	if err := db.CreateTableWallets(); err != nil {
		return err
	}
	if err := db.CreateTableRates(); err != nil {
		return err
	}
	err := db.CreateTableTransactions()
	return err

}
