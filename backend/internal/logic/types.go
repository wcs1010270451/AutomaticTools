package logic

import (
	"time"

	"automatictools/backend/internal/config"

	"gorm.io/gorm"
)

type Dependencies struct {
	Config      config.Config
	DB          *gorm.DB
	EmailSender EmailSender
}

type Service struct {
	cfg         config.Config
	db          *gorm.DB
	emailSender EmailSender
}

func New(deps Dependencies) *Service {
	return &Service{
		cfg:         deps.Config,
		db:          deps.DB,
		emailSender: deps.EmailSender,
	}
}

type EmailSender interface {
	SendRegistrationCode(to string, code string, validMinutes int) error
}

type RequestMeta struct {
	IP        string
	UserAgent string
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

type AdminDTO struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Status      string `json:"status"`
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

type AdminToolDTO struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"priceCents"`
	Currency    string `json:"currency"`
	Lifetime    bool   `json:"lifetime"`
	Active      bool   `json:"active"`
	CreatedAt   int64  `json:"createdAt"`
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
