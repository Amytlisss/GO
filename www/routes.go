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
