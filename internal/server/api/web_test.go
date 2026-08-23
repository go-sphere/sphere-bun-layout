package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestAdminRoutesRequireJWT(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	web := NewWebServer(Config{
		JWT:  "test-jwt-secret",
		HTTP: HTTPConfig{Address: address},
	}, nil)
	startErr := make(chan error, 1)
	go func() {
		startErr <- web.Start(context.Background())
	}()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = web.Stop(ctx)
		select {
		case err := <-startErr:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				t.Errorf("server start: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("server did not stop")
		}
	})

	client := &http.Client{Timeout: time.Second}
	url := "http://" + address + "/api/admin/list?page_size=10"
	deadline := time.Now().Add(2 * time.Second)
	for {
		response, err := client.Get(url)
		if err == nil {
			defer response.Body.Close()
			if response.StatusCode != http.StatusUnauthorized {
				t.Fatalf("unauthenticated admin request status = %d, want %d", response.StatusCode, http.StatusUnauthorized)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not accept requests: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
