package retryrequest

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// https://github.com/hashicorp/go-retryablehttp/blob/master/client_test.go#L132

// Always returns 500
var handlerAlways500 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
})

// Always timeout requests
var handlerAlwaysTimeout = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	// Block
	select {
	case <-r.Context().Done():
	case <-time.After(time.Hour):
	}
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

func TestRegularClient(t *testing.T) {

	// Mock server which always responds 500.
	ts := httptest.NewServer(handlerAlways500)
	defer ts.Close()

	// Create the client. Use short retry windows so we fail faster.
	client := http.Client{
		Timeout: time.Millisecond * 10,
	}

	// Create the request
	req, err := http.NewRequest("GET", ts.URL, nil)
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

func TestRetryClient(t *testing.T) {

	// Mock server which always responds 500.
	ts := httptest.NewServer(handler500Until(2))
	defer ts.Close()

	// Create the client. Use short retry windows so we fail faster.
	client := &http.Client{
		Timeout: time.Millisecond * 10,
	}

	// Create the request
	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	resp, err := Do(client, req, 3, time.Millisecond)
	// resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("unexpected error, got: %#v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Invalid response code: %d", resp.StatusCode)
	}
}

func TestRetryTimeout(t *testing.T) {

	// Mock server which always responds 500.
	ts := httptest.NewServer(handlerAlwaysTimeout)
	defer ts.Close()

	// Create the client. Use short retry windows so we fail faster.
	client := &http.Client{
		Timeout: time.Millisecond * 10,
	}

	// Create the request
	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	resp, err := Do(client, req, 2, time.Millisecond)
	// resp, err := client.Do(req)

	if err == nil {
		t.Fatalf("expected error, got: %#v", err)
	}

	if ne, ok := err.(net.Error); ok && !ne.Timeout() {
		t.Fatalf("expected timeout error, got: %#v", err)
	}

	if resp != nil {
		t.Fatalf("Invalid response: %v", resp)
	}
}
