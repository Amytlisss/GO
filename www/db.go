package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

// Инициализация базы данных
func initDB() {
	var err error
	connStr := "user=postgres password=0000 dbname=priyutik sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	createTables()
}

func createTables() {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT,
			phone TEXT UNIQUE,
			email TEXT UNIQUE,
			password TEXT,
			role TEXT DEFAULT 'user'
		);`,
		`CREATE TABLE IF NOT EXISTS meetings (
			id SERIAL PRIMARY KEY,
			user_id INT REFERENCES users(id),
			date TIMESTAMP,
			cancelled BOOLEAN DEFAULT FALSE
		);`,
	}

	for _, stmt := range tables {
		if _, err := db.Exec(stmt); err != nil {
			log.Fatalf("%q: %s\n", err, stmt)
		}
	}
}

func closeDB() {
	if err := db.Close(); err != nil {
		log.Printf("Ошибка при закрытии базы данных: %v", err)
	}
}
