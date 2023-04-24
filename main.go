package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"practice-docker/api"
	db "practice-docker/db/sqlc"
)

// constants
const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://root:secret@neidn.com:5432/simple_bank?sslmode=disable"
	serverAddress = ":8080"
)

// dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"

func main() {
	conn, err := sql.Open(dbDriver, dbSource)

	if err != nil {
		log.Fatalln("Failed to connect to database: ", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(serverAddress)

	if err != nil {
		log.Fatalln("Failed to start server: ", err)
	}
}
