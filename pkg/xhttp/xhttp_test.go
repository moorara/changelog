package xhttp

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientError(t *testing.T) {
	req, err := http.NewRequest("GET", "/item", nil)
	assert.NoError(t, err)

	tests := []struct {
		name               string
		resp               *http.Response
		expectedError      string
		expectedStatusCode int
	}{
		{
			name: "OK",
			resp: &http.Response{
				StatusCode: 500,
				Body:       ioutil.NopCloser(strings.NewReader("internal server error")),
				Request:    req,
			},
			expectedError:      "GET /item 500: internal server error",
			expectedStatusCode: 500,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := NewClientError(tc.resp)

			assert.Equal(t, tc.expectedError, err.Error())
			assert.Equal(t, tc.expectedStatusCode, err.StatusCode())
		})
	}
}
