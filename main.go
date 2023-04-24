package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"practice-docker/api"
	db "practice-docker/db/sqlc"
	"practice-docker/util"
)

func main() {
	config, err := util.LoadConfig(".") // config file is in the same directory as main.go
	if err != nil {
		log.Fatalln("Failed to load config: ", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)

	if err != nil {
		log.Fatalln("Failed to connect to database: ", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)

	if err != nil {
		log.Fatalln("Failed to start server: ", err)
	}
}
