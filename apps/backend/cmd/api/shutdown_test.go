package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestShutdownPreservesActiveResponseAndRejectsNewConnections(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	server, address, served := startShutdownTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-release:
			_, _ = io.WriteString(w, "completed")
		case <-r.Context().Done():
		}
	}))
	response := make(chan string, 1)
	go func() {
		client := &http.Client{Timeout: 3 * time.Second}
		request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+address, nil)
		if err != nil {
			response <- err.Error()
			return
		}
		resp, err := client.Do(request)
		if err != nil {
			response <- err.Error()
			return
		}
		body, readErr := io.ReadAll(resp.Body)
		if err := errors.Join(readErr, resp.Body.Close()); err != nil {
			response <- err.Error()
			return
		}
		response <- string(body)
	}()
	waitShutdownEvent(t, started)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stopped := make(chan error, 1)
	go func() { stopped <- shutdownHTTP(ctx, server) }()
	waitShutdownEvent(t, served) // Shutdown has closed the listener.
	dialer := net.Dialer{Timeout: time.Second}
	connection, err := dialer.DialContext(t.Context(), "tcp", address)
	if err == nil {
		if err := connection.Close(); err != nil {
			t.Errorf("close unexpected connection: %v", err)
		}
		t.Fatal("new connection accepted while draining")
	}
	select {
	case err := <-stopped:
		t.Fatalf("shutdown returned before active response: %v", err)
	default:
	}
	close(release)
	select {
	case got := <-response:
		if got != "completed" {
			t.Fatalf("active response = %q", got)
		}
	case <-ctx.Done():
		t.Fatal("active response did not complete")
	}
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("shutdown did not complete")
	}
}

func TestShutdownDeadlineCancelsRemainingRequest(t *testing.T) {
	started := make(chan struct{})
	canceled := make(chan struct{})
	server, address, _ := startShutdownTestServer(t, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
		close(canceled)
	}))
	clientFinished := make(chan struct{})
	go func() {
		defer close(clientFinished)
		client := &http.Client{Timeout: 3 * time.Second}
		request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+address, nil)
		if err != nil {
			t.Errorf("create request: %v", err)
			return
		}
		resp, err := client.Do(request)
		if err == nil {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close response: %v", err)
			}
		}
	}()
	waitShutdownEvent(t, started)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := shutdownHTTP(ctx, server); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("shutdown error = %v", err)
	}
	waitShutdownEvent(t, canceled)
	waitShutdownEvent(t, clientFinished)
}

func startShutdownTestServer(t *testing.T, handler http.Handler) (*http.Server, string, <-chan struct{}) {
	t.Helper()
	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: handler, ReadHeaderTimeout: time.Second}
	t.Cleanup(func() { _ = server.Close() })
	served := make(chan struct{})
	go func() { defer close(served); _ = server.Serve(listener) }()
	return server, listener.Addr().String(), served
}

func waitShutdownEvent(t *testing.T, event <-chan struct{}) {
	t.Helper()
	select {
	case <-event:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for shutdown event")
	}
}
