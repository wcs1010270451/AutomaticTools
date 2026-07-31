package logic

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"automatictools/backend/internal/store"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	toolCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
	currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)
)

type CreateToolRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"priceCents"`
	Currency    string `json:"currency"`
	Lifetime    *bool  `json:"lifetime"`
	Active      *bool  `json:"active"`
}

type UpdateToolRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"priceCents"`
	Currency    string `json:"currency"`
	Lifetime    bool   `json:"lifetime"`
	Active      bool   `json:"active"`
}

func (a *Service) ListAdminTools(ctx context.Context) ([]AdminToolDTO, error) {
	var records []store.Tool
	if err := a.db.WithContext(ctx).Order("id").Find(&records).Error; err != nil {
		return nil, err
	}

	tools := make([]AdminToolDTO, 0, len(records))
	for _, record := range records {
		tools = append(tools, adminToolDTO(record))
	}
	return tools, nil
}

func (a *Service) CreateTool(
	ctx context.Context,
	actorAdminID int64,
	req CreateToolRequest,
	meta RequestMeta,
) (AdminToolDTO, error) {
	req.Code = strings.ToLower(strings.TrimSpace(req.Code))
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.Currency = normalizeCurrency(req.Currency)
	if err := validateToolWrite(req.Code, req.Name, req.Description, req.PriceCents, req.Currency); err != nil {
		return AdminToolDTO{}, err
	}

	lifetime := true
	if req.Lifetime != nil {
		lifetime = *req.Lifetime
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}

	record := store.Tool{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		PriceCents:  req.PriceCents,
		Currency:    req.Currency,
		Lifetime:    lifetime,
		Active:      active,
		CreatedAt:   unixNow(),
	}
	if err := a.db.WithContext(ctx).Create(&record).Error; err != nil {
		if isToolCodeConflict(err) {
			return AdminToolDTO{}, conflict("工具编码已存在。")
		}
		return AdminToolDTO{}, err
	}

	a.auditToolChange(ctx, meta, actorAdminID, "tool.create", record)
	return adminToolDTO(record), nil
}

func (a *Service) UpdateTool(
	ctx context.Context,
	actorAdminID int64,
	code string,
	req UpdateToolRequest,
	meta RequestMeta,
) (AdminToolDTO, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.Currency = normalizeCurrency(req.Currency)
	if err := validateToolWrite(code, req.Name, req.Description, req.PriceCents, req.Currency); err != nil {
		return AdminToolDTO{}, err
	}

	result := a.db.WithContext(ctx).
		Model(&store.Tool{}).
		Where("code = ?", code).
		Updates(map[string]any{
			"name":        req.Name,
			"description": req.Description,
			"price_cents": req.PriceCents,
			"currency":    req.Currency,
			"lifetime":    req.Lifetime,
			"active":      req.Active,
		})
	if result.Error != nil {
		return AdminToolDTO{}, result.Error
	}
	if result.RowsAffected == 0 {
		return AdminToolDTO{}, notFound("工具不存在。")
	}

	var record store.Tool
	if err := a.db.WithContext(ctx).Where("code = ?", code).Take(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AdminToolDTO{}, notFound("工具不存在。")
		}
		return AdminToolDTO{}, err
	}

	a.auditToolChange(ctx, meta, actorAdminID, "tool.update", record)
	return adminToolDTO(record), nil
}

func adminToolDTO(record store.Tool) AdminToolDTO {
	return AdminToolDTO{
		ID:          record.ID,
		Code:        record.Code,
		Name:        record.Name,
		Description: record.Description,
		PriceCents:  record.PriceCents,
		Currency:    record.Currency,
		Lifetime:    record.Lifetime,
		Active:      record.Active,
		CreatedAt:   record.CreatedAt,
	}
}

func normalizeCurrency(currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return "CNY"
	}
	return currency
}

func validateToolWrite(code string, name string, description string, priceCents int64, currency string) error {
	if !toolCodePattern.MatchString(code) {
		return badRequest("工具编码必须以小写字母开头，只能包含小写字母、数字和下划线，最长 64 位。")
	}
	if name == "" || utf8.RuneCountInString(name) > 100 {
		return badRequest("工具名称不能为空，且不能超过 100 个字符。")
	}
	if utf8.RuneCountInString(description) > 2000 {
		return badRequest("工具描述不能超过 2000 个字符。")
	}
	if priceCents < 0 {
		return badRequest("工具价格不能小于 0。")
	}
	if !currencyPattern.MatchString(currency) {
		return badRequest("货币代码必须是 3 位大写字母。")
	}
	return nil
}

func isToolCodeConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (a *Service) auditToolChange(
	ctx context.Context,
	meta RequestMeta,
	actorAdminID int64,
	action string,
	record store.Tool,
) {
	detail, _ := json.Marshal(map[string]any{
		"actorAdminId": actorAdminID,
		"toolCode":     record.Code,
		"name":         record.Name,
		"priceCents":   record.PriceCents,
		"currency":     record.Currency,
		"lifetime":     record.Lifetime,
		"active":       record.Active,
	})
	_ = a.audit(ctx, meta, nil, action, string(detail))
}
