package main

import (
	"log"
	"os"

	"github.com/KirillAPDS/go_final_project/pkg/db"
	"github.com/KirillAPDS/go_final_project/pkg/server"
)

func main() {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}
	if err := db.Init(dbFile); err != nil {
		log.Fatalf("DB init failed: %v", err)
	}

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	if err := server.Run(port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
