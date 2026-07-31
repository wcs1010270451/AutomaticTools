package logic

import (
	"context"
	"strings"

	"automatictools/backend/internal/store"
)

const (
	defaultAdminUserPageSize = 20
	maxAdminUserPageSize     = 100
	maxAdminUserSearchLength = 100
)

type AdminUserListQuery struct {
	Page     int
	PageSize int
	Search   string
	Status   string
}

type AdminUserListResult struct {
	Users    []UserDTO `json:"users"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
}

func (a *Service) ListUsers(ctx context.Context, query AdminUserListQuery) (AdminUserListResult, error) {
	query, err := normalizeAdminUserListQuery(query)
	if err != nil {
		return AdminUserListResult{}, err
	}

	dbQuery := a.db.WithContext(ctx).Model(&store.User{})
	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}
	if query.Search != "" {
		search := "%" + query.Search + "%"
		dbQuery = dbQuery.Where(
			"username ILIKE ? OR email ILIKE ? OR phone ILIKE ?",
			search,
			search,
			search,
		)
	}

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return AdminUserListResult{}, err
	}

	var records []store.User
	offset := (query.Page - 1) * query.PageSize
	if err := dbQuery.
		Order("id DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&records).Error; err != nil {
		return AdminUserListResult{}, err
	}

	users := make([]UserDTO, 0, len(records))
	for _, record := range records {
		users = append(users, userDTO(record))
	}

	return AdminUserListResult{
		Users:    users,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func normalizeAdminUserListQuery(query AdminUserListQuery) (AdminUserListQuery, error) {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = defaultAdminUserPageSize
	}
	if query.Page < 1 {
		return AdminUserListQuery{}, badRequest("page 必须大于 0。")
	}
	if query.PageSize < 1 || query.PageSize > maxAdminUserPageSize {
		return AdminUserListQuery{}, badRequest("pageSize 必须在 1 到 100 之间。")
	}

	query.Search = strings.TrimSpace(query.Search)
	if len([]rune(query.Search)) > maxAdminUserSearchLength {
		return AdminUserListQuery{}, badRequest("search 最多支持 100 个字符。")
	}

	query.Status = strings.ToLower(strings.TrimSpace(query.Status))
	if query.Status != "" && query.Status != "active" && query.Status != "disabled" {
		return AdminUserListQuery{}, badRequest("status 必须是 active 或 disabled。")
	}
	return query, nil
}

func userDTO(record store.User) UserDTO {
	return UserDTO{
		ID:          record.ID,
		Username:    record.Username,
		Email:       stringValue(record.Email),
		Phone:       stringValue(record.Phone),
		Status:      record.Status,
		CreatedAt:   record.CreatedAt,
		LastLoginAt: record.LastLoginAt,
	}
}
