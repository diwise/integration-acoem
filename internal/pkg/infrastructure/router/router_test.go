package router

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/matryer/is"
	"github.com/rs/zerolog/log"
)

func TestThatHealthEndpointReturns204(t *testing.T) {
	is := is.New(t)

	r := newRouterForTesting()
	ts := httptest.NewServer(r.router)
	defer ts.Close()

	resp, _ := testRequest(is, ts, "GET", "/health", nil)

	is.Equal(resp.StatusCode, http.StatusNoContent) // health endpoint status code not ok
}

func newRouterForTesting() *routerStruct {
	r := chi.NewRouter()
	log := log.Logger

	return setupRouter(r, log)
}

func testRequest(is *is.I, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, _ := http.NewRequest(method, ts.URL+path, body)
	resp, _ := http.DefaultClient.Do(req)
	respBody, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	return resp, string(respBody)
}
