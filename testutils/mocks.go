package testutils

import "net/http"

// MockURLGetter - allows stubbing out any external http calls via the http.Get,
// urlfetch.Get or other methods that match the interface
type MockURLGetter struct {
	Err    error
	Body   string
	Status int
}

func (u MockURLGetter) Get(url string) (*http.Response, error) {
	if u.Err != nil {
		return nil, u.Err
	}
	r := http.Response{Body: MockReadCloser{err: u.Err, data: []byte(u.Body)}}
	r.StatusCode = u.Status
	return &r, nil
}

// MockReadCloser - Used within the mockURLGetter to stub out response data.
type MockReadCloser struct {
	err  error
	data []byte
}

func (m MockReadCloser) Read(data []byte) (int, error) {
	if m.err != nil {
		return 0, m.err
	}

	copy(data, m.data)
	return len(m.data), nil
}

func (m MockReadCloser) Close() error {
	return nil
}
