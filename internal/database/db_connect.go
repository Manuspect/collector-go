package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConnect(user, pass, host, port, name string) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, name)
	db, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	return db, nil
}
