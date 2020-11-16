package github

import (
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gorilla/mux"
)

func parseGitHubTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	return t
}

func parseGitHubTimePtr(s string) *time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	return &t
}

type MockResponse struct {
	Method             string
	Path               string
	ResponseStatusCode int
	ResponseHeader     http.Header
	ResponseBody       string
}

func newHTTPTestServer(mocks ...MockResponse) *httptest.Server {
	r := mux.NewRouter()
	for _, m := range mocks {
		m := m
		r.Methods(m.Method).Path(m.Path).HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			for k, vals := range m.ResponseHeader {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(m.ResponseStatusCode)
			_, _ = io.WriteString(w, m.ResponseBody)
		})
	}

	return httptest.NewServer(r)
}
