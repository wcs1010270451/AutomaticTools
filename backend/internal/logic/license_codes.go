package logic

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"automatictools/backend/internal/store"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	licenseCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	maxLicenseCodeBatch = 100
)

type GenerateLicenseCodesRequest struct {
	ToolCode string `json:"toolCode"`
	Count    int    `json:"count"`
	Note     string `json:"note"`
}

type GeneratedLicenseCodeDTO struct {
	Code     string `json:"code"`
	ToolCode string `json:"toolCode"`
	BatchNo  string `json:"batchNo"`
}

type LicenseCodeDTO struct {
	ID                     int64  `json:"id"`
	CodeHint               string `json:"codeHint"`
	ToolCode               string `json:"toolCode"`
	ToolName               string `json:"toolName"`
	BatchNo                string `json:"batchNo"`
	Note                   string `json:"note"`
	Status                 string `json:"status"`
	CreatedByAdminUsername string `json:"createdByAdminUsername,omitempty"`
	RedeemedByUserID       *int64 `json:"redeemedByUserId,omitempty"`
	RedeemedByUsername     string `json:"redeemedByUsername,omitempty"`
	RedeemedByEmail        string `json:"redeemedByEmail,omitempty"`
	RedeemedAt             *int64 `json:"redeemedAt,omitempty"`
	RevokedAt              *int64 `json:"revokedAt,omitempty"`
	CreatedAt              int64  `json:"createdAt"`
}

type LicenseCodeListFilter struct {
	Page     int
	PageSize int
	Status   string
	ToolCode string
	Search   string
}

type LicenseCodeListResult struct {
	Codes    []LicenseCodeDTO `json:"codes"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

type RedeemLicenseCodeRequest struct {
	Code string `json:"code"`
}

type RedeemLicenseCodeResult struct {
	Purchase PurchaseDTO `json:"purchase"`
	Tool     ToolDTO     `json:"tool"`
}

type licenseCodeListRow struct {
	store.LicenseCode
	ToolName               string `gorm:"column:tool_name"`
	CreatedByAdminUsername string `gorm:"column:created_by_admin_username"`
	RedeemedByUsername     string `gorm:"column:redeemed_by_username"`
	RedeemedByEmail        string `gorm:"column:redeemed_by_email"`
}

func (a *Service) GenerateLicenseCodes(
	ctx context.Context,
	actorAdminID int64,
	req GenerateLicenseCodesRequest,
	meta RequestMeta,
) (string, []GeneratedLicenseCodeDTO, error) {
	req.ToolCode = strings.ToLower(strings.TrimSpace(req.ToolCode))
	req.Note = strings.TrimSpace(req.Note)
	if req.Count < 1 || req.Count > maxLicenseCodeBatch {
		return "", nil, badRequest("每次只能生成 1 到 100 个授权码。")
	}
	if utf8.RuneCountInString(req.Note) > 200 {
		return "", nil, badRequest("备注不能超过 200 个字符。")
	}

	var tool store.Tool
	if err := a.db.WithContext(ctx).
		Where("code = ? AND active = ?", req.ToolCode, true).
		Take(&tool).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, badRequest("工具不存在或已下架。")
		}
		return "", nil, err
	}

	batchNo, err := newLicenseBatchNo()
	if err != nil {
		return "", nil, err
	}
	now := unixNow()
	records := make([]store.LicenseCode, 0, req.Count)
	generated := make([]GeneratedLicenseCodeDTO, 0, req.Count)
	seen := make(map[string]struct{}, req.Count)
	for len(records) < req.Count {
		code, generateErr := newLicenseCode()
		if generateErr != nil {
			return "", nil, generateErr
		}
		codeHash := hashLicenseCode(code)
		if _, exists := seen[codeHash]; exists {
			continue
		}
		seen[codeHash] = struct{}{}
		records = append(records, store.LicenseCode{
			CodeHash:         codeHash,
			CodeHint:         licenseCodeHint(code),
			ToolCode:         tool.Code,
			BatchNo:          batchNo,
			Note:             req.Note,
			Status:           "active",
			CreatedByAdminID: &actorAdminID,
			CreatedAt:        now,
		})
		generated = append(generated, GeneratedLicenseCodeDTO{
			Code:     code,
			ToolCode: tool.Code,
			BatchNo:  batchNo,
		})
	}

	if err := a.db.WithContext(ctx).CreateInBatches(records, 100).Error; err != nil {
		return "", nil, err
	}
	detail, _ := json.Marshal(map[string]any{
		"actorAdminId": actorAdminID,
		"batchNo":      batchNo,
		"toolCode":     tool.Code,
		"count":        len(records),
		"note":         req.Note,
	})
	_ = a.audit(ctx, meta, nil, "license_code.generate", string(detail))
	return batchNo, generated, nil
}

func (a *Service) ListLicenseCodes(
	ctx context.Context,
	filter LicenseCodeListFilter,
) (LicenseCodeListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	filter.ToolCode = strings.ToLower(strings.TrimSpace(filter.ToolCode))
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.Status != "" && filter.Status != "active" && filter.Status != "redeemed" && filter.Status != "revoked" {
		return LicenseCodeListResult{}, badRequest("授权码状态无效。")
	}

	query := a.db.WithContext(ctx).
		Table("license_codes").
		Joins("JOIN tools ON tools.code = license_codes.tool_code").
		Joins("LEFT JOIN admins creator ON creator.id = license_codes.created_by_admin_id").
		Joins("LEFT JOIN users redeemer ON redeemer.id = license_codes.redeemed_by_user_id")
	if filter.Status != "" {
		query = query.Where("license_codes.status = ?", filter.Status)
	}
	if filter.ToolCode != "" {
		query = query.Where("license_codes.tool_code = ?", filter.ToolCode)
	}
	if filter.Search != "" {
		if normalized, err := normalizeLicenseCode(filter.Search); err == nil {
			query = query.Where("license_codes.code_hash = ?", hashLicenseCode(normalized))
		} else {
			like := "%" + filter.Search + "%"
			query = query.Where(
				"license_codes.batch_no ILIKE ? OR license_codes.code_hint ILIKE ? OR license_codes.note ILIKE ? OR redeemer.username ILIKE ? OR redeemer.email ILIKE ?",
				like, like, like, like, like,
			)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return LicenseCodeListResult{}, err
	}

	var rows []licenseCodeListRow
	err := query.
		Select(`license_codes.*, tools.name AS tool_name,
			COALESCE(creator.username, '') AS created_by_admin_username,
			COALESCE(redeemer.username, '') AS redeemed_by_username,
			COALESCE(redeemer.email, '') AS redeemed_by_email`).
		Order("license_codes.id DESC").
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Scan(&rows).Error
	if err != nil {
		return LicenseCodeListResult{}, err
	}

	codes := make([]LicenseCodeDTO, 0, len(rows))
	for _, row := range rows {
		codes = append(codes, licenseCodeDTO(row))
	}
	return LicenseCodeListResult{
		Codes:    codes,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (a *Service) RevokeLicenseCode(
	ctx context.Context,
	actorAdminID int64,
	licenseCodeID int64,
	meta RequestMeta,
) (LicenseCodeDTO, error) {
	var record store.LicenseCode
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", licenseCodeID).
			Take(&record).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return notFound("授权码不存在。")
			}
			return err
		}
		if record.Status == "redeemed" {
			return conflict("已兑换的授权码不能作废。")
		}
		if record.Status == "revoked" {
			return nil
		}
		now := unixNow()
		if err := tx.Model(&store.LicenseCode{}).
			Where("id = ? AND status = 'active'", record.ID).
			Updates(map[string]any{"status": "revoked", "revoked_at": now}).Error; err != nil {
			return err
		}
		record.Status = "revoked"
		record.RevokedAt = &now
		return nil
	})
	if err != nil {
		return LicenseCodeDTO{}, err
	}

	detail, _ := json.Marshal(map[string]any{
		"actorAdminId":  actorAdminID,
		"licenseCodeId": record.ID,
		"codeHint":      record.CodeHint,
		"toolCode":      record.ToolCode,
		"batchNo":       record.BatchNo,
	})
	_ = a.audit(ctx, meta, nil, "license_code.revoke", string(detail))
	return a.getLicenseCodeDTO(ctx, record.ID)
}

func (a *Service) RedeemLicenseCode(
	ctx context.Context,
	userID int64,
	req RedeemLicenseCodeRequest,
	meta RequestMeta,
) (RedeemLicenseCodeResult, error) {
	normalized, err := normalizeLicenseCode(req.Code)
	if err != nil {
		return RedeemLicenseCodeResult{}, badRequest("授权码格式不正确。")
	}

	var codeRecord store.LicenseCode
	var tool store.Tool
	var entitlement store.Entitlement
	alreadyRedeemed := false
	err = a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("code_hash = ?", hashLicenseCode(normalized)).
			Take(&codeRecord).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return badRequest("授权码不存在或格式不正确。")
			}
			return err
		}

		if codeRecord.Status == "redeemed" {
			if codeRecord.RedeemedByUserID != nil && *codeRecord.RedeemedByUserID == userID {
				alreadyRedeemed = true
				return tx.Where("user_id = ? AND tool_code = ?", userID, codeRecord.ToolCode).
					Take(&entitlement).Error
			}
			return conflict("该授权码已被其他用户兑换。")
		}
		if codeRecord.Status == "revoked" {
			return conflict("该授权码已作废。")
		}

		if err := tx.Where("code = ? AND active = ?", codeRecord.ToolCode, true).
			Take(&tool).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return conflict("授权码对应的工具已下架，请联系卖家。")
			}
			return err
		}

		var existing store.Entitlement
		existingErr := tx.Where("user_id = ? AND tool_code = ?", userID, codeRecord.ToolCode).
			Take(&existing).Error
		if existingErr == nil {
			return conflict("你已经拥有该工具，无需重复兑换。")
		}
		if !errors.Is(existingErr, gorm.ErrRecordNotFound) {
			return existingErr
		}

		now := unixNow()
		entitlement = store.Entitlement{
			UserID:    userID,
			ToolCode:  codeRecord.ToolCode,
			Source:    "license_code",
			OrderNo:   fmt.Sprintf("license_%d", codeRecord.ID),
			ExpiresAt: nil,
			CreatedAt: now,
		}
		if err := tx.Create(&entitlement).Error; err != nil {
			return err
		}
		result := tx.Model(&store.LicenseCode{}).
			Where("id = ? AND status = 'active'", codeRecord.ID).
			Updates(map[string]any{
				"status":              "redeemed",
				"redeemed_by_user_id": userID,
				"redeemed_at":         now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return conflict("授权码状态已变化，请重试。")
		}
		codeRecord.Status = "redeemed"
		codeRecord.RedeemedByUserID = &userID
		codeRecord.RedeemedAt = &now
		return nil
	})
	if err != nil {
		return RedeemLicenseCodeResult{}, err
	}

	if tool.Code == "" {
		if err := a.db.WithContext(ctx).Where("code = ?", codeRecord.ToolCode).Take(&tool).Error; err != nil {
			return RedeemLicenseCodeResult{}, err
		}
	}
	if !alreadyRedeemed {
		detail, _ := json.Marshal(map[string]any{
			"licenseCodeId": codeRecord.ID,
			"codeHint":      codeRecord.CodeHint,
			"toolCode":      codeRecord.ToolCode,
			"batchNo":       codeRecord.BatchNo,
		})
		_ = a.audit(ctx, meta, &userID, "license_code.redeem", string(detail))
	}

	return RedeemLicenseCodeResult{
		Purchase: PurchaseDTO{
			ToolCode:    entitlement.ToolCode,
			OrderNo:     entitlement.OrderNo,
			PurchasedAt: entitlement.CreatedAt,
		},
		Tool: ToolDTO{
			Code:        tool.Code,
			Name:        tool.Name,
			Description: tool.Description,
			PriceCents:  tool.PriceCents,
			Currency:    tool.Currency,
			Lifetime:    tool.Lifetime,
		},
	}, nil
}

func (a *Service) getLicenseCodeDTO(ctx context.Context, id int64) (LicenseCodeDTO, error) {
	var row licenseCodeListRow
	err := a.db.WithContext(ctx).
		Table("license_codes").
		Select(`license_codes.*, tools.name AS tool_name,
			COALESCE(creator.username, '') AS created_by_admin_username,
			COALESCE(redeemer.username, '') AS redeemed_by_username,
			COALESCE(redeemer.email, '') AS redeemed_by_email`).
		Joins("JOIN tools ON tools.code = license_codes.tool_code").
		Joins("LEFT JOIN admins creator ON creator.id = license_codes.created_by_admin_id").
		Joins("LEFT JOIN users redeemer ON redeemer.id = license_codes.redeemed_by_user_id").
		Where("license_codes.id = ?", id).
		Scan(&row).Error
	if err != nil {
		return LicenseCodeDTO{}, err
	}
	return licenseCodeDTO(row), nil
}

func licenseCodeDTO(row licenseCodeListRow) LicenseCodeDTO {
	return LicenseCodeDTO{
		ID:                     row.ID,
		CodeHint:               row.CodeHint,
		ToolCode:               row.ToolCode,
		ToolName:               row.ToolName,
		BatchNo:                row.BatchNo,
		Note:                   row.Note,
		Status:                 row.Status,
		CreatedByAdminUsername: row.CreatedByAdminUsername,
		RedeemedByUserID:       row.RedeemedByUserID,
		RedeemedByUsername:     row.RedeemedByUsername,
		RedeemedByEmail:        row.RedeemedByEmail,
		RedeemedAt:             row.RedeemedAt,
		RevokedAt:              row.RevokedAt,
		CreatedAt:              row.CreatedAt,
	}
}

func newLicenseCode() (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	characters := make([]byte, len(random))
	for i, value := range random {
		characters[i] = licenseCodeAlphabet[int(value)&31]
	}
	return fmt.Sprintf(
		"AT-%s-%s-%s-%s",
		characters[0:4],
		characters[4:8],
		characters[8:12],
		characters[12:16],
	), nil
}

func newLicenseBatchNo() (string, error) {
	random := make([]byte, 5)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return fmt.Sprintf("lc_%d_%s", unixNow(), hex.EncodeToString(random)), nil
}

func normalizeLicenseCode(value string) (string, error) {
	compact := strings.Map(func(r rune) rune {
		switch r {
		case '-', ' ', '\t', '\r', '\n':
			return -1
		default:
			return r
		}
	}, strings.ToUpper(strings.TrimSpace(value)))
	if len(compact) != 18 || !strings.HasPrefix(compact, "AT") {
		return "", errors.New("invalid license code length")
	}
	for _, char := range compact[2:] {
		if !strings.ContainsRune(licenseCodeAlphabet, char) {
			return "", errors.New("invalid license code character")
		}
	}
	return fmt.Sprintf(
		"AT-%s-%s-%s-%s",
		compact[2:6],
		compact[6:10],
		compact[10:14],
		compact[14:18],
	), nil
}

func hashLicenseCode(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func licenseCodeHint(value string) string {
	if len(value) < 4 {
		return "AT-…"
	}
	return "AT-…-" + value[len(value)-4:]
}
