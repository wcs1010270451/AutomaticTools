package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const requestIDKey contextKey = "requestID"

type httpError struct {
	status  int
	message string
}

func New(deps Dependencies) http.Handler {
	app := &App{
		cfg:    deps.Config,
		db:     deps.DB,
		logger: deps.Logger,
		mux:    http.NewServeMux(),
	}
	app.routes()
	return app.middleware(app.mux)
}

func (a *App) routes() {
	a.mux.HandleFunc("GET /health", a.health)
	a.mux.HandleFunc("POST /api/auth/register", a.register)
	a.mux.HandleFunc("POST /api/auth/login", a.login)
	a.mux.HandleFunc("GET /api/me", a.me)
	a.mux.HandleFunc("GET /api/products", a.products)
	a.mux.HandleFunc("GET /api/tools", a.tools)
	a.mux.HandleFunc("POST /api/orders", a.createOrder)
	a.mux.HandleFunc("GET /api/me/orders", a.myOrders)
	a.mux.HandleFunc("GET /api/me/entitlements", a.myEntitlements)
	a.mux.HandleFunc("POST /api/devices/bind", a.bindDevice)
	a.mux.HandleFunc("POST /api/admin/entitlements/grant", a.adminGrantEntitlement)
	a.mux.HandleFunc("POST /api/admin/orders/confirm", a.adminConfirmOrder)
}

func (a *App) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := newRequestID()
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Admin-Key")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r.WithContext(ctx))

		a.logger.Info(
			"http request",
			"requestId", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"durationMs", time.Since(start).Milliseconds(),
			"ip", clientIP(r),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

func requestID(r *http.Request) string {
	value, _ := r.Context().Value(requestIDKey).(string)
	return value
}

func newRequestID() string {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(bytes[:])
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (a *App) fail(w http.ResponseWriter, r *http.Request, status int, message string, attrs ...any) {
	a.logger.Warn(
		"request failed",
		append([]any{"requestId", requestID(r), "status", status, "error", message}, attrs...)...,
	)
	writeJSON(w, status, ErrorResponse{Error: message, RequestID: requestID(r)})
}

func (a *App) handleErr(w http.ResponseWriter, r *http.Request, err error) {
	var httpErr httpError
	if errors.As(err, &httpErr) {
		a.fail(w, r, httpErr.status, httpErr.message)
		return
	}
	a.logger.Error("unexpected error", "requestId", requestID(r), "error", err)
	writeJSON(w, http.StatusInternalServerError, ErrorResponse{
		Error:     "服务器内部错误。",
		RequestID: requestID(r),
	})
}

func badRequest(message string) httpError {
	return httpError{status: http.StatusBadRequest, message: message}
}

func unauthorized(message string) httpError {
	return httpError{status: http.StatusUnauthorized, message: message}
}

func forbidden(message string) httpError {
	return httpError{status: http.StatusForbidden, message: message}
}

func conflict(message string) httpError {
	return httpError{status: http.StatusConflict, message: message}
}

func clientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (e httpError) Error() string {
	return e.message
}

func logAttrsForUser(userID int64) slog.Attr {
	return slog.Int64("userId", userID)
}
