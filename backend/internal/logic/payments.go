package logic

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"automatictools/backend/internal/payment"
	"automatictools/backend/internal/store"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const paymentQRCodeTTL = 10 * time.Minute

type CreateAlipayPaymentRequest struct {
	ToolCode string `json:"toolCode"`
}

type PaymentSessionDTO struct {
	Order     OrderDTO `json:"order"`
	Channel   string   `json:"channel"`
	QRCode    string   `json:"qrCode"`
	ExpiresAt int64    `json:"expiresAt"`
}

func (a *Service) ListPurchases(ctx context.Context, userID int64) ([]PurchaseDTO, error) {
	var records []store.Entitlement
	err := a.db.WithContext(ctx).
		Where("user_id = ? AND (expires_at IS NULL OR expires_at > ?)", userID, unixNow()).
		Order("id").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	result := make([]PurchaseDTO, 0, len(records))
	for _, record := range records {
		result = append(result, PurchaseDTO{
			ToolCode:    record.ToolCode,
			OrderNo:     record.OrderNo,
			PurchasedAt: record.CreatedAt,
		})
	}
	return result, nil
}

func (a *Service) CreateAlipayPayment(
	ctx context.Context,
	userID int64,
	req CreateAlipayPaymentRequest,
	meta RequestMeta,
) (PaymentSessionDTO, error) {
	if !a.payment.Enabled() {
		return PaymentSessionDTO{}, serviceUnavailable("支付宝支付尚未配置。", payment.ErrNotConfigured)
	}
	req.ToolCode = strings.TrimSpace(req.ToolCode)
	if req.ToolCode == "" {
		return PaymentSessionDTO{}, badRequest("toolCode 不能为空。")
	}

	var tool store.Tool
	var order store.Order
	returnExisting := false
	now := unixNow()
	expiresAt := now + int64(paymentQRCodeTTL/time.Second)
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user store.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", userID).
			Take(&user).Error; err != nil {
			return err
		}

		var purchaseCount int64
		if err := tx.Model(&store.Entitlement{}).
			Where("user_id = ? AND tool_code = ? AND (expires_at IS NULL OR expires_at > ?)", userID, req.ToolCode, now).
			Count(&purchaseCount).Error; err != nil {
			return err
		}
		if purchaseCount > 0 {
			return conflict("该工具已经购买，无需重复支付。")
		}

		if err := tx.Where("code = ? AND active = ?", req.ToolCode, true).
			Take(&tool).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return badRequest("工具不存在或已下架。")
			}
			return err
		}
		if tool.Currency != "CNY" || tool.PriceCents <= 0 {
			return badRequest("该工具暂不支持支付宝购买。")
		}

		err := tx.Where(
			"user_id = ? AND tool_code = ? AND pay_channel = ? AND status = ? AND payment_expires_at > ?",
			userID,
			req.ToolCode,
			"alipay",
			"pending",
			now,
		).Order("id DESC").Take(&order).Error
		if err == nil {
			if order.PaymentQRCode == "" {
				return conflict("支付二维码正在生成，请稍后重试。")
			}
			returnExisting = true
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := tx.Model(&store.Order{}).
			Where("user_id = ? AND tool_code = ? AND pay_channel = ? AND status = ?", userID, req.ToolCode, "alipay", "pending").
			Updates(map[string]any{"status": "cancelled", "updated_at": now}).Error; err != nil {
			return err
		}

		order = store.Order{
			OrderNo:          newOrderNo(),
			UserID:           userID,
			ToolCode:         tool.Code,
			AmountCents:      tool.PriceCents,
			Currency:         tool.Currency,
			Status:           "pending",
			PayChannel:       "alipay",
			PaymentExpiresAt: &expiresAt,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		return tx.Create(&order).Error
	})
	if err != nil {
		return PaymentSessionDTO{}, err
	}
	if returnExisting {
		return paymentSessionDTO(order), nil
	}

	created, err := a.payment.Precreate(ctx, payment.PrecreateRequest{
		OrderNo:     order.OrderNo,
		Subject:     tool.Name,
		Body:        tool.Description,
		AmountCents: order.AmountCents,
	})
	if err != nil {
		_ = a.db.WithContext(ctx).Model(&store.Order{}).
			Where("id = ? AND status = ?", order.ID, "pending").
			Updates(map[string]any{"status": "failed", "updated_at": unixNow()}).Error
		return PaymentSessionDTO{}, serviceUnavailable("创建支付宝二维码失败，请稍后重试。", err)
	}
	order.PaymentQRCode = created.QRCode
	if err := a.db.WithContext(ctx).Model(&store.Order{}).
		Where("id = ? AND status = ?", order.ID, "pending").
		Updates(map[string]any{
			"payment_qr_code": created.QRCode,
			"updated_at":      unixNow(),
		}).Error; err != nil {
		return PaymentSessionDTO{}, err
	}
	_ = a.audit(ctx, meta, &userID, "payment.alipay.precreate", order.OrderNo)
	return paymentSessionDTO(order), nil
}

func (a *Service) GetPaymentOrder(ctx context.Context, userID int64, orderNo string) (OrderDTO, error) {
	orderNo = strings.TrimSpace(orderNo)
	if orderNo == "" {
		return OrderDTO{}, badRequest("orderNo 不能为空。")
	}
	var order store.Order
	err := a.db.WithContext(ctx).
		Where("order_no = ? AND user_id = ?", orderNo, userID).
		Take(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return OrderDTO{}, notFound("订单不存在。")
	}
	if err != nil {
		return OrderDTO{}, err
	}
	return orderDTO(order), nil
}

func (a *Service) HandleAlipayNotification(ctx context.Context, values url.Values, meta RequestMeta) error {
	if !a.payment.Enabled() {
		return serviceUnavailable("支付宝支付尚未配置。", payment.ErrNotConfigured)
	}
	result, err := a.payment.DecodeNotification(ctx, values)
	if err != nil {
		return badRequest("支付宝通知验签失败。")
	}
	return a.applyAlipayTrade(ctx, result, meta)
}

func (a *Service) applyAlipayTrade(ctx context.Context, result payment.TradeResult, meta RequestMeta) error {
	if strings.TrimSpace(result.OrderNo) == "" {
		return badRequest("支付宝通知缺少商户订单号。")
	}
	if result.Status != "TRADE_SUCCESS" && result.Status != "TRADE_FINISHED" {
		return nil
	}
	if strings.TrimSpace(result.ProviderTradeNo) == "" {
		return badRequest("支付宝通知缺少交易号。")
	}
	amountCents, err := parseAmountCents(result.Amount)
	if err != nil {
		return badRequest("支付宝通知金额无效。")
	}

	var order store.Order
	wasPaid := false
	err = a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_no = ?", result.OrderNo).
			Take(&order).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return notFound("支付宝通知对应的订单不存在。")
		}
		if err != nil {
			return err
		}
		if order.PayChannel != "alipay" || order.Currency != "CNY" {
			return badRequest("订单支付渠道或币种不匹配。")
		}
		if order.AmountCents != amountCents {
			return badRequest("支付宝通知金额与订单金额不一致。")
		}
		if order.ProviderTradeNo != "" && order.ProviderTradeNo != result.ProviderTradeNo {
			return badRequest("支付宝交易号与订单记录不一致。")
		}
		if order.Status == "paid" {
			wasPaid = true
			return nil
		}

		now := unixNow()
		if err := tx.Model(&store.Order{}).
			Where("id = ?", order.ID).
			Updates(map[string]any{
				"status":            "paid",
				"provider_trade_no": result.ProviderTradeNo,
				"payment_qr_code":   "",
				"paid_at":           now,
				"updated_at":        now,
			}).Error; err != nil {
			return err
		}

		entitlement := store.Entitlement{
			UserID:    order.UserID,
			ToolCode:  order.ToolCode,
			Source:    "alipay",
			OrderNo:   order.OrderNo,
			ExpiresAt: nil,
			CreatedAt: now,
		}
		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "tool_code"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"source",
				"order_no",
				"expires_at",
			}),
		}).Create(&entitlement).Error
	})
	if err != nil {
		return err
	}
	if !wasPaid {
		_ = a.audit(ctx, meta, &order.UserID, "payment.alipay.paid", order.OrderNo)
	}
	return nil
}

func paymentSessionDTO(order store.Order) PaymentSessionDTO {
	expiresAt := int64(0)
	if order.PaymentExpiresAt != nil {
		expiresAt = *order.PaymentExpiresAt
	}
	return PaymentSessionDTO{
		Order:     orderDTO(order),
		Channel:   "alipay",
		QRCode:    order.PaymentQRCode,
		ExpiresAt: expiresAt,
	}
}

func orderDTO(order store.Order) OrderDTO {
	return OrderDTO{
		OrderNo:          order.OrderNo,
		ToolCode:         order.ToolCode,
		AmountCents:      order.AmountCents,
		Currency:         order.Currency,
		Status:           order.Status,
		PayChannel:       order.PayChannel,
		PaymentExpiresAt: order.PaymentExpiresAt,
		PaidAt:           order.PaidAt,
		CreatedAt:        order.CreatedAt,
	}
}

func parseAmountCents(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "-") {
		return 0, errors.New("invalid amount")
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, errors.New("invalid amount")
	}
	yuan, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if len(fraction) > 2 {
		return 0, errors.New("amount has more than two decimal places")
	}
	for len(fraction) < 2 {
		fraction += "0"
	}
	fen := int64(0)
	if fraction != "" {
		fen, err = strconv.ParseInt(fraction, 10, 64)
		if err != nil {
			return 0, err
		}
	}
	if yuan > (1<<63-1-fen)/100 {
		return 0, errors.New("amount overflow")
	}
	return yuan*100 + fen, nil
}
