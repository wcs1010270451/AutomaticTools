package router

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGinHealthRoute(t *testing.T) {
	handler := New(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	if response.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing request ID header")
	}

	var payload struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.OK {
		t.Fatal("health response should be ok")
	}
}

func TestGinCORSPreflight(t *testing.T) {
	handler := New(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	request := httptest.NewRequest(http.MethodOptions, "/api/tools", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatal("missing CORS allow headers")
	}
}

func TestGinProtectedRouteRequiresToken(t *testing.T) {
	handler := New(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	var payload struct {
		Error     string `json:"error"`
		RequestID string `json:"requestId"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error == "" || payload.RequestID == "" {
		t.Fatalf("unexpected error response: %+v", payload)
	}
}

func TestGinPaymentRoutesRequireToken(t *testing.T) {
	handler := New(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	for _, tt := range []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/api/me/purchases"},
		{method: http.MethodPost, path: "/api/payments/alipay/precreate", body: `{"toolCode":"auto_click"}`},
		{method: http.MethodGet, path: "/api/payments/orders/ord_test/status"},
	} {
		request := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s returned %d", tt.method, tt.path, response.Code)
		}
	}
}

func TestGinLicenseCodeRoutesRequireToken(t *testing.T) {
	tests := []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodPost, path: "/api/license-codes/redeem", body: `{"code":"AT-TEST-TEST-TEST-TEST"}`},
		{method: http.MethodGet, path: "/api/admin/license-codes"},
		{method: http.MethodPost, path: "/api/admin/license-codes/generate", body: `{"toolCode":"auto_click","count":1}`},
		{method: http.MethodPost, path: "/api/admin/license-codes/1/revoke"},
	}
	for _, tt := range tests {
		request := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		New(Dependencies{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}).ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s returned %d", tt.method, tt.path, response.Code)
		}
	}
}

func TestGinAlipayNotificationRejectsWhenDisabled(t *testing.T) {
	handler := New(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/payments/alipay/notify",
		bytes.NewBufferString("out_trade_no=ord_test"),
	)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || response.Body.String() != "failure" {
		t.Fatalf("unexpected response: %d %q", response.Code, response.Body.String())
	}
}

func TestGinAdminProtectedRouteRequiresAdminToken(t *testing.T) {
	handler := New(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	body := bytes.NewBufferString(`{"orderNo":"ord_test"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/admin/orders/confirm", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", response.Code)
	}
}

func TestGinAdminLoginRouteRejectsInvalidJSON(t *testing.T) {
	handler := New(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	request := httptest.NewRequest(http.MethodPost, "/api/admin/auth/login", bytes.NewBufferString(`{`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", response.Code)
	}
}

func TestGinAdminManagementRoutesRequireAdminToken(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "list", method: http.MethodGet, path: "/api/admin/admins"},
		{name: "list users", method: http.MethodGet, path: "/api/admin/users"},
		{name: "create", method: http.MethodPost, path: "/api/admin/admins", body: `{}`},
		{name: "update", method: http.MethodPut, path: "/api/admin/admins/2", body: `{}`},
		{name: "delete", method: http.MethodDelete, path: "/api/admin/admins/2"},
		{name: "list tools", method: http.MethodGet, path: "/api/admin/tools"},
		{name: "create tool", method: http.MethodPost, path: "/api/admin/tools", body: `{}`},
		{name: "update tool", method: http.MethodPut, path: "/api/admin/tools/auto_click", body: `{}`},
		{name: "list license codes", method: http.MethodGet, path: "/api/admin/license-codes"},
		{name: "generate license codes", method: http.MethodPost, path: "/api/admin/license-codes/generate", body: `{}`},
		{name: "revoke license code", method: http.MethodPost, path: "/api/admin/license-codes/1/revoke"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(Dependencies{
				Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			})
			request := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("unexpected status: %d", response.Code)
			}
		})
	}
}

func TestGinRejectsUnknownJSONFields(t *testing.T) {
	handler := New(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	body := bytes.NewBufferString(`{"username":"user","password":"secret","unknown":true}`)
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", response.Code)
	}
}

func TestGinRegisterRejectsPasswordBeyondBcryptLimit(t *testing.T) {
	handler := New(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	body, err := json.Marshal(map[string]string{
		"username": "test_user",
		"password": string(bytes.Repeat([]byte("a"), 73)),
	})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", response.Code)
	}
}

func TestGinLoginRequiresAccountAndPassword(t *testing.T) {
	handler := New(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/login",
		bytes.NewBufferString(`{"account":"","password":""}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", response.Code)
	}
}

func TestGinEmailCodeRejectsInvalidEmail(t *testing.T) {
	handler := New(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/email-code",
		bytes.NewBufferString(`{"email":"not-an-email"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", response.Code)
	}
}

func TestGinRegisterRequiresEmailCode(t *testing.T) {
	handler := New(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		bytes.NewBufferString(`{"email":"user@example.com","password":"123456"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", response.Code)
	}
}
