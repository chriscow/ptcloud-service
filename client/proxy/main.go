package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"google.golang.org/api/idtoken"
)

type identifyResponse struct {
	Status string `json:"status"`
}

// makeIAPRequest makes a request to an application protected by Identity-Aware
// Proxy with the given audience.
func makeIAPRequest(w io.Writer, request *http.Request, audience string) error {
	// request, err := http.NewRequest("GET", "http://example.com", nil)
	// audience := "IAP_CLIENT_ID.apps.googleusercontent.com"
	ctx := context.Background()

	// client is a http.Client that automatically adds an "Authorization" header
	// to any requests made.
	client, err := idtoken.NewClient(ctx, audience)
	if err != nil {
		return fmt.Errorf("idtoken.NewClient: %v", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("client.Do: %v", err)
	}
	defer response.Body.Close()
	if _, err := io.Copy(w, response.Body); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}

	return nil
}

func connect(ctx *cli.Context) error {
	if ctx.Bool("verbose") {
		fmt.Println("connecting ...")
	}

	return nil
}

// hello tests that we can connect and authenticate with Google Cloud
func heartbeat(ctx *cli.Context) error {
	endpoint := os.Getenv("LOCATOR_ENDPOINT")
	resp, err := http.Get(fmt.Sprintf("http://%s/v1/heartbeat", endpoint))
	if err != nil {
		return fmt.Errorf("heartbeat: %v", err)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read: %v", err)
	}

	log.Println(string(bytes))
	return nil
}

// identify reads in the file given as an argument.
func identify(ctx *cli.Context) error {

	if ctx.Args().Len() == 0 {
		return cli.Exit("You forgot to pass in the file name", 1)
	}

	if ctx.Args().Len() > 1 {
		return cli.Exit("Too many arguments were passed in. Just pass the filename", 1)
	}

	filename := ctx.Args().Get(0)
	endpoint := os.Getenv("LOCATOR_ENDPOINT")
	uri := fmt.Sprintf("http://%s/v1/identify", endpoint)

	// get a multipart form post request with the file embedded
	request, err := uploadFileRequest(uri, "file", filename)
	if err != nil {
		return fmt.Errorf("failed to create multipart body: %v", err)
	}

	// post the file to the server
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("Error sending request to server: %v", err)
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Server return an error: %s", string(body))
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	fmt.Println("body:", string(bytes))

	ir := &identifyResponse{}
	if err := json.Unmarshal(bytes, &ir); err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return nil
}

func uploadFileRequest(uri, key, filename string) (*http.Request, error) {

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

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file", err)
	}

	// CLI App takes care of parsing commands, flags and printing help etc.
	app := &cli.App{
		Name:  "struCIM",
		Usage: "struCIM cloud interface for part identification",
	}

	// global flags i.e. proxy --verbose
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Enable detailed output",
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:    "identify",
			Aliases: []string{"id"},
			Usage:   "Identify a part from point cloud csv `FILE`",
			Action:  identify,
		},
		{
			Name:   "heartbeat",
			Usage:  "Tests connectivity and authentication with the cloud",
			Action: heartbeat,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
