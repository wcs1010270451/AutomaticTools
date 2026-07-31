package logic

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"automatictools/backend/internal/store"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CreateAdminRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Status   string `json:"status"`
}

type UpdateAdminRequest struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Status   string `json:"status"`
}

func (a *Service) ListAdmins(ctx context.Context) ([]AdminDTO, error) {
	var records []store.Admin
	err := a.db.WithContext(ctx).
		Order("id").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	admins := make([]AdminDTO, 0, len(records))
	for _, record := range records {
		admins = append(admins, adminDTO(record))
	}
	return admins, nil
}

func (a *Service) CreateAdmin(
	ctx context.Context,
	actorAdminID int64,
	req CreateAdminRequest,
	meta RequestMeta,
) (AdminDTO, error) {
	req.Username = normalizeUsername(req.Username)
	req.Status = normalizeAdminStatus(req.Status)
	if err := validateAdminWrite(req.Username, req.Password, req.Status, true); err != nil {
		return AdminDTO{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return AdminDTO{}, err
	}

	now := unixNow()
	record := store.Admin{
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		Status:       req.Status,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := a.db.WithContext(ctx).Create(&record).Error; err != nil {
		if isAdminUsernameConflict(err) {
			return AdminDTO{}, conflict("管理员用户名已存在。")
		}
		return AdminDTO{}, err
	}

	detail, _ := json.Marshal(map[string]any{
		"actorAdminId": actorAdminID,
		"adminId":      record.ID,
		"username":     record.Username,
		"status":       record.Status,
	})
	_ = a.audit(ctx, meta, nil, "admin.create", string(detail))
	return adminDTO(record), nil
}

func (a *Service) UpdateAdmin(
	ctx context.Context,
	actorAdminID int64,
	adminID int64,
	req UpdateAdminRequest,
	meta RequestMeta,
) (AdminDTO, error) {
	req.Username = normalizeUsername(req.Username)
	req.Status = normalizeAdminStatus(req.Status)
	if err := validateAdminWrite(req.Username, req.Password, req.Status, false); err != nil {
		return AdminDTO{}, err
	}
	if actorAdminID == adminID && req.Status != "active" {
		return AdminDTO{}, forbidden("不能停用当前登录的管理员。")
	}

	updates := map[string]any{
		"username":   req.Username,
		"status":     req.Status,
		"updated_at": unixNow(),
	}
	if req.Password != "" {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return AdminDTO{}, err
		}
		updates["password_hash"] = string(passwordHash)
	}

	result := a.db.WithContext(ctx).
		Model(&store.Admin{}).
		Where("id = ?", adminID).
		Updates(updates)
	if result.Error != nil {
		if isAdminUsernameConflict(result.Error) {
			return AdminDTO{}, conflict("管理员用户名已存在。")
		}
		return AdminDTO{}, result.Error
	}
	if result.RowsAffected == 0 {
		return AdminDTO{}, notFound("管理员不存在。")
	}

	var record store.Admin
	if err := a.db.WithContext(ctx).Where("id = ?", adminID).Take(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AdminDTO{}, notFound("管理员不存在。")
		}
		return AdminDTO{}, err
	}

	detail, _ := json.Marshal(map[string]any{
		"actorAdminId": actorAdminID,
		"adminId":      record.ID,
		"username":     record.Username,
		"status":       record.Status,
		"passwordSet":  req.Password != "",
	})
	_ = a.audit(ctx, meta, nil, "admin.update", string(detail))
	return adminDTO(record), nil
}

func (a *Service) DeleteAdmin(
	ctx context.Context,
	actorAdminID int64,
	adminID int64,
	meta RequestMeta,
) error {
	if actorAdminID == adminID {
		return forbidden("不能删除当前登录的管理员。")
	}

	var deleted store.Admin
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", adminID).
			Take(&deleted).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return notFound("管理员不存在。")
			}
			return err
		}

		if deleted.Status == "active" {
			var activeAdmins []store.Admin
			if err := tx.
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("status = ?", "active").
				Find(&activeAdmins).Error; err != nil {
				return err
			}
			if len(activeAdmins) <= 1 {
				return conflict("至少需要保留一个启用状态的管理员。")
			}
		}
		return tx.Delete(&store.Admin{}, adminID).Error
	})
	if err != nil {
		return err
	}

	detail, _ := json.Marshal(map[string]any{
		"actorAdminId": actorAdminID,
		"adminId":      deleted.ID,
		"username":     deleted.Username,
	})
	_ = a.audit(ctx, meta, nil, "admin.delete", string(detail))
	return nil
}

func adminDTO(record store.Admin) AdminDTO {
	return AdminDTO{
		ID:          record.ID,
		Username:    record.Username,
		Status:      record.Status,
		CreatedAt:   record.CreatedAt,
		LastLoginAt: record.LastLoginAt,
	}
}

func normalizeAdminStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return "active"
	}
	return status
}

func validateAdminWrite(username string, password string, status string, passwordRequired bool) error {
	if !usernamePattern.MatchString(username) {
		return badRequest("管理员用户名必须是 3 到 32 位，只能包含字母、数字、下划线、横线或点。")
	}
	if passwordRequired && password == "" {
		return badRequest("管理员密码不能为空。")
	}
	if password != "" && len(password) < 6 {
		return badRequest("管理员密码至少需要 6 个字符。")
	}
	if status != "active" && status != "disabled" {
		return badRequest("管理员状态必须是 active 或 disabled。")
	}
	return nil
}

func isAdminUsernameConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
