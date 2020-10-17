package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// Bucket describes a Cloud Storage bucket
type Bucket struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

// GetBucket returns a Bucket structure if the bucket exists
func GetBucket(name string) (*Bucket, error) {
	ctx := context.Background()
	credsPath := os.Getenv("GCP_CREDENTIALS_PATH")

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credsPath))
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close() // close the client when the function ends

	// set a timeout of 10 seconds
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	attrs, err := client.Bucket(name).Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("Bucket(%q).Attrs: %v", name, err)
	}

	return &Bucket{
		Name:     attrs.Name,
		Location: attrs.Location,
	}, nil
}

// CreateBucket creates a bucket in GCP
func CreateBucket(name string) (*Bucket, error) {
	ctx := context.Background()

	projectID := os.Getenv("GCP_PROJECT_ID")
	credsPath := os.Getenv("GCP_CREDENTIALS_PATH")

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credsPath))
	if err != nil {
		return nil, err
	}
	defer client.Close() // close the client when the function ends

	// gets us a bucket handle, even if it doesn't exist yet
	bkt := client.Bucket(name)

	if err := bkt.Create(ctx, projectID, nil); err != nil {
		return nil, err
	}

	attrs, err := bkt.Attrs(ctx) // bucket attributes
	if err != nil {
		return nil, err
	}

	b := &Bucket{Name: name, Location: attrs.Location}

	return b, nil // no error
}

// Store puts a file into the bucket
func (b *Bucket) Store(filename string, blob io.Reader) error {
	ctx := context.Background()

	credsPath := os.Getenv("GCP_CREDENTIALS_PATH")

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credsPath))
	if err != nil {
		return err
	}
	defer client.Close()

	// gets us a bucket writer
	writer := client.Bucket(b.Name).Object(filename).NewWriter(ctx)

	// copy the bytes of the file to the bucket
	if _, err := io.Copy(writer, blob); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("writer.Close: %v", err)
	}

	return nil
}
