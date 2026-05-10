package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	pgxvector "github.com/pgvector/pgvector-go/pgx"
)

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {

	config, err := pgxpool.ParseConfig("postgres://rinha:rinha@localhost:5432/rinha-2026")
	if err != nil {
		return nil, err
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvector.RegisterTypes(ctx, conn)
	}

	config.MaxConns = 20000
	config.MinConns = 10000

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func FindLegitCounts(ctx context.Context, pool *pgxpool.Pool, embedding [14]float32) int {
	const query = `SELECT COUNT(t.legit) AS LegitCount FROM (
		SELECT legit FROM transactions ORDER BY embedding <-> $1 LIMIT 5) as t
		GROUP BY t.legit
		HAVING t.legit = true
		`

	row := pool.QueryRow(ctx, query, pgvector.NewVector(embedding[:]))

	result := 0
	err := row.Scan(&result)

	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	return result
}
