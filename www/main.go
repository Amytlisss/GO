package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/sessions"

	_ "github.com/lib/pq"
)

var db *sql.DB
var store = sessions.NewCookieStore([]byte("secret-key"))

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
    phone TEXT UNIQUE,
    email TEXT UNIQUE,
    role TEXT DEFAULT 'user'
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
	Role  string
}

func home_page(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/home_page.html")
	if err != nil {
		log.Printf("Ошибка при загрузке шаблона: %v", err)
		http.Error(w, "Ошибка при загрузке шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Ошибка при выполнении шаблона: %v", err)
		http.Error(w, "Ошибка при выполнении шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

func registerUser(u User) error {
	_, err := db.Exec("INSERT INTO users (name, phone, email, role) VALUES ($1, $2, $3, $4)", u.Name, u.Phone, u.Email, u.Role)
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
			Role:  "user", // Устанавливаем роль по умолчанию
		}

		if err := registerUser(user); err != nil {
			if err.Error() == "pq: duplicate key value violates unique constraint \"users_phone_key\"" {
				http.Error(w, "Пользователь с таким номером телефона уже зарегистрирован", http.StatusBadRequest)
				return
			}
			http.Error(w, "Ошибка при сохранении данных", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Регистрация успешна! Имя: %s, Телефон: %s, Электронная почта: %s", user.Name, user.Phone, user.Email)
	}
}

func login_page(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("templates/login.html")
		tmpl.Execute(w, nil)
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		phone := r.FormValue("phone")

		var user User
		err := db.QueryRow("SELECT name, phone, email, role FROM users WHERE phone = $1", phone).Scan(&user.Name, &user.Phone, &user.Email, &user.Role)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Пользователь с таким номером телефона не найден", http.StatusUnauthorized)
			} else {
				http.Error(w, "Ошибка при выполнении запроса", http.StatusInternalServerError)
			}
			return
		}

		// Создание сессии
		session, _ := store.Get(r, "session-name")
		session.Values["user"] = user
		session.Save(r, w)

		fmt.Fprintf(w, "Добро пожаловать, %s! Вы вошли как %s.", user.Name, user.Role)
	}
}

func isAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session-name")
		if session.Values["user"] == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func protectedPage(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	user := session.Values["user"].(User)
	fmt.Fprintf(w, "Это защищенная страница. Добро пожаловать, %s! Вы вошли как %s.", user.Name, user.Role)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	delete(session.Values, "user")
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func handleRequest() {
	http.HandleFunc("/", home_page)
	http.HandleFunc("/register", register_page)
	http.HandleFunc("/login", login_page)
	http.Handle("/protected", isAuthenticated(http.HandlerFunc(protectedPage)))
	http.HandleFunc("/logout", logout)
	http.ListenAndServe("0.0.0.0:8080", nil)
}

func main() {
	initDB()
	defer closeDB()
	handleRequest()
}
