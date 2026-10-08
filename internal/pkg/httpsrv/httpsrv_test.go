package httpsrv_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sphere/httpx"
	apiv1 "github.com/go-sphere/sphere-bun-layout/api/api/v1"
	"github.com/go-sphere/sphere-bun-layout/internal/pkg/httpsrv"
	"github.com/go-sphere/sphere/server/httpz"
)

type adminService struct {
	apiv1.AdminServiceHTTPServer
}

func (adminService) ListAdmins(context.Context, *apiv1.ListAdminsRequest) (*apiv1.ListAdminsResponse, error) {
	return &apiv1.ListAdminsResponse{}, nil
}

func TestValidationErrorRendersBadRequest(t *testing.T) {
	engine := httpsrv.NewServer("test", "127.0.0.1:0", httpsrv.Options{})
	apiv1.RegisterAdminServiceHTTPServer(engine.Group("/"), adminService{})
	requester, ok := httpx.AsTestRequester(engine)
	if !ok {
		t.Fatal("engine does not support in-process requests")
	}

	response, err := requester.Do(httptest.NewRequest(http.MethodGet, "/api/admin/list?page=-1", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
	var body httpz.ErrorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Success {
		t.Error("success = true, want false")
	}
	if !strings.Contains(body.Message, "greater than or equal to 0") {
		t.Errorf("message = %q, want the page violation", body.Message)
	}
}

func TestBodyCapErrorRendersRequestEntityTooLarge(t *testing.T) {
	engine := httpsrv.NewServer("test", "127.0.0.1:0", httpsrv.Options{})
	// A binder wraps the read error as a 400; the cap must still win.
	engine.Group("/").POST("/upload", httpz.WithJson(func(httpx.Context) (string, error) {
		return "", httpx.WrapBindError(&http.MaxBytesError{Limit: 64})
	}))
	requester, ok := httpx.AsTestRequester(engine)
	if !ok {
		t.Fatal("engine does not support in-process requests")
	}
	response, err := requester.Do(httptest.NewRequest(http.MethodPost, "/upload", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusRequestEntityTooLarge)
	}
}
