package main

import (
	"log"
	"net/http"
	"os"

	"github.com/chriscow/strucim/internal/gateway"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

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

// Check for environment variable values.  Set defaults if they aren't set
func verifyEnvironment() error {

	if os.Getenv("GCP_CREDENTIALS_PATH") == "" {
		// set default
		os.Setenv("GCP_CREDENTIALS_PATH", "./.secrets/strucim-gateway-keys.json")
	}

	_, err := os.Stat(os.Getenv("GCP_CREDENTIALS_PATH"))
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Google credentials file not found")
		}
		return err
	}

	if os.Getenv("GATEWAY_ENDPOINT") == "" {
		os.Setenv("GATEWAY_ENDPOINT", ":8080")
	}

	if os.Getenv("IDENTIFY_BUCKET") == "" {
		os.Setenv("IDENTIFY_BUCKET", "pointcloud-identification")
	}

	if os.Getenv("IDENTIFY_POINTCLOUD_TOPIC") == "" {
		os.Setenv("IDENTIFY_POINTCLOUD_TOPIC", "identify-request")
	}

	if os.Getenv("IDENTIFY_POINTCLOUD_STATUS_TOPIC") == "" {
		os.Setenv("IDENTIFY_POINTCLOUD_STATUS_TOPIC", "identify-status")
	}

	return nil
}
