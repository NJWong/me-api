package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var Client *sql.DB

func Setup() {
	db, err := sql.Open("mysql", os.Getenv("DSN"))

	if err != nil {
		log.Fatal("Failed to open db connection", err)
	}

	Client = db
}
