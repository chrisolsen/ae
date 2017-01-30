package image

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/disintegration/imaging"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/file"
	"google.golang.org/appengine/urlfetch"
)

// SizedURL points to the image, of the specified sizes, in Google Cloud Storage
func SizedURL(c context.Context, scheme, name string, width, height int) (string, error) {
	bucketName, err := file.DefaultBucketName(c)
	if err != nil {
		return "", fmt.Errorf("getting default bucket name: %v", err)
	}

	// check if image already exists
	sizeName := fmt.Sprintf("%s_w%dh%d", name, width, height)
	sizedURL := fmt.Sprintf("%s://storage.googleapis.com/%s/%s", scheme, bucketName, sizeName)
	client := urlfetch.Client(c)
	resp, err := client.Head(sizedURL)
	if err != nil {
		return "", fmt.Errorf("failed HEAD request: %v", err)
	}
	defer resp.Body.Close()
	// image already exists; return the url
	if resp.StatusCode == http.StatusOK {
		return sizedURL, nil
	}

	// fetch initial image and resize it
	reader, contentType, err := NewReader(c, name)
	if err != nil {
		return "", fmt.Errorf("reading fetched image: %v", err)
	}
	img, err := imaging.Decode(reader)
	if err != nil {
		return "", fmt.Errorf("failed image decoding: %v", err)
	}
	defer reader.Close()
	resized := imaging.Resize(img, width, height, imaging.Lanczos)

	// save image to storage
	writer, err := NewWriter(c, sizeName, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to create writer: %v", err)
	}
	defer writer.Close()
	format, ok := format(contentType)
	if !ok {
		return "", errors.New("Invalid image type")
	}
	err = imaging.Encode(writer, resized, format)
	if err != nil {
		return "", fmt.Errorf("failed to save resized image: %v", err)
	}

	return sizedURL, nil
}

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

func format(contentType string) (imaging.Format, bool) {
	formats := map[string]imaging.Format{
		"jpg":  imaging.JPEG,
		"jpeg": imaging.JPEG,
		"png":  imaging.PNG,
		"tif":  imaging.TIFF,
		"tiff": imaging.TIFF,
		"bmp":  imaging.BMP,
		"gif":  imaging.GIF,
	}

	f := strings.ToLower(strings.Split(contentType, "/")[1])
	format, ok := formats[f]
	return format, ok
}
