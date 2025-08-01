package api

import (
	"encoding/json"
	"fmt"
	"github.com/KirillAPDS/go_final_project/pkg/db"
	"net/http"
	"time"
)

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
		writeJSON(w, map[string]string{"error": "ID is not specified"})
		return
	}
	err := db.DeleteTask(id)
	if err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, map[string]string{})
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSON(w, map[string]string{"error": "Incorrect JSON"})
		return
	}

	if task.Title == "" {
		writeJSON(w, map[string]string{"error": "Task title is not specified"})
		return
	}

	if err := checkAndFixDate(&task); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, map[string]string{"id": formatID(id)})
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, map[string]string{"error": "ID is not specified"})
		return
	}
	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, map[string]string{"error": "Task not found"})
		return
	}
	writeJSON(w, task)
}

func editTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSON(w, map[string]string{"error": "Incorrect JSON"})
		return
	}

	if task.ID == "" || task.Title == "" {
		writeJSON(w, map[string]string{"error": "ID or title is incorrect"})
		return
	}

	if err := checkAndFixDate(&task); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	err := db.UpdateTask(&task)
	if err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, map[string]string{})
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(data)
}

func formatID(id int64) string {
	return fmt.Sprintf("%d", id)
}

func doneHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, map[string]string{"error": "ID is not specified"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, map[string]string{"error": "Task not found"})
		return
	}

	if task.Repeat == "" {
		err = db.DeleteTask(id)
		if err != nil {
			writeJSON(w, map[string]string{"error": err.Error()})
			return
		}
	} else {
		next, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			writeJSON(w, map[string]string{"error": err.Error()})
			return
		}
		err = db.UpdateDate(id, next)
		if err != nil {
			writeJSON(w, map[string]string{"error": err.Error()})
			return
		}
	}

	writeJSON(w, map[string]string{})
}

func checkAndFixDate(task *db.Task) error {
	const layout = "20060102"
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

