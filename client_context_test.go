package turnkey

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type localTransport func(*http.Request) (*http.Response, error)

func (f localTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestLocalRequestCancellation(t *testing.T) {
	started := make(chan context.Context, 1)
	release := make(chan struct{})
	done := make(chan error, 1)
	var attempts atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer close(release)
	transport := localTransport(func(r *http.Request) (*http.Response, error) {
		attempts.Add(1)
		started <- r.Context()
		select {
		case <-r.Context().Done():
			return nil, r.Context().Err()
		case <-release:
			return nil, errors.New("transport released")
		}
	})
	c := &Client{config: &config{httpClient: &http.Client{Transport: transport}, http: httpConfig{maxRetries: 2, retryDelay: time.Millisecond}}}
	go func() {
		_, err := c.postWithRetry(ctx, "https://example.invalid", "/fixture", nil, nil, nil)
		done <- err
	}()
	var requestContext context.Context
	select {
	case requestContext = <-started:
	case <-time.After(time.Second):
		t.Fatal("transport did not start")
	}
	cancel()
	select {
	case <-requestContext.Done():
	case <-time.After(time.Second):
		t.Fatal("request context did not cancel")
	}
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("request did not return")
	}
	if attempts.Load() != 1 {
		t.Fatalf("attempts = %d", attempts.Load())
	}
}

func TestLocalRequestSuccess(t *testing.T) {
	c := &Client{config: &config{httpClient: &http.Client{Transport: localTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"ok":true}`)), Request: r}, nil
	})}}}
	data, err := c.postWithRetry(context.Background(), "https://example.invalid", "/fixture", nil, nil, nil)
	if err != nil || string(data) != `{"ok":true}` {
		t.Fatalf("%s %v", data, err)
	}
}
