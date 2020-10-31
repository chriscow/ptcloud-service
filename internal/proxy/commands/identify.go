package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/chriscow/strucim/internal/messages"

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

	// Encode the file and post it.  The server will store it in Cloud Storage
	encoded, _ := encodeFile(filename, endpoint, url)

	req, err := json.Marshal(&messages.IdentifyRequest{File: encoded})
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(req))
	if err != nil {
		return err
	}

	pred, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("Server return an error: %s", string(pred))
	}

	log.Println("Received message:", string(pred))

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
