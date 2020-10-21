package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"strucim/proxy/messages"

	"github.com/urfave/cli/v2"
)

// Identify reads in the file given as an argument.
func Identify(ctx *cli.Context) error {

	if err := validateArgs(ctx); err != nil {
		return err
	}

	filename := ctx.Args().Get(0)
	endpoint := os.Getenv("LOCATOR_ENDPOINT")
	url := fmt.Sprintf("http://%s/v1/identify", endpoint)

	encoded, _ := encodeFile(filename, endpoint, url)

	req, err := json.Marshal(&messages.IdentifyRequest{File: encoded})
	if err != nil {
		return cli.Exit(err.Error(), -1)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(req))
	if err != nil {
		return cli.Exit(err.Error(), -1)
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Server return an error: %s", string(body))
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
