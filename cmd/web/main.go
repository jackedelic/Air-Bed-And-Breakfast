package main

import (
	"net/http"

	"github.com/jackedelic/go-overview-trevor-sawler/pkg/handlers"
)

func main() {
	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/about", handlers.About)
	_ = http.ListenAndServe(":8080", nil)
}
