package logic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"automatictools/backend/internal/store"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (a *Service) GetUser(ctx context.Context, userID int64) (UserDTO, error) {
	var record store.User
	err := a.db.WithContext(ctx).
		Where("id = ?", userID).
		Take(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return UserDTO{}, unauthorized("登录状态无效。")
		}
		return UserDTO{}, err
	}

	return userDTO(record), nil
}

func (a *Service) ListTools(ctx context.Context) ([]ToolDTO, error) {
	var records []store.Tool
	err := a.db.WithContext(ctx).
		Where("active = ?", true).
		Order("id").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	tools := make([]ToolDTO, 0, len(records))
	for _, record := range records {
		tools = append(tools, ToolDTO{
			Code:        record.Code,
			Name:        record.Name,
			Description: record.Description,
			PriceCents:  record.PriceCents,
			Currency:    record.Currency,
			Lifetime:    record.Lifetime,
		})
	}
	return tools, nil
}

type CreateOrderRequest struct {
	ToolCode   string `json:"toolCode"`
	PayChannel string `json:"payChannel"`
}

func (a *Service) CreateOrder(ctx context.Context, userID int64, req CreateOrderRequest, meta RequestMeta) (OrderDTO, error) {
	req.ToolCode = strings.TrimSpace(req.ToolCode)
	if req.ToolCode == "" {
		return OrderDTO{}, badRequest("toolCode 不能为空。")
	}
	if req.PayChannel == "" {
		req.PayChannel = "manual"
	}

	var tool store.Tool
	err := a.db.WithContext(ctx).
		Where("code = ? AND active = ?", req.ToolCode, true).
		Take(&tool).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return OrderDTO{}, badRequest("工具不存在或已下架。")
		}
		return OrderDTO{}, err
	}

	now := unixNow()
	order := store.Order{
		OrderNo:     newOrderNo(),
		UserID:      userID,
		ToolCode:    req.ToolCode,
		AmountCents: tool.PriceCents,
		Currency:    tool.Currency,
		Status:      "pending",
		PayChannel:  req.PayChannel,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := a.db.WithContext(ctx).Create(&order).Error; err != nil {
		return OrderDTO{}, err
	}

	_ = a.audit(ctx, meta, &userID, "order.create", order.OrderNo)
	return OrderDTO{
		OrderNo:     order.OrderNo,
		ToolCode:    order.ToolCode,
		AmountCents: order.AmountCents,
		Currency:    order.Currency,
		Status:      order.Status,
		PayChannel:  order.PayChannel,
		CreatedAt:   order.CreatedAt,
	}, nil
}

func (a *Service) ListOrders(ctx context.Context, userID int64) ([]OrderDTO, error) {
	var records []store.Order
	err := a.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id DESC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	orders := make([]OrderDTO, 0, len(records))
	for _, record := range records {
		orders = append(orders, OrderDTO{
			OrderNo:     record.OrderNo,
			ToolCode:    record.ToolCode,
			AmountCents: record.AmountCents,
			Currency:    record.Currency,
			Status:      record.Status,
			PayChannel:  record.PayChannel,
			PaidAt:      record.PaidAt,
			CreatedAt:   record.CreatedAt,
		})
	}
	return orders, nil
}

func (a *Service) ListEntitlements(ctx context.Context, userID int64) ([]EntitlementDTO, error) {
	var records []store.Entitlement
	err := a.db.WithContext(ctx).
		Where("user_id = ? AND (expires_at IS NULL OR expires_at > ?)", userID, unixNow()).
		Order("id").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	result := make([]EntitlementDTO, 0, len(records))
	for _, record := range records {
		result = append(result, EntitlementDTO{
			ToolCode:  record.ToolCode,
			Source:    record.Source,
			ExpiresAt: record.ExpiresAt,
			CreatedAt: record.CreatedAt,
		})
	}
	return result, nil
}

type BindDeviceRequest struct {
	DeviceID   string `json:"deviceId"`
	DeviceName string `json:"deviceName"`
	Platform   string `json:"platform"`
}

func (a *Service) BindDevice(ctx context.Context, userID int64, req BindDeviceRequest) error {
	if strings.TrimSpace(req.DeviceID) == "" {
		return badRequest("deviceId 不能为空。")
	}
	if err := validateOptionalDevice(req.DeviceID, req.DeviceName, req.Platform); err != nil {
		return err
	}
	return a.upsertDevice(ctx, userID, req.DeviceID, req.DeviceName, req.Platform)
}

func (a *Service) upsertDevice(ctx context.Context, userID int64, deviceID string, deviceName string, platform string) error {
	if platform == "" {
		platform = "unknown"
	}
	now := unixNow()
	device := store.Device{
		UserID:     userID,
		DeviceID:   strings.TrimSpace(deviceID),
		DeviceName: strings.TrimSpace(deviceName),
		Platform:   strings.TrimSpace(platform),
		LastSeenAt: now,
		CreatedAt:  now,
	}
	return a.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "device_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"device_name",
				"platform",
				"last_seen_at",
			}),
		}).
		Create(&device).Error
}

type GrantEntitlementRequest struct {
	UserID      int64  `json:"userId"`
	ToolCode    string `json:"toolCode"`
	ProductCode string `json:"productCode,omitempty"`
	Source      string `json:"source"`
	OrderNo     string `json:"orderNo"`
	OrderID     string `json:"orderId,omitempty"`
}

func (a *Service) GrantEntitlement(ctx context.Context, adminID int64, req GrantEntitlementRequest, meta RequestMeta) error {
	if req.ToolCode == "" {
		req.ToolCode = req.ProductCode
	}
	if req.OrderNo == "" {
		req.OrderNo = req.OrderID
	}
	if req.UserID <= 0 || strings.TrimSpace(req.ToolCode) == "" {
		return badRequest("userId 和 toolCode 必填。")
	}
	if req.Source == "" {
		req.Source = "admin"
	}

	entitlement := store.Entitlement{
		UserID:    req.UserID,
		ToolCode:  req.ToolCode,
		Source:    req.Source,
		OrderNo:   req.OrderNo,
		ExpiresAt: nil,
		CreatedAt: unixNow(),
	}
	err := a.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "tool_code"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"source",
				"order_no",
				"expires_at",
			}),
		}).
		Create(&entitlement).Error
	if err != nil {
		return err
	}

	detail, _ := json.Marshal(map[string]any{
		"adminId": adminID,
		"request": req,
	})
	_ = a.audit(ctx, meta, &req.UserID, "entitlement.grant", string(detail))
	return nil
}

type ConfirmOrderRequest struct {
	OrderNo string `json:"orderNo"`
}

func (a *Service) ConfirmOrder(ctx context.Context, adminID int64, req ConfirmOrderRequest, meta RequestMeta) error {
	req.OrderNo = strings.TrimSpace(req.OrderNo)
	if req.OrderNo == "" {
		return badRequest("orderNo 不能为空。")
	}

	var order store.Order
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_no = ?", req.OrderNo).
			Take(&order).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return badRequest("订单不存在。")
		}
		if err != nil {
			return err
		}
		if order.Status == "paid" {
			return nil
		}

		now := unixNow()
		err = tx.Model(&store.Order{}).
			Where("id = ?", order.ID).
			Updates(map[string]any{
				"status":     "paid",
				"paid_at":    now,
				"updated_at": now,
			}).Error
		if err != nil {
			return err
		}

		entitlement := store.Entitlement{
			UserID:    order.UserID,
			ToolCode:  order.ToolCode,
			Source:    "order",
			OrderNo:   req.OrderNo,
			ExpiresAt: nil,
			CreatedAt: now,
		}
		return tx.
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "user_id"}, {Name: "tool_code"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"source",
					"order_no",
					"expires_at",
				}),
			}).
			Create(&entitlement).Error
	})
	if err != nil {
		return err
	}
	if order.Status != "paid" {
		detail, _ := json.Marshal(map[string]any{
			"adminId": adminID,
			"orderNo": req.OrderNo,
		})
		_ = a.audit(ctx, meta, &order.UserID, "order.confirm", string(detail))
	}
	return nil
}

func (a *Service) audit(ctx context.Context, meta RequestMeta, userID *int64, action string, detail string) error {
	record := store.AuditLog{
		UserID:    userID,
		Action:    action,
		IP:        meta.IP,
		UserAgent: meta.UserAgent,
		Detail:    detail,
		CreatedAt: unixNow(),
	}
	return a.db.WithContext(ctx).Create(&record).Error
}

func newOrderNo() string {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "ord_" + fmtSprint(unixNow())
	}
	return "ord_" + hex.EncodeToString(bytes[:])
}
