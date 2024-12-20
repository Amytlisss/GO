package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Meeting struct {
	ID        int
	UserID    int
	Date      time.Time
	Cancelled bool
	CreatedAt time.Time
}

func createMeeting(userID int, date time.Time) error {
	var exists bool
	if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("пользователь с ID %d не существует", userID)
	}

	_, err := db.Exec("INSERT INTO meetings (user_id, date, created_at) VALUES ($1, $2, $3)", userID, date, time.Now())
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

func handleSession(w http.ResponseWriter, r *http.Request) (User, bool) {
	session, _ := store.Get(r, "session-name")
	user, ok := session.Values["user"].(User)
	return user, ok
}

func cancelMeetingHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := handleSession(w, r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	meetingID, err := strconv.Atoi(r.URL.Query().Get("id"))
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

func updateMeeting(meetingID int, newDate time.Time) error {
	_, err := db.Exec("UPDATE meetings SET date = $1 WHERE id = $2", newDate, meetingID)
	return err
}

func getMeetingByID(meetingID int) (Meeting, error) {
	var meeting Meeting
	err := db.QueryRow("SELECT id, user_id, date, cancelled FROM meetings WHERE id = $1 AND cancelled = FALSE", meetingID).Scan(&meeting.ID, &meeting.UserID, &meeting.Date, &meeting.Cancelled)
	if err != nil {
		return Meeting{}, err
	}
	return meeting, nil
}

func editMeetingPageHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := handleSession(w, r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	meetingID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Неверный ID встречи", http.StatusBadRequest)
		return
	}

	meetings, err := getMeetings(user.ID)
	if err != nil {
		http.Error(w, "Ошибка при получении встреч", http.StatusInternalServerError)
		return
	}

	var meeting Meeting
	for _, m := range meetings {
		if m.ID == meetingID {
			meeting = m
			break
		}
	}

	if meeting.ID == 0 {
		meeting, err = getMeetingByID(meetingID)
		if err != nil {
			http.Error(w, "Ошибка при получении встречи", http.StatusInternalServerError)
			return
		}
	}

	tmpl, err := template.ParseFiles("templates/edit_meeting.html")
	if err != nil {
		http.Error(w, "Ошибка при загрузке шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, struct {
		User    User
		Meeting Meeting
	}{User: user, Meeting: meeting}); err != nil {
		http.Error(w, "Ошибка при выполнении шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func editMeetingHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := handleSession(w, r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	meetingID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Неверный ID встречи", http.StatusBadRequest)
		return
	}

	// Получаем встречу по ID
	meeting, err := getMeetingByID(meetingID)
	if err != nil {
		http.Error(w, "Ошибка при получении встречи", http.StatusInternalServerError)
		return
	}

	// Проверка прав доступа: пользователь должен быть создателем встречи или администратором
	if meeting.UserID != user.ID && user.Role != "admin" {
		http.Error(w, "У вас нет прав для редактирования этой встречи", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		newDateStr := strings.TrimSpace(r.FormValue("date"))
		newTimeStr := strings.TrimSpace(r.FormValue("time"))

		if newDateStr == "" || newTimeStr == "" {
			http.Error(w, "Дата и время не могут быть пустыми", http.StatusBadRequest)
			return
		}

		newDateTime, err := time.Parse("2006-01-02 15:04", newDateStr+" "+newTimeStr)
		if err != nil {
			http.Error(w, "Неверный формат времени", http.StatusBadRequest)
			return
		}

		if err := updateMeeting(meetingID, newDateTime); err != nil {
			fmt.Println("Ошибка при обновлении встречи:", err) // Логирование ошибки
			http.Error(w, "Ошибка при изменении времени встречи", http.StatusInternalServerError)
			return
		}

		// Перенаправление в зависимости от роли пользователя
		if user.Role == "admin" {
			http.Redirect(w, r, "/admin_page", http.StatusFound) // Замените на реальный путь для администраторов
		} else {
			http.Redirect(w, r, "/meetings", http.StatusFound)
		}
		return
	}

	// Здесь мы уже имеем переменную meeting, полученную ранее
	tmpl, err := template.ParseFiles("templates/edit_meeting.html")
	if err != nil {
		http.Error(w, "Ошибка при загрузке шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, struct {
		User    User
		Meeting Meeting
	}{User: user, Meeting: meeting}); err != nil {
		http.Error(w, "Ошибка при выполнении шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
