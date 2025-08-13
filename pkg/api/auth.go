package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("todo-jwt-secret")
var todoPassword string

func InitAuth() {
	todoPassword = os.Getenv("TODO_PASSWORD")
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]string{"error": "Request error"})
		return
	}

	if todoPassword == "" || req.Password != todoPassword {
		writeJSON(w, map[string]string{"error": "Incorrect password"})
		return
	}

	claims := jwt.MapClaims{
		"hash": fmt.Sprintf("%x", todoPassword),
		"exp":  time.Now().Add(8 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString(jwtSecret)

	writeJSON(w, map[string]string{"token": signed})
}
