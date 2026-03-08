package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := getEnv("PORT", "8080")
	message := getEnv("MESSAGE", "hello from watchforge test server")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("request received:", r.Method, r.URL.Path)
		fmt.Fprintf(w, "%s\n", message)
	})

	http.HandleFunc("/time", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().Format(time.RFC3339)
		fmt.Fprintf(w, "server time: %s\n", now)
	})

	log.Println("server starting on port", port)
	log.Println("message:", message)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
