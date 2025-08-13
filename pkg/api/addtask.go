package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/KirillAPDS/go_final_project/pkg/db"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type IDResponse struct {
	ID string `json:"id"`
}

const layout = "20060102"

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		editTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSONStatus(w, http.StatusBadRequest, ErrorResponse{Error: "ID is not specified"})
		return
	}
	err := db.DeleteTask(id)
	if err != nil {
		writeJSONStatus(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, map[string]string{})
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSONStatus(w, http.StatusBadRequest, ErrorResponse{Error: "Incorrect JSON"})
		return
	}

	if task.Title == "" {
		writeJSONStatus(w, http.StatusBadRequest, ErrorResponse{Error: "Task title is not specified"})
		return
	}

	if err := checkAndFixDate(&task); err != nil {
		writeJSONStatus(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJSONStatus(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, IDResponse{ID: formatID(id)})
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSONStatus(w, http.StatusBadRequest, ErrorResponse{Error: "ID is not specified"})
		return
	}
	task, err := db.GetTask(id)
	if err != nil {
		writeJSONStatus(w, http.StatusBadRequest, ErrorResponse{Error: "Task not found"})
		return
	}
	writeJSON(w, task)
}

func editTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSONStatus(w, http.StatusBadRequest, ErrorResponse{Error: "Incorrect JSON"})
		return
	}

	if task.ID == "" || task.Title == "" {
		writeJSONStatus(w, http.StatusBadRequest, ErrorResponse{Error: "ID or title is incorrect"})
		return
	}

	if err := checkAndFixDate(&task); err != nil {
		writeJSONStatus(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	err := db.UpdateTask(&task)
	if err != nil {
		writeJSONStatus(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, map[string]string{})
}

func writeJSONStatus(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode JSON", http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, data any) {
	writeJSONStatus(w, http.StatusOK, data)
}

func formatID(id int64) string {
	return fmt.Sprintf("%d", id)
}

func doneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSONStatus(w, http.StatusBadRequest, ErrorResponse{Error: "ID is not specified"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSONStatus(w, http.StatusBadRequest, ErrorResponse{Error: "Task not found"})
		return
	}

	if task.Repeat == "" {
		err = db.DeleteTask(id)
		if err != nil {
			writeJSONStatus(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	} else {
		next, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			writeJSONStatus(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		err = db.UpdateDate(id, next)
		if err != nil {
			writeJSONStatus(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	writeJSON(w, map[string]string{})
}

func checkAndFixDate(task *db.Task) error {
	now := time.Now()

	// Подставить сегодняшнюю дату, если не указана
	if task.Date == "" {
		task.Date = now.Format(layout)
	}

	// Проверка формата даты
	startDate, err := time.Parse(layout, task.Date)
	if err != nil {
		return fmt.Errorf("Incorrect date format: %w", err)
	}

	// Проверить и пересчитать повтор
	if task.Repeat != "" {
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("Repeat rule error: %w", err)
		}
		// Если дата в прошлом — заменить на next
		if afterNow(now, startDate) {
			task.Date = next
		}
	} else {
		// Если повтора нет, но дата в прошлом — заменим на сегодня
		if afterNow(now, startDate) {
			task.Date = now.Format(layout)
		}
	}

	return nil
}
