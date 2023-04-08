package main

import (
	"database/sql"
	"log"

	api "github.com/homocode/bank_demo/api"
	db "github.com/homocode/bank_demo/db/sqlc"
	"github.com/homocode/bank_demo/util"
	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Can`t load configuration enviroment variables")
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Can`t connect to DB", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)
	err = server.Start(config.ServerAddress)

	if err != nil {
		log.Fatal("Can't start server", err)
	}
}
