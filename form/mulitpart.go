package form

import (
	"io"
	"io/ioutil"
	"net/http"
)

// MultipartItem .
type MultipartItem struct {
	Value       []byte
	ContentType string
}

// ExtractMultipartItems is a helper to extract multipart encoded form data
func ExtractMultipartItems(r *http.Request) (map[string]*MultipartItem, error) {
	var items = make(map[string]*MultipartItem)

	reader, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}

DONE:
	for {
		err = func() error {
			part, err := reader.NextPart()
			if err != nil {
				return err
			}
			if part == nil {
				return io.EOF
			}
			defer part.Close()

			b, err := ioutil.ReadAll(part)
			if err != nil {
				return err
			}

			items[part.FormName()] = &MultipartItem{
				Value:       b,
				ContentType: part.Header.Get("Content-Type"),
			}

			return nil
		}()

		if err == io.EOF {
			break DONE
		}
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}
