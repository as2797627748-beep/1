package main

import (
	"log"
	"net/http"
	"os"

	"autocode-platform/internal/app"
)

func main() {
	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	server := app.New()
	log.Printf("autocode platform listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}
