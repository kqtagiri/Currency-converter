package app

import (
	"Converter/database"
	"context"
	"sync"
)

type App struct {
	Db    database.Database
	RWmtx sync.RWMutex
}

func (app *App) FillAPP() error {

	db := database.Database{}
	ctx := context.Background()
	conn, err := database.Connect(ctx)
	if err != nil {
		return err
	}
	db.InsertConn(ctx, conn)
	app.RWmtx = sync.RWMutex{}
	app.Db = db
	return nil

}
