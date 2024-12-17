// package main

// import (
// 	"net/http"
// )

// func handleRequest() {
// 	http.HandleFunc("/", home_page)
// 	http.HandleFunc("/register", register_page)
// 	http.HandleFunc("/login", login_page)
// 	http.Handle("/protected", isAuthenticated(http.HandlerFunc(protectedPage)))
// 	http.Handle("/user", isAuthenticated(http.HandlerFunc(userProfile)))
// 	http.Handle("/meetings", isAuthenticated(meetingsPage))               // Страница встреч
// 	http.Handle("/cancel_meeting", isAuthenticated(cancelMeetingHandler)) // Обработчик отмены встречи
// 	http.HandleFunc("/meetings/update", isAuthenticated(updateMeetingTimeHandler).Methods("POST"))
// 	http.HandleFunc("/meetings/delete", isAuthenticated(deleteMeetingHandler).Methods("POST"))
// 	http.HandleFunc("/logout", logout) // Обработчик выхода
// 	http.ListenAndServe("0.0.0.0:8080", nil)
// }

// package main

// import (
// 	"net/http"

// 	"github.com/gorilla/mux"
// )

// func handleRequest() {
// 	r := mux.NewRouter() // Создаем новый маршрутизатор

// 	r.HandleFunc("/", home_page).Methods("GET")
// 	r.HandleFunc("/register", register_page).Methods("GET", "POST")
// 	r.HandleFunc("/login", login_page).Methods("GET", "POST")
// 	r.Handle("/protected", isAuthenticated(http.HandlerFunc(protectedPage)))
// 	r.Handle("/user", isAuthenticated(http.HandlerFunc(userProfile)))
// 	r.Handle("/meetings", isAuthenticated(http.HandlerFunc(meetingsPage))).Methods("GET")
// 	r.Handle("/cancel_meeting", isAuthenticated(http.HandlerFunc(cancelMeetingHandler))).Methods("POST")
// 	r.HandleFunc("/meetings/update", isAuthenticated(http.HandlerFunc(updateMeetingTimeHandler))).Methods("POST")
// 	r.HandleFunc("/meetings/delete", isAuthenticated(http.HandlerFunc(deleteMeetingHandler))).Methods("POST")
// 	r.HandleFunc("/logout", logout).Methods("POST") // Обработчик выхода

// 	http.ListenAndServe("0.0.0.0:8080", r) // Запускаем сервер с маршрутизатором
// }

// package main

// import (
// 	"net/http"

// 	"github.com/gorilla/mux"
// )

// func handleRequest() {
// 	r := mux.NewRouter() // Создаем новый маршрутизатор

// 	r.HandleFunc("/", home_page).Methods("GET")
// 	r.HandleFunc("/register", register_page).Methods("GET", "POST")
// 	r.HandleFunc("/login", login_page).Methods("GET", "POST")
// 	r.Handle("/protected", isAuthenticated(http.HandlerFunc(protectedPage)))
// 	r.Handle("/user", isAuthenticated(http.HandlerFunc(userProfile)))
// 	r.Handle("/meetings", isAuthenticated(http.HandlerFunc(meetingsPage))).Methods("GET")
// 	r.Handle("/cancel_meeting", isAuthenticated(http.HandlerFunc(cancelMeetingHandler))).Methods("POST")
// 	r.Handle("/meetings/update", isAuthenticated(http.HandlerFunc(updateMeetingTimeHandler))).Methods("POST")
// 	r.Handle("/meetings/delete", isAuthenticated(http.HandlerFunc(deleteMeetingHandler))).Methods("POST")
// 	r.HandleFunc("/logout", logout).Methods("POST") // Обработчик выхода

// 	http.ListenAndServe("0.0.0.0:8080", r) // Запускаем сервер с маршрутизатором
// }

// package main

// import (
// 	"net/http"

// 	"github.com/gorilla/mux"
// )

// func handleRequest() {
// 	r := mux.NewRouter() // Создаем новый маршрутизатор

// 	// Основные маршруты
// 	r.HandleFunc("/", home_page).Methods("GET")
// 	r.HandleFunc("/register", register_page).Methods("GET", "POST")
// 	r.HandleFunc("/login", loginHandler).Methods("GET", "POST") // Обработчик входа
// 	r.HandleFunc("/logout", logout).Methods("POST")             // Обработчик выхода

// 	// Защищенные маршруты
// 	r.Handle("/protected", isAuthenticated(http.HandlerFunc(protectedPage)))
// 	r.Handle("/user_profile", isAuthenticated(http.HandlerFunc(userProfile))).Methods("GET")                  // Профиль пользователя
// 	r.Handle("/meetings", isAuthenticated(http.HandlerFunc(meetingsPage))).Methods("GET")                     // Страница встреч
// 	r.Handle("/cancel_meeting", isAuthenticated(http.HandlerFunc(cancelMeetingHandler))).Methods("POST")      // Отмена встречи
// 	r.Handle("/meetings/update", isAuthenticated(http.HandlerFunc(updateMeetingTimeHandler))).Methods("POST") // Обновление времени встречи
// 	r.Handle("/meetings/delete", isAuthenticated(http.HandlerFunc(deleteMeetingHandler))).Methods("POST")     // Удаление встречи
// 	// Перенаправление на профиль пользователя после успешного входа
// 	r.HandleFunc("/login/success", func(w http.ResponseWriter, r *http.Request) {
// 		http.Redirect(w, r, "/user_profile", http.StatusFound) // Перенаправление на профиль пользователя
// 	}).Methods("GET")

// 	http.ListenAndServe("0.0.0.0:8080", r) // Запускаем сервер с маршрутизатором
// }

package main

import (
	"encoding/gob"
	"net/http"

	"github.com/gorilla/mux"
)

func init() {
	// Регистрация типа User для использования в сессиях
	gob.Register(User{})
}

func handleRequest() {
	r := mux.NewRouter() // Создаем новый маршрутизатор

	// Основные маршруты
	r.HandleFunc("/", home_page).Methods("GET")
	r.HandleFunc("/register", register_page).Methods("GET", "POST")
	r.HandleFunc("/login", loginHandler).Methods("GET", "POST") // Обработчик входа
	r.HandleFunc("/logout", logout).Methods("GET", "POST")      // Обработчик выхода

	// Защищенные маршруты
	r.Handle("/protected", isAuthenticated(http.HandlerFunc(protectedPage)))
	r.Handle("/user_profile", isAuthenticated(http.HandlerFunc(userProfile))).Methods("GET")                         // Профиль пользователя
	r.Handle("/meetings", isAuthenticated(http.HandlerFunc(meetingsPage))).Methods("GET", "POST")                    // Страница встреч
	r.Handle("/cancel_meeting", isAuthenticated(http.HandlerFunc(cancelMeetingHandler))).Methods("GET", "POST")      // Отмена встречи
	r.Handle("/meetings/update", isAuthenticated(http.HandlerFunc(updateMeetingTimeHandler))).Methods("GET", "POST") // Обновление времени встречи
	r.Handle("/meetings/delete", isAuthenticated(http.HandlerFunc(deleteMeetingHandler))).Methods("GET", "POST")     // Удаление встречи

	http.ListenAndServe("0.0.0.0:8080", r) // Запускаем сервер с маршрутизатором
}
