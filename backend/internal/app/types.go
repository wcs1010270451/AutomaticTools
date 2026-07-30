package app

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"automatictools/backend/internal/config"
)

type Dependencies struct {
	Config config.Config
	DB     *sql.DB
	Logger *slog.Logger
}

type App struct {
	cfg    config.Config
	db     *sql.DB
	logger *slog.Logger
	mux    *http.ServeMux
}

type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"requestId"`
}

type UserDTO struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Status      string `json:"status,omitempty"`
	CreatedAt   int64  `json:"createdAt,omitempty"`
	LastLoginAt *int64 `json:"lastLoginAt,omitempty"`
}

type ToolDTO struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"priceCents"`
	Currency    string `json:"currency"`
	Lifetime    bool   `json:"lifetime"`
}

type EntitlementDTO struct {
	ToolCode  string `json:"toolCode"`
	Source    string `json:"source"`
	ExpiresAt *int64 `json:"expiresAt"`
	CreatedAt int64  `json:"createdAt"`
}

type OrderDTO struct {
	OrderNo     string `json:"orderNo"`
	ToolCode    string `json:"toolCode"`
	AmountCents int64  `json:"amountCents"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
	PayChannel  string `json:"payChannel"`
	PaidAt      *int64 `json:"paidAt"`
	CreatedAt   int64  `json:"createdAt"`
}

func unixNow() int64 {
	return time.Now().Unix()
}
