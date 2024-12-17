package main

import (
	"net/http"
	"strconv"
	"time"
)

type Meeting struct {
	ID        int
	UserID    int
	Date      time.Time
	Cancelled bool
	CreatedAt time.Time // Время записи на встречу
}

func createMeeting(userID int, date time.Time) error {
	createdAt := time.Now() // Получаем текущее время
	_, err := db.Exec("INSERT INTO meetings (user_id, date, created_at) VALUES ($1, $2, $3)", userID, date, createdAt)
	return err
}

func getMeetings(userID int) ([]Meeting, error) {
	rows, err := db.Query("SELECT id, user_id, date, cancelled, created_at FROM meetings WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var meetings []Meeting
	for rows.Next() {
		var meeting Meeting
		if err := rows.Scan(&meeting.ID, &meeting.UserID, &meeting.Date, &meeting.Cancelled, &meeting.CreatedAt); err != nil {
			return nil, err
		}
		meetings = append(meetings, meeting)
	}
	return meetings, nil
}

func cancelMeeting(meetingID int) error {
	_, err := db.Exec("UPDATE meetings SET cancelled = TRUE WHERE id = $1", meetingID)
	return err
}

func cancelMeetingHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	_, ok := session.Values["user"].(User) // Игнорируем переменную user
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	meetingIDStr := r.URL.Query().Get("id")
	meetingID, err := strconv.Atoi(meetingIDStr) // Преобразование строки в int
	if err != nil {
		http.Error(w, "Неверный ID встречи", http.StatusBadRequest)
		return
	}

	if err := cancelMeeting(meetingID); err != nil {
		http.Error(w, "Ошибка при отмене встречи", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/meetings", http.StatusFound)
}

func updateMeetingTimeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	_, ok := session.Values["user"].(User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	meetingIDStr := r.URL.Query().Get("id")
	meetingID, err := strconv.Atoi(meetingIDStr) // Преобразование строки в int
	if err != nil {
		http.Error(w, "Неверный ID встречи", http.StatusBadRequest)
		return
	}

	// Здесь вы можете получить новое время встречи из запроса
	newTimeStr := r.URL.Query().Get("new_time")
	newTime, err := time.Parse("2006-01-02 15:04", newTimeStr) // Пример формата времени
	if err != nil {
		http.Error(w, "Неверный формат времени", http.StatusBadRequest)
		return
	}

	// Логика для обновления времени встречи
	if err := updateMeetingTime(meetingID, newTime); err != nil {
		http.Error(w, "Ошибка при изменении времени встречи", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/meetings", http.StatusFound)
}

func deleteMeetingHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	_, ok := session.Values["user"].(User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	meetingIDStr := r.URL.Query().Get("id")
	meetingID, err := strconv.Atoi(meetingIDStr) // Преобразование строки в int
	if err != nil {
		http.Error(w, "Неверный ID встречи", http.StatusBadRequest)
		return
	}

	// Логика для удаления встречи
	if err := deleteMeeting(meetingID); err != nil {
		http.Error(w, "Ошибка при удалении встречи", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/meetings", http.StatusFound)
}
