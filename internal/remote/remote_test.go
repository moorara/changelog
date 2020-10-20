package remote

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

type MockResponse struct {
	Method       string
	Path         string
	StatusCode   int
	ResponseBody string
}

func createMockHTTPServer(mocks ...MockResponse) *httptest.Server {
	r := mux.NewRouter()
	for _, m := range mocks {
		r.Methods(m.Method).Path(m.Path).HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(m.StatusCode)
			_, _ = io.WriteString(w, m.ResponseBody)
		})
	}

	return httptest.NewServer(r)
}
