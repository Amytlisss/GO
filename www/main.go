package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

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

	// Создание таблицы
	sqlStmt := `
    CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT,
    phone TEXT,
    email TEXT
);
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
		return
	}
}

// Закрытие базы данных
func closeDB() {
	db.Close()
}

type User struct {
	Name  string
	Phone string
	Email string
}

func home_page(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/home_page.html")
	tmpl.Execute(w, nil)
}

func registerUser(u User) error {
	_, err := db.Exec("INSERT INTO users (name, phone, email) VALUES ($1, $2, $3)", u.Name, u.Phone, u.Email)
	return err
}
func register_page(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("templates/register.html")
		tmpl.Execute(w, nil)
	} else if r.Method == http.MethodPost {
		r.ParseForm()

		user := User{
			Name:  r.FormValue("name"),
			Phone: r.FormValue("phone"),
			Email: r.FormValue("email"),
		}

		if err := registerUser(user); err != nil {
			http.Error(w, "Ошибка при сохранении данных", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Регистрация успешна! Имя: %s, Телефон: %s, Электронная почта: %s", user.Name, user.Phone, user.Email)
	}
}

func handleRequest() {
	http.HandleFunc("/", home_page)
	http.HandleFunc("/register", register_page) // Убедитесь, что этот маршрут добавлен
	http.ListenAndServe("0.0.0.0:8080", nil)
}

func main() {
	initDB()
	defer closeDB()
	handleRequest()
}
