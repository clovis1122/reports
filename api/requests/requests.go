package requests

import (
	"io/ioutil"
	"net/http"
)

var client = http.Client{}

// Option represents a way to configure the http request.
type Option func(*http.Request)

// SetBasicAuth sets the username and password of a request.
func SetBasicAuth(username, password string) Option {
	return func(req *http.Request) {
		req.SetBasicAuth(username, password)
	}
}

// SetClient changes the internal Client instance. Useful for tests.
func SetClient(cl http.Client) {
	client = cl
}

// Get executes a request against the given URL.
func Get(url string, options ...Option) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return []byte{}, err
	}

	for i := range options {
		options[i](req)
	}

	res, err := client.Do(req)

	if err != nil {
		return []byte{}, err
	}

	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}
