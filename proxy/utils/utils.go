package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// NewUploadFileRequest creates a multipart form request with the contents
// of the file embedded
func NewUploadFileRequest(uri, key, filename string) (*http.Request, error) {

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %s", filename)
	}
	defer file.Close()

	// We are going to create a multipart form to upload the file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(key, file.Name())
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}

	// This sets the content type with the part boundary for the file
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// fmt.Println("Content-Length", req.ContentLength, req.Body.Bytes())

	return req, err
}
