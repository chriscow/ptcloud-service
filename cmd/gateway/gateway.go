package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/chriscow/strucim/internal/gateway"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

/*
	TODO:
	File upload service
	Inference service
	Notification service
*/

type dashboard struct {
	Title string
	Host  string
}

func main() {
	godotenv.Load()

	if err := verifyEnvironment(); err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	handler := http.HandlerFunc(gateway.UploadHandler)
	logger := handlers.CombinedLoggingHandler(os.Stdout, handler)

	path := "/v1/identify"
	r.HandleFunc(path, logger.ServeHTTP).
		Methods("POST")

	r.HandleFunc(path, gateway.NotifyHandler).
		Methods("GET")

	log.Fatal(http.ListenAndServe(os.Getenv("GATEWAY_ENDPOINT"), r))
}

func verifyEnvironment() error {
	_, err := os.Stat(os.Getenv("GCP_CREDENTIALS_PATH"))
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Google file not found")
		}

		return err
	}

	if os.Getenv("IDENTIFY_BUCKET") == "" {
		return errors.New("IDENTIFY_BUCKET is not defined")
	}

	if os.Getenv("IDENTIFY_POINTCLOUD_TOPIC") == "" {
		return errors.New("IDENTIFY_POINTCLOUD_TOPIC is not defined")
	}

	if os.Getenv("IDENTIFY_POINTCLOUD_STATUS_TOPIC") == "" {
		return errors.New("IDENTIFY_POINTCLOUD_STATUS_TOPIC is not defined")
	}

	return err
}
