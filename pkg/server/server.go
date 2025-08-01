package server

import (
	"fmt"
	"net/http"

	"github.com/KirillAPDS/go_final_project/pkg/api"
)

func Run(port string) error {
	api.Init()

	webDir := "web"
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Starting server at http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}
