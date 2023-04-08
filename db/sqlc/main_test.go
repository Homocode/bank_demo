package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/homocode/bank_demo/util"
	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgres://root:123@localhost:5432/bank?sslmode=disable"
)

var testQueries *Queries
var testDb *sql.DB

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("Can`t load configuration enviroment variables")
	}

	testDb, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Can`t connect to DB", err)
	}

	testQueries = New(testDb)

	os.Exit(m.Run())
}
