package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/chriscow/strucim/internal/messages"

	"github.com/urfave/cli/v2"
)

type prediction struct {
	Name string  `json:"name"`
	ID   int     `json:"id"`
	Prob float64 `json:"prob"`
}

// Identify reads in the file given as an argument.
func Identify(ctx *cli.Context) error {

	if err := validateArgs(ctx); err != nil {
		return err
	}

	filename := ctx.Args().Get(0)
	endpoint := os.Getenv("LOCATOR_ENDPOINT")
	url := fmt.Sprintf("%s/v1/identify", endpoint)

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

	data := make([]prediction, 10)
	if err := json.Unmarshal(pred, &data); err != nil {
		return err
	}

	sort.Slice(data, func(i, j int) bool {
		if data[i].Prob > data[j].Prob {
			return true
		}

		return false
	})

	fmt.Printf("%s (%s%% probability)\n", data[0].Name, strconv.FormatFloat(data[0].Prob*100, 'f', 2, 64))
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
