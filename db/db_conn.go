package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var PG *pgxpool.Pool

// func PG()*pgxpool.Pool {
// 	return pg
// }

func DBConnect() {

	dbConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to parse configuration: %v\n", err)
		os.Exit(1)
	}
	dbConfig.MaxConns = 10
	PG, err = pgxpool.NewWithConfig(context.Background(), dbConfig)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

}
