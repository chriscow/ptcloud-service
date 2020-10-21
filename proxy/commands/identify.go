package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
)

type idRequest struct {
	File string `json:"file"`
}

type idResponse struct {
	Bucket   string `json:"bucket"`
	Filename string `json:"filename"`
	Status   string `json:"status"`
	Result   string `json:"result"`
}

// Identify reads in the file given as an argument.
func Identify(ctx *cli.Context) error {

	if err := validateArgs(ctx); err != nil {
		return err
	}

	filename := ctx.Args().Get(0)
	endpoint := os.Getenv("LOCATOR_ENDPOINT")
	url := fmt.Sprintf("http://%s/v1/identify", endpoint)

	encoded, _ := encodeFile(filename, endpoint, url)

	req, err := json.Marshal(&idRequest{File: encoded})
	if err != nil {
		return cli.Exit(err.Error(), -1)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(req))
	if err != nil {
		return cli.Exit(err.Error(), -1)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("Server return an error: %s", string(body))
	}

	respJSON := idResponse{}
	if err := json.Unmarshal(body, &respJSON); err != nil {
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

func encodeFile(filename, endpoint, uri string) (string, error) {

	csv, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("Failed to read file: %v", err)
	}

	return base64.StdEncoding.EncodeToString(csv), nil
}
