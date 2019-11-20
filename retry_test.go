package retryrequest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// https://github.com/hashicorp/go-retryablehttp/blob/master/client_test.go#L132

// Always returns 500
var handlerAlways500 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
})

// Returns 500 until after X requests are made
type handler500 struct {
	After    int
	requests int
}

func (h *handler500) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.requests++
	if h.requests <= h.After {
		w.WriteHeader(500)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func handler500Until(after int) http.Handler {
	return &handler500{After: after}
}

func TestClient(t *testing.T) {

	// Mock server which always responds 500.
	ts := httptest.NewServer(handler500Until(2))
	defer ts.Close()

	// Create the client. Use short retry windows so we fail faster.
	client := http.DefaultClient

	// Create the request
	req, err := http.NewRequest("POST", ts.URL, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("expected giving up error, got: %#v", err)
	}

	if resp.StatusCode != 500 {
		t.Fatalf("Invalid response code: %d", resp.StatusCode)
	}
}

// func TestHandler(t *testing.T) {
//
// 	rr := httptest.NewRecorder()
// 	rr := httptest.NewRecorder()
//
// }
