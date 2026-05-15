package repository

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	pgxvector "github.com/pgvector/pgvector-go/pgx"
)

var DbSemaphore chan int

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {

	config, err := pgxpool.ParseConfig("postgres://rinha:rinha@db:5432/rinha-2026")
	if err != nil {
		return nil, err
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvector.RegisterTypes(ctx, conn)
	}

	config.MinConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func GetNearFraudTransactionsCount(ctx context.Context, pool *pgxpool.Pool, embedding [14]float32) int {
	const query = `SELECT COUNT(t.legit) AS LegitCount FROM (
		SELECT legit FROM transactions ORDER BY embedding <-> $1 LIMIT 5) as t
		GROUP BY t.legit
		HAVING t.legit = false
		`

	now := time.Now()
	log.Printf("Queries on channel: %d", len(DbSemaphore))
	DbSemaphore <- 1
	row := pool.QueryRow(ctx, query, pgvector.NewVector(embedding[:]))
	<-DbSemaphore
	elapsed := time.Since(now)
	log.Printf("Query took %s", elapsed)
	var result int
	err := row.Scan(&result)

	if err != nil {
		return 0
	}
	return result
}
