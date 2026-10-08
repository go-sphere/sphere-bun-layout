package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-sphere/sphere-bun-layout/internal/pkg/httpsrv"
	"github.com/go-sphere/sphere/server/auth/jwtauth"
)

const testJWTSecret = "test-jwt-secret"

// startTestWeb assembles and starts the server as the app does and returns its
// base URL once it answers requests.
func startTestWeb(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	web := NewWebServer(Config{
		JWT:  testJWTSecret,
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

	deadline := time.Now().Add(2 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return "http://" + address
		}
		select {
		case err := <-startErr:
			t.Fatalf("server start: %v", err)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not accept connections: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestAdminRoutesRequireJWT(t *testing.T) {
	baseURL := startTestWeb(t)
	client := &http.Client{Timeout: time.Second}
	response, err := client.Get(baseURL + "/api/admin/list?page_size=10")
	if err != nil {
		t.Fatalf("GET /api/admin/list: %v", err)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatalf("close response body: %v", err)
	}
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated admin request status = %d, want %d", response.StatusCode, http.StatusUnauthorized)
	}
}

func TestOversizedBodyIsRejectedWith413(t *testing.T) {
	baseURL := startTestWeb(t)
	token, err := jwtauth.NewJwtAuth[jwtauth.RBACClaims[int64]](testJWTSecret).GenerateToken(
		t.Context(),
		jwtauth.NewRBACClaims[int64](1, "admin", nil, time.Now().Add(time.Hour)),
	)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	body := `{"admin":{"username":"` + strings.Repeat("a", int(httpsrv.DefaultMaxBodyBytes)) + `"}}`
	request, err := http.NewRequest(http.MethodPost, baseURL+"/api/admin/create", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := (&http.Client{Timeout: 3 * time.Second}).Do(request)
	if err != nil {
		t.Fatalf("POST /api/admin/create: %v", err)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatalf("close response body: %v", err)
	}
	if response.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusRequestEntityTooLarge)
	}
}
