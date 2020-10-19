package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
)

type identifyRequest struct {
	File     string `json:"file"`
	Callback string `json:"callback"`
}

// Identify reads in the file given as an argument.
func Identify(ctx *cli.Context) error {

	if err := validateArgs(ctx); err != nil {
		return err
	}

	filename := ctx.Args().Get(0)
	endpoint := os.Getenv("LOCATOR_ENDPOINT")
	url := fmt.Sprintf("http://%s/v1/identify", endpoint)

	req, _ := newUploadRequest(filename, endpoint, url)

	resp, _ := http.Post(url, "application/json", req)

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Server return an error: %s", string(body))
	}

	type idRespJSON struct {
		Status string `json:"status"`
	}
	respJSON := &idRespJSON{}
	if err := json.NewDecoder(resp.Body).Decode(&respJSON); err != nil {
		return fmt.Errorf("Failed to decode response body: %v", err)
	}

	return nil
}

func validateArgs(ctx *cli.Context) error {
	if ctx.Args().Len() == 0 {
		return cli.Exit("You forgot to pass in the file name", 1)
	}

	if ctx.Args().Len() > 1 {
		return cli.Exit("Too many arguments were passed in. Just pass the filename", 1)
	}

	return nil
}

func newUploadRequest(filename, endpoint, uri string) (io.Reader, error) {

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file: %v", err)
	}

	encoded := base64.StdEncoding.EncodeToString(file)

	req := &identifyRequest{
		File: encoded,
	}

	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}
