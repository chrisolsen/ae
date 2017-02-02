package attachment

import (
	"errors"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/file"
)

// NewWriter returns a writer that writes to GCS.
func NewWriter(c context.Context, filename, contentType string) (io.WriteCloser, error) {
	if len(filename) == 0 {
		return nil, errors.New("filename is required")
	}

	if len(contentType) == 0 {
		return nil, errors.New("contentType is required")
	}

	bucketName, err := file.DefaultBucketName(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %v", err)
	}

	client, err := storage.NewClient(c)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	wc := bucket.Object(filename).NewWriter(c)
	wc.ContentType = contentType
	wc.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	wc.CacheControl = "public, max-age=86400"
	wc.Metadata = map[string]string{
		"x-goog-project-id": appengine.AppID(c),
		"x-goog-acl":        "public-read",
	}

	return wc, nil
}

// NewReader returns a reader that reads the image from GCS
func NewReader(c context.Context, filename string) (io.ReadCloser, string, error) {
	if len(filename) == 0 {
		return nil, "", errors.New("filename is required")
	}

	bucketName, err := file.DefaultBucketName(c)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get bucketName: %v", bucketName)
	}

	client, err := storage.NewClient(c)
	if err != nil {
		return nil, "", fmt.Errorf("failed creating client: %v", err)
	}
	defer client.Close()

	file := client.Bucket(bucketName).Object(filename)
	attrs, err := file.Attrs(c)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get image attrs: %v", err)
	}
	reader, err := file.NewReader(c)
	return reader, attrs.ContentType, err
}
