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
	ID               int64  `gorm:"column:id;primaryKey"`
	OrderNo          string `gorm:"column:order_no"`
	UserID           int64  `gorm:"column:user_id"`
	ToolCode         string `gorm:"column:tool_code"`
	AmountCents      int64  `gorm:"column:amount_cents"`
	Currency         string `gorm:"column:currency"`
	Status           string `gorm:"column:status"`
	PayChannel       string `gorm:"column:pay_channel"`
	ProviderTradeNo  string `gorm:"column:provider_trade_no"`
	PaymentQRCode    string `gorm:"column:payment_qr_code"`
	PaymentExpiresAt *int64 `gorm:"column:payment_expires_at"`
	PaidAt           *int64 `gorm:"column:paid_at"`
	CreatedAt        int64  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt        int64  `gorm:"column:updated_at;autoUpdateTime:false"`
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

type LicenseCode struct {
	ID               int64  `gorm:"column:id;primaryKey"`
	CodeHash         string `gorm:"column:code_hash"`
	CodeHint         string `gorm:"column:code_hint"`
	ToolCode         string `gorm:"column:tool_code"`
	BatchNo          string `gorm:"column:batch_no"`
	Note             string `gorm:"column:note"`
	Status           string `gorm:"column:status"`
	CreatedByAdminID *int64 `gorm:"column:created_by_admin_id"`
	RedeemedByUserID *int64 `gorm:"column:redeemed_by_user_id"`
	RedeemedAt       *int64 `gorm:"column:redeemed_at"`
	RevokedAt        *int64 `gorm:"column:revoked_at"`
	CreatedAt        int64  `gorm:"column:created_at;autoCreateTime:false"`
}

func (LicenseCode) TableName() string {
	return "license_codes"
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

// ===== Game Models =====

type GamePlayer struct {
	ID                  int64  `gorm:"column:id;primaryKey"`
	UserID              int64  `gorm:"column:user_id"`
	Nickname            string `gorm:"column:nickname"`
	Level               int    `gorm:"column:level"`
	Exp                 int64  `gorm:"column:exp"`
	Gold                int64  `gorm:"column:gold"`
	CombatPower         int64  `gorm:"column:combat_power"`
	Wins                int    `gorm:"column:wins"`
	Losses              int    `gorm:"column:losses"`
	BoxPityCounter      int    `gorm:"column:box_pity_counter"`
	DailyChallengeCount int    `gorm:"column:daily_challenge_count"`
	DailyChallengeDate  string `gorm:"column:daily_challenge_date"`
	CreatedAt           int64  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt           int64  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (GamePlayer) TableName() string {
	return "game_players"
}

type GameEquipmentTemplate struct {
	ID          int64  `gorm:"column:id;primaryKey"`
	Name        string `gorm:"column:name"`
	Slot        string `gorm:"column:slot"`
	Rarity      int    `gorm:"column:rarity"`
	BaseAtk     int    `gorm:"column:base_atk"`
	BaseDef     int    `gorm:"column:base_def"`
	BaseHp      int    `gorm:"column:base_hp"`
	BaseSpd     int    `gorm:"column:base_spd"`
	IconPath    string `gorm:"column:icon_path"`
	Description string `gorm:"column:description"`
	CreatedAt   int64  `gorm:"column:created_at;autoCreateTime:false"`
}

func (GameEquipmentTemplate) TableName() string {
	return "game_equipment_templates"
}

type GamePlayerEquipment struct {
	ID         int64  `gorm:"column:id;primaryKey"`
	PlayerID   int64  `gorm:"column:player_id"`
	TemplateID int64  `gorm:"column:template_id"`
	Rarity     int    `gorm:"column:rarity"`
	Atk        int    `gorm:"column:atk"`
	Def        int    `gorm:"column:def"`
	Hp         int    `gorm:"column:hp"`
	Spd        int    `gorm:"column:spd"`
	BonusAttrs string `gorm:"column:bonus_attrs"`
	Equipped   bool   `gorm:"column:equipped"`
	ObtainedAt int64  `gorm:"column:obtained_at;autoCreateTime:false"`
}

func (GamePlayerEquipment) TableName() string {
	return "game_player_equipments"
}

type GameBattle struct {
	ID         int64  `gorm:"column:id;primaryKey"`
	AttackerID int64  `gorm:"column:attacker_id"`
	DefenderID int64  `gorm:"column:defender_id"`
	WinnerID   int64  `gorm:"column:winner_id"`
	BattleLog  string `gorm:"column:battle_log"`
	RewardGold int64  `gorm:"column:reward_gold"`
	RewardExp  int64  `gorm:"column:reward_exp"`
	CreatedAt  int64  `gorm:"column:created_at;autoCreateTime:false"`
}

func (GameBattle) TableName() string {
	return "game_battles"
}

type GameLeaderboard struct {
	ID        int64  `gorm:"column:id;primaryKey"`
	PlayerID  int64  `gorm:"column:player_id"`
	RankType  string `gorm:"column:rank_type"`
	Score     int64  `gorm:"column:score"`
	Rank      int    `gorm:"column:rank"`
	UpdatedAt int64  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (GameLeaderboard) TableName() string {
	return "game_leaderboard"
}

type GameBoxRecord struct {
	ID                int64  `gorm:"column:id;primaryKey"`
	PlayerID          int64  `gorm:"column:player_id"`
	BoxType           string `gorm:"column:box_type"`
	ResultEquipmentID *int64 `gorm:"column:result_equipment_id"`
	IsPity            bool   `gorm:"column:is_pity"`
	CreatedAt         int64  `gorm:"column:created_at;autoCreateTime:false"`
}

func (GameBoxRecord) TableName() string {
	return "game_box_records"
}
