package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

// Tiny echo server. Usage: go run backends.go <port> <name>
func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: go run backends.go <port> <name>")
	}
	port, _ := strconv.Atoi(os.Args[1])
	name := os.Args[2]

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%s:%d] %s %s\n", name, port, r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "hello from %s (port %d)\npath=%s\n", name, port, r.URL.Path)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	log.Printf("%s listening on :%d", name, port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
