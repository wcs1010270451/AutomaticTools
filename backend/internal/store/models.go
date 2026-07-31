package store

type User struct {
	ID           int64   `gorm:"column:id;primaryKey"`
	Username     string  `gorm:"column:username"`
	Email        *string `gorm:"column:email"`
	Phone        *string `gorm:"column:phone"`
	PasswordHash string  `gorm:"column:password_hash"`
	Status       string  `gorm:"column:status"`
	CreatedAt    int64   `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt    int64   `gorm:"column:updated_at;autoUpdateTime:false"`
	LastLoginAt  *int64  `gorm:"column:last_login_at"`
}

func (User) TableName() string {
	return "users"
}

type Admin struct {
	ID           int64  `gorm:"column:id;primaryKey"`
	Username     string `gorm:"column:username"`
	PasswordHash string `gorm:"column:password_hash"`
	Status       string `gorm:"column:status"`
	CreatedAt    int64  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt    int64  `gorm:"column:updated_at;autoUpdateTime:false"`
	LastLoginAt  *int64 `gorm:"column:last_login_at"`
}

func (Admin) TableName() string {
	return "admins"
}

type EmailVerificationCode struct {
	ID           int64  `gorm:"column:id;primaryKey"`
	Email        string `gorm:"column:email"`
	Purpose      string `gorm:"column:purpose"`
	CodeHash     string `gorm:"column:code_hash"`
	ExpiresAt    int64  `gorm:"column:expires_at"`
	ConsumedAt   *int64 `gorm:"column:consumed_at"`
	AttemptCount int    `gorm:"column:attempt_count"`
	CreatedAt    int64  `gorm:"column:created_at;autoCreateTime:false"`
}

func (EmailVerificationCode) TableName() string {
	return "email_verification_codes"
}

type Tool struct {
	ID          int64  `gorm:"column:id;primaryKey"`
	Code        string `gorm:"column:code"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	PriceCents  int64  `gorm:"column:price_cents"`
	Currency    string `gorm:"column:currency"`
	Lifetime    bool   `gorm:"column:lifetime"`
	Active      bool   `gorm:"column:active"`
	CreatedAt   int64  `gorm:"column:created_at;autoCreateTime:false"`
}

func (Tool) TableName() string {
	return "tools"
}

type Order struct {
	ID          int64  `gorm:"column:id;primaryKey"`
	OrderNo     string `gorm:"column:order_no"`
	UserID      int64  `gorm:"column:user_id"`
	ToolCode    string `gorm:"column:tool_code"`
	AmountCents int64  `gorm:"column:amount_cents"`
	Currency    string `gorm:"column:currency"`
	Status      string `gorm:"column:status"`
	PayChannel  string `gorm:"column:pay_channel"`
	PaidAt      *int64 `gorm:"column:paid_at"`
	CreatedAt   int64  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt   int64  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (Order) TableName() string {
	return "orders"
}

type Entitlement struct {
	ID        int64  `gorm:"column:id;primaryKey"`
	UserID    int64  `gorm:"column:user_id"`
	ToolCode  string `gorm:"column:tool_code"`
	Source    string `gorm:"column:source"`
	OrderNo   string `gorm:"column:order_no"`
	ExpiresAt *int64 `gorm:"column:expires_at"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:false"`
}

func (Entitlement) TableName() string {
	return "entitlements"
}

type Device struct {
	ID         int64  `gorm:"column:id;primaryKey"`
	UserID     int64  `gorm:"column:user_id"`
	DeviceID   string `gorm:"column:device_id"`
	DeviceName string `gorm:"column:device_name"`
	Platform   string `gorm:"column:platform"`
	LastSeenAt int64  `gorm:"column:last_seen_at"`
	CreatedAt  int64  `gorm:"column:created_at;autoCreateTime:false"`
}

func (Device) TableName() string {
	return "devices"
}

type AuditLog struct {
	ID        int64  `gorm:"column:id;primaryKey"`
	UserID    *int64 `gorm:"column:user_id"`
	Action    string `gorm:"column:action"`
	IP        string `gorm:"column:ip"`
	UserAgent string `gorm:"column:user_agent"`
	Detail    string `gorm:"column:detail"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:false"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
