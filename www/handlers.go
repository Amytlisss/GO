package main

import (
	//"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	//"strconv"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var store = sessions.NewCookieStore([]byte("secret-key"))

type User struct {
	ID       int
	Name     string
	Phone    string
	Email    string
	Password string
	Role     string
}

func home_page(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	user, ok := session.Values["user"].(User)

	tmpl, err := template.ParseFiles("templates/home_page.html")
	if err != nil {
		log.Printf("Ошибка при загрузке шаблона: %v", err)
		http.Error(w, "Ошибка при загрузке шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, struct {
		User       User
		IsLoggedIn bool
	}{
		User:       user,
		IsLoggedIn: ok,
	})
	if err != nil {
		log.Printf("Ошибка при выполнении шаблона: %v", err)
		http.Error(w, "Ошибка при выполнении шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

func registerUser(u User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (name, phone, email, password, role) VALUES ($1, $2, $3, $4, $5)", u.Name, u.Phone, u.Email, hashedPassword, u.Role)
	return err
}

func register_page(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("templates/register.html")
		if err != nil {
			log.Printf("Ошибка при загрузке шаблона: %v", err)
			http.Error(w, "Ошибка при загрузке шаблона: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, nil)
	} else if r.Method == http.MethodPost {
		r.ParseForm()

		user := User{
			Name:     r.FormValue("name"),
			Phone:    r.FormValue("phone"),
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
			Role:     "user",
		}

		if err := registerUser(user); err != nil {
			if err.Error() == "pq: duplicate key value violates unique constraint \"users_phone_key\"" {
				http.Error(w, "Пользователь с таким номером телефона уже зарегистрирован", http.StatusBadRequest)
				return
			}
			http.Error(w, "Ошибка при сохранении данных", http.StatusInternalServerError)
			return
		}

		// Создание сессии
		session, _ := store.Get(r, "session-name")
		session.Values["user"] = user
		session.Save(r, w)

		// Перенаправление на личный кабинет
		http.Redirect(w, r, "/user", http.StatusFound)
	}
}

func login_page(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("templates/login.html")
		if err != nil {
			log.Printf("Ошибка при загрузке шаблона: %v", err)
			http.Error(w, "Ошибка при загрузке шаблона: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		phone := r.FormValue("phone")
		password := r.FormValue("password")

		user, err := authenticateUser(phone, password)
		if err != nil {
			http.Error(w, "Неверный номер телефона или пароль", http.StatusUnauthorized)
			return
		}

		// Создание сессии
		session, _ := store.Get(r, "session-name")
		session.Values["user"] = user
		session.Save(r, w)

		// Перенаправление на личный кабинет
		http.Redirect(w, r, "/user", http.StatusFound)
	}
}

func meetingsPage(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	user, ok := session.Values["user"].(User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if r.Method == http.MethodPost {
		dateStr := r.FormValue("date")
		timeStr := r.FormValue("time")

		// Проверяем, что оба поля заполнены
		if dateStr == "" || timeStr == "" {
			log.Println("Дата или время не указаны")
			http.Error(w, "Дата и время должны быть указаны", http.StatusBadRequest)
			return
		}

		// Объединяем дату и время
		dateTimeStr := dateStr + "T" + timeStr
		dateTime, err := time.Parse("2006-01-02T15:04", dateTimeStr)
		if err != nil {
			log.Printf("Ошибка при парсинге даты и времени: %v", err)
			http.Error(w, "Неверный формат даты или времени", http.StatusBadRequest)
			return
		}

		if err := createMeeting(user.ID, dateTime); err != nil {
			log.Printf("Ошибка при создании встречи: %v", err)
			http.Error(w, "Ошибка при создании встречи", http.StatusInternalServerError)
			return
		}
	}

	meetings, err := getMeetings(user.ID)
	if err != nil {
		log.Printf("Ошибка при получении встреч: %v", err)
		http.Error(w, "Ошибка при получении встреч", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/meetings.html")
	if err != nil {
		log.Printf("Ошибка при загрузке шаблона: %v", err)
		http.Error(w, "Ошибка при загрузке шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, struct {
		User     User
		Meetings []Meeting
		Now      time.Time
	}{
		User:     user,
		Meetings: meetings,
		Now:      time.Now(),
	})
	if err != nil {
		log.Printf("Ошибка при выполнении шаблона: %v", err)
		http.Error(w, "Ошибка при выполнении шаблона: "+err.Error(), http.StatusInternalServerError)
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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		phone := r.FormValue("phone")
		password := r.FormValue("password")

		// Проверка пользователя
		user, err := authenticateUser(phone, password)
		if err != nil {
			log.Printf("Ошибка аутентификации: %v", err)
			http.Error(w, "Неверный номер телефона или пароль", http.StatusUnauthorized)
			return
		}

		// Создание сессии
		session, _ := store.Get(r, "session-name")
		session.Values["user"] = user
		if err := session.Save(r, w); err != nil {
			log.Printf("Ошибка сохранения сессии: %v", err)
			http.Error(w, "Ошибка при сохранении сессии", http.StatusInternalServerError)
			return
		}

		// Перенаправление на профиль пользователя
		http.Redirect(w, r, "/user_profile", http.StatusFound)
		return
	}

	// Если метод не POST, отобразите страницу входа
	http.ServeFile(w, r, "templates/login.html")
}

func authenticateUser(phone, password string) (User, error) {
	var user User
	var hashedPassword string

	// Получаем хешированный пароль из базы данных
	err := db.QueryRow("SELECT id, phone, name, email, role, password FROM users WHERE phone = $1", phone).Scan(&user.ID, &user.Phone, &user.Name, &user.Email, &user.Role, &hashedPassword)
	if err != nil {
		return User{}, err
	}

	// Сравниваем введенный пароль с хешированным
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return User{}, err // Неверный пароль
	}

	return user, nil
}

func userProfile(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	user, ok := session.Values["user"].(User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	meetings, err := getMeetings(user.ID)
	if err != nil {
		log.Printf("Ошибка при получении встреч: %v", err)
		http.Error(w, "Ошибка при получении встреч: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		User     User
		Meetings []Meeting
	}{
		User:     user,
		Meetings: meetings,
	}

	tmpl, err := template.ParseFiles("templates/user_profile.html")
	if err != nil {
		log.Printf("Ошибка при загрузке шаблона: %v", err)
		http.Error(w, "Ошибка при загрузке шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Ошибка при выполнении шаблона: %v", err)
		http.Error(w, "Ошибка при выполнении шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}