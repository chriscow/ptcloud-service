package gateway

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/chriscow/strucim/internal/messages"
)

func writeError(w http.ResponseWriter, status int, msg string, err error) {
	w.WriteHeader(status)
	log.Printf("%s: %v", msg, err)
	fmt.Fprintf(w, "%s: %v", msg, err)
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {

	req := messages.IdentifyRequest{}
	body, _ := ioutil.ReadAll(r.Body)

	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, 400, "Invalid request JSON", err)
		return
	}

	csvBytes, err := base64.StdEncoding.DecodeString(req.File)
	if err != nil {
		writeError(w, 400, "Failed to decode file: %s", err)
		return
	}

	sr := bytes.NewReader(csvBytes)
	bucket := os.Getenv("IDENTIFY_BUCKET")
	filename := fmt.Sprintf("pointcloud-%d", time.Now().Unix())

	if err := storeFile(bucket, filename, sr); err != nil {
		writeError(w, 500, "Failed to write to storage", err)
		return
	}

	msg := messages.IdentifyResponse{
		Bucket:   bucket,
		Filename: filename,
		Status:   "init",
	}

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		writeError(w, 500, "Failed to marshal pubsub message", err)
		return
	}

	log.Println("publishing status json to ", os.Getenv("IDENTIFY_POINTCLOUD_STATUS_TOPIC"))
	if err := Publish(os.Getenv("IDENTIFY_POINTCLOUD_STATUS_TOPIC"), msgJSON); err != nil {
		writeError(w, 500, "Failed to publish identify job", err)
		return
	}

	log.Println("publishing json to ", os.Getenv("IDENTIFY_POINTCLOUD_TOPIC"))
	// if err := Publish(os.Getenv("IDENTIFY_POINTCLOUD_TOPIC"), msgJSON); err != nil {
	// 	writeError(w, 500, "Failed to publish identify job", err)
	// 	return
	// }
	infUrl := fmt.Sprintf("%s?bucket=%s&filename=%s",
		os.Getenv("INFERENCE_SERVICE"),
		bucket, filename)

	resp, err := http.DefaultClient.Get(infUrl)

	if err != nil {
		writeError(w, 500, "Failed to post to cloud run", err)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		writeError(w, 500, "Failed to read body", err)
		return
	}

	if resp.StatusCode != 200 {
		writeError(w, resp.StatusCode, string(b), nil)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}
