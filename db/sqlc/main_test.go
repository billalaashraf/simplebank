package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/billalaashraf/simplebank/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var connection *sql.DB

func TestMain(m *testing.M) {

	config, err := util.LoadConfig("../../")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	connection, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testQueries = New(connection)

	os.Exit(m.Run())
}
