package retryrequest

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

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

// Takes a second to finish processing
var handlerSecond = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	// Block
	select {
	case <-r.Context().Done():
	case <-time.After(time.Second):
	}

	w.Write([]byte("Done"))
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

func TestContextTimeout(t *testing.T) {

	// Mock server which always responds 500.
	ts := httptest.NewServer(handlerAlwaysTimeout)
	defer ts.Close()

	// Create the client. Use short retry windows so we fail faster.
	client := &http.Client{
		Timeout: time.Millisecond * 100,
	}

	// Create the request
	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)

	go func() {
		<-time.After(time.Millisecond * 10)
		cancel()
	}()

	resp, err := Do(client, req, 2, time.Millisecond*100)

	// Checking our cancel of the context is challenging
	// https://github.com/golang/go/blob/cc8838d645b2b7026c1f3aaceb011775c5ca3a08/src/net/http/client.go#L645-L649

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Expected: %q, got %q", ctx.Err(), err)
	}

	// if ne, ok := err.(net.Error); ok && !ne.Timeout() {
	// 	t.Fatalf("expected timeout error, got: %v", err)
	// }

	if !strings.Contains(err.Error(), ctx.Err().Error()) {
		t.Fatalf("expected context timeout error, got: %v", err)
	}

	if resp != nil {
		t.Fatalf("Invalid response: %v", resp)
	}
}
