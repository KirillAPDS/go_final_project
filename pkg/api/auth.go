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

func signinHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]string{"error": "Request error"})
		return
	}

	pass := os.Getenv("TODO_PASSWORD")
	if pass == "" || req.Password != pass {
		writeJSON(w, map[string]string{"error": "Incorrect password"})
		return
	}

	claims := jwt.MapClaims{
		"hash": fmt.Sprintf("%x", pass),
		"exp":  time.Now().Add(8 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString(jwtSecret)

	writeJSON(w, map[string]string{"token": signed})
}
