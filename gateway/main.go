package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"

	"github.com/gorilla/mux"

	"github.com/joho/godotenv"
)

type dashboard struct {
	Title string
	Host  string
}

func main() {
	godotenv.Load()
	r := mux.NewRouter()

	handler := http.HandlerFunc(uploadHandler)
	logger := handlers.CombinedLoggingHandler(os.Stdout, handler)

	path := "/v1/identify"
	r.HandleFunc(path, logger.ServeHTTP).
		Methods("POST")

	r.HandleFunc(path, wsHandler).
		Methods("GET")

	log.Fatal(http.ListenAndServe(os.Getenv("GATEWAY_ENDPOINT"), r))
}
