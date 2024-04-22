package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbSource = "host=localhost port=5432 user=root password=secret dbname=simple_bank sslmode=disable pool_max_conns=10"
)

var testQueries *Queries
var connection *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background();
	config, err := pgxpool.ParseConfig(dbSource)
	if err != nil {
		log.Fatal("cannot parse config:", err)
	}
	connection, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer connection.Close()

	testQueries = New(connection);

	os.Exit(m.Run())
}