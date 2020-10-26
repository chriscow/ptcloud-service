package gateway

import (
	"context"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// Store puts a file into the bucket
func storeFile(bucket, filename string, blob io.Reader) error {
	ctx := context.Background()

	credsPath := os.Getenv("GCP_CREDENTIALS_PATH")

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credsPath))
	if err != nil {
		return err
	}
	defer client.Close()

	// gets us a bucket writer
	writer := client.Bucket(bucket).Object(filename).NewWriter(ctx)

	// copy the bytes of the file to the bucket
	if _, err := io.Copy(writer, blob); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("writer.Close: %v", err)
	}

	return nil
}
