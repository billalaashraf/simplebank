package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

const (
	dbSource = "host=localhost port=5432 user=root password=secret dbname=simple_bank sslmode=disable"
)

var testQueries *Queries
var connection *pgx.Conn

func TestMain(m *testing.M) {
	ctx := context.Background();
	config, err := pgx.ParseConfig(dbSource)
	if err != nil {
		log.Fatal("cannot parse config:", err)
	}
	connection, err = pgx.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer connection.Close(ctx)

	testQueries = New(connection);

	os.Exit(m.Run())
}