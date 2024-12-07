package main

import (
	"fmt"
	"net/http"
	"time"
)

type Info struct {
	ID        int       `json:"id"`
	EventTime time.Time `json:"event_time"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "В kданном проекте реализовано поднятие БД ClickHouse и скрипта на Go в Docker.")
		fmt.Fprintln(w, "ClickHouse тут  http://localhost:8123/play")
	})

	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", nil)
}
