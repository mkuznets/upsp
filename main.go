package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"mkuznets.com/go/gateway/gateway"
	"mkuznets.com/go/gateway/gateway/store"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}
	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		panic(err)
	}

	db := store.NewDb(pool)
	st := store.New(db)

	api := gateway.NewApi(st)
	api.Start()
}
