package xhttp

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// ClientError is a custom error type for errors happening when calling an HTTP endpoint.
type ClientError struct {
	message    string
	statusCode int
}

// NewClientError creates a new HTTP client error.
func NewClientError(resp *http.Response) *ClientError {
	var message string
	if resp.Body != nil {
		if b, e := ioutil.ReadAll(resp.Body); e == nil {
			message = fmt.Sprintf("%s %s %d: %s", resp.Request.Method, resp.Request.URL.Path, resp.StatusCode, string(b))
		}
	}

	return &ClientError{
		message:    message,
		statusCode: resp.StatusCode,
	}
}

func (e *ClientError) Error() string {
	return e.message
}

// StatusCode returns the status code of the HTTP response.
func (e *ClientError) StatusCode() int {
	return e.statusCode
}
