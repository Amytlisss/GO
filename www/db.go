package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

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

	// Создание таблицы пользователей
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY generated always as identity,
    name TEXT,
    phone TEXT UNIQUE,
    email TEXT UNIQUE,
    password TEXT,
    role TEXT DEFAULT 'user'
);
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
		return
	}

	// Создание таблицы для встреч
	sqlStmt = `
	CREATE TABLE IF NOT EXISTS meetings (
		id SERIAL PRIMARY KEY,
		user_id INT REFERENCES users(id),
		date TIMESTAMP,
		cancelled BOOLEAN DEFAULT FALSE
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
		return
	}
}

// Функция для обновления времени встречи
func updateMeetingTime(meetingID int, newTime time.Time) error {
	query := `UPDATE meetings SET time = $1 WHERE id = $2`
	_, err := db.Exec(query, newTime, meetingID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении времени встречи: %w", err)
	}
	return nil
}

// Функция для удаления встречи
func deleteMeeting(meetingID int) error {
	query := `DELETE FROM meetings WHERE id = $1`
	_, err := db.Exec(query, meetingID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении встречи: %w", err)
	}
	return nil
}

// Закрытие базы данных
func closeDB() {
	db.Close()
}
