package httpsrv_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sphere/httpx"
	apiv1 "github.com/go-sphere/sphere-bun-layout/api/api/v1"
	"github.com/go-sphere/sphere-bun-layout/internal/pkg/httpsrv"
	"github.com/go-sphere/sphere/log"
	"github.com/go-sphere/sphere/log/logbuffer"
	"github.com/go-sphere/sphere/server/httpz"
	"github.com/go-sphere/sphere/server/middleware/requestid"
	"github.com/go-sphere/sphere/storage/storageerr"
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

// A storage sentinel carries no status of its own, so the process-wide parser
// must classify it; falling back to httpx.ParseError answered 500.
func TestStorageErrorsRenderTheirStatus(t *testing.T) {
	engine := httpsrv.NewServer("test", "127.0.0.1:0", httpsrv.Options{})
	fail := func(err error) httpx.Handler {
		return httpz.WithJson(func(httpx.Context) (string, error) { return "", err })
	}
	group := engine.Group("/storage")
	group.GET("/missing", fail(storageerr.ErrNotFound))
	group.GET("/exists", fail(fmt.Errorf("move upload: %w", storageerr.ErrDestExists)))
	requester, ok := httpx.AsTestRequester(engine)
	if !ok {
		t.Fatal("engine does not support in-process requests")
	}

	for _, tt := range []struct {
		path        string
		wantStatus  int
		wantMessage string
	}{
		{path: "/storage/missing", wantStatus: http.StatusNotFound, wantMessage: http.StatusText(http.StatusNotFound)},
		{path: "/storage/exists", wantStatus: http.StatusBadRequest, wantMessage: http.StatusText(http.StatusBadRequest)},
	} {
		t.Run(tt.path, func(t *testing.T) {
			response, err := requester.Do(httptest.NewRequest(http.MethodGet, tt.path, nil))
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			defer func() { _ = response.Body.Close() }()
			if response.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.StatusCode, tt.wantStatus)
			}
			var body httpz.ErrorResponse
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", body.Message, tt.wantMessage)
			}
		})
	}
}

func TestServerAssignsRequestID(t *testing.T) {
	buf := logbuffer.New(32)
	log.InitWithBackends(buf)
	t.Cleanup(func() {
		log.InitWithBackends(log.NewStdioBackend())
	})
	// No route is registered: the engine-level stack must cover a 404 too.
	send := func(opts httpsrv.Options, id string) string {
		t.Helper()
		engine := httpsrv.NewServer("test", "127.0.0.1:0", opts)
		tr, ok := httpx.AsTestRequester(engine)
		if !ok {
			t.Fatalf("engine %T does not support in-process test requests", engine)
		}
		req := httptest.NewRequest(http.MethodGet, "http://example.com/missing", nil)
		req.Header.Set(requestid.DefaultHeader, id)
		resp, err := tr.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", resp.StatusCode)
		}
		return resp.Header.Get(requestid.DefaultHeader)
	}

	generated := send(httpsrv.Options{}, "client-id.1")
	if generated == "" || generated == "client-id.1" {
		t.Fatalf("default: response header = %q, want a generated ID", generated)
	}
	trusted := httpsrv.Options{TrustRequestID: true}
	if got := send(trusted, "client-id.1"); got != "client-id.1" {
		t.Fatalf("TrustRequestID: response header = %q, want the caller's ID kept", got)
	}
	if got := send(trusted, "not a valid id"); got == "" || got == "not a valid id" {
		t.Fatalf("TrustRequestID with a malformed ID: response header = %q, want a generated ID", got)
	}

	entries, _ := buf.History(0, 8, log.LevelDebug)
	for _, entry := range entries {
		if entry.Message == "/missing" && entry.Attrs["request_id"] == generated {
			return
		}
	}
	t.Fatalf("access log entry with request_id %q missing from buffer: %+v", generated, entries)
}
