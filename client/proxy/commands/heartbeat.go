package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
)

// Heartbeat tests that we can connect and authenticate with Google Cloud
func Heartbeat(ctx *cli.Context) error {
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
