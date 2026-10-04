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
		resp, err := client.Get("http://" + address)
		if err != nil {
			response <- err.Error()
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
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
	connection, err := net.DialTimeout("tcp", address, time.Second)
	if err == nil {
		connection.Close()
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
	server, address, _ := startShutdownTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
		close(canceled)
	}))
	clientFinished := make(chan struct{})
	go func() {
		defer close(clientFinished)
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("http://" + address)
		if err == nil {
			resp.Body.Close()
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
	listener, err := net.Listen("tcp", "127.0.0.1:0")
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
