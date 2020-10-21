package main

import (
	"log"
	"net/http"
	"os"

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

	r.HandleFunc("/v1/identify", uploadHandler).
		Methods("POST")

	log.Fatal(http.ListenAndServe(os.Getenv("GATEWAY_ENDPOINT"), r))
}
