package image

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/chrisolsen/ae/attachment"
	"github.com/disintegration/imaging"
	"golang.org/x/net/context"
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
	reader, contentType, err := attachment.NewReader(c, name)
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
	writer, err := attachment.NewWriter(c, sizeName, contentType)
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
