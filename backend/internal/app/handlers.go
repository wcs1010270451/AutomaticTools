package app

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *App) me(w http.ResponseWriter, r *http.Request) {
	userID, _, err := a.currentUser(r)
	if err != nil {
		a.handleErr(w, r, err)
		return
	}

	row := a.db.QueryRowContext(
		r.Context(),
		`SELECT id, username, email, phone, status, created_at, last_login_at FROM users WHERE id = $1`,
		userID,
	)

	var user UserDTO
	var email sql.NullString
	var phone sql.NullString
	var lastLogin sql.NullInt64
	if err := row.Scan(&user.ID, &user.Username, &email, &phone, &user.Status, &user.CreatedAt, &lastLogin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.handleErr(w, r, unauthorized("登录状态无效。"))
			return
		}
		a.handleErr(w, r, err)
		return
	}
	user.Email = email.String
	user.Phone = phone.String
	if lastLogin.Valid {
		user.LastLoginAt = &lastLogin.Int64
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (a *App) products(w http.ResponseWriter, r *http.Request) {
	a.tools(w, r)
}

func (a *App) tools(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(
		r.Context(),
		`SELECT code, name, description, price_cents, currency, lifetime
		 FROM tools
		 WHERE active = TRUE
		 ORDER BY id`,
	)
	if err != nil {
		a.handleErr(w, r, err)
		return
	}
	defer rows.Close()

	tools := make([]ToolDTO, 0)
	for rows.Next() {
		var tool ToolDTO
		if err := rows.Scan(&tool.Code, &tool.Name, &tool.Description, &tool.PriceCents, &tool.Currency, &tool.Lifetime); err != nil {
			a.handleErr(w, r, err)
			return
		}
		tools = append(tools, tool)
	}

	writeJSON(w, http.StatusOK, map[string]any{"tools": tools})
}

type createOrderRequest struct {
	ToolCode   string `json:"toolCode"`
	PayChannel string `json:"payChannel"`
}

func (a *App) createOrder(w http.ResponseWriter, r *http.Request) {
	userID, _, err := a.currentUser(r)
	if err != nil {
		a.handleErr(w, r, err)
		return
	}

	var req createOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		a.handleErr(w, r, badRequest("请求体必须是合法 JSON。"))
		return
	}
	req.ToolCode = strings.TrimSpace(req.ToolCode)
	if req.ToolCode == "" {
		a.handleErr(w, r, badRequest("toolCode 不能为空。"))
		return
	}
	if req.PayChannel == "" {
		req.PayChannel = "manual"
	}

	var priceCents int64
	var currency string
	err = a.db.QueryRowContext(
		r.Context(),
		`SELECT price_cents, currency FROM tools WHERE code = $1 AND active = TRUE`,
		req.ToolCode,
	).Scan(&priceCents, &currency)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.handleErr(w, r, badRequest("工具不存在或已下架。"))
			return
		}
		a.handleErr(w, r, err)
		return
	}

	orderNo := newOrderNo()
	now := unixNow()
	_, err = a.db.ExecContext(
		r.Context(),
		`INSERT INTO orders(order_no, user_id, tool_code, amount_cents, currency, status, pay_channel, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, 'pending', $6, $7, $8)`,
		orderNo,
		userID,
		req.ToolCode,
		priceCents,
		currency,
		req.PayChannel,
		now,
		now,
	)
	if err != nil {
		a.handleErr(w, r, err)
		return
	}

	_ = a.audit(r, &userID, "order.create", orderNo)
	writeJSON(w, http.StatusCreated, map[string]any{
		"order": OrderDTO{
			OrderNo:     orderNo,
			ToolCode:    req.ToolCode,
			AmountCents: priceCents,
			Currency:    currency,
			Status:      "pending",
			PayChannel:  req.PayChannel,
			CreatedAt:   now,
		},
	})
}

func (a *App) myOrders(w http.ResponseWriter, r *http.Request) {
	userID, _, err := a.currentUser(r)
	if err != nil {
		a.handleErr(w, r, err)
		return
	}

	orders, err := a.ordersForUser(r, userID)
	if err != nil {
		a.handleErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"orders": orders})
}

func (a *App) ordersForUser(r *http.Request, userID int64) ([]OrderDTO, error) {
	rows, err := a.db.QueryContext(
		r.Context(),
		`SELECT order_no, tool_code, amount_cents, currency, status, pay_channel, paid_at, created_at
		 FROM orders
		 WHERE user_id = $1
		 ORDER BY id DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]OrderDTO, 0)
	for rows.Next() {
		var order OrderDTO
		var paidAt sql.NullInt64
		if err := rows.Scan(&order.OrderNo, &order.ToolCode, &order.AmountCents, &order.Currency, &order.Status, &order.PayChannel, &paidAt, &order.CreatedAt); err != nil {
			return nil, err
		}
		if paidAt.Valid {
			order.PaidAt = &paidAt.Int64
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (a *App) myEntitlements(w http.ResponseWriter, r *http.Request) {
	userID, _, err := a.currentUser(r)
	if err != nil {
		a.handleErr(w, r, err)
		return
	}

	entitlements, err := a.entitlementsForUser(r, userID)
	if err != nil {
		a.handleErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"entitlements": entitlements})
}

func (a *App) entitlementsForUser(r *http.Request, userID int64) ([]EntitlementDTO, error) {
	rows, err := a.db.QueryContext(
		r.Context(),
		`SELECT tool_code, source, expires_at, created_at
		 FROM entitlements
		 WHERE user_id = $1 AND (expires_at IS NULL OR expires_at > $2)
		 ORDER BY id`,
		userID,
		unixNow(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]EntitlementDTO, 0)
	for rows.Next() {
		var item EntitlementDTO
		var expiresAt sql.NullInt64
		if err := rows.Scan(&item.ToolCode, &item.Source, &expiresAt, &item.CreatedAt); err != nil {
			return nil, err
		}
		if expiresAt.Valid {
			item.ExpiresAt = &expiresAt.Int64
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

type bindDeviceRequest struct {
	DeviceID   string `json:"deviceId"`
	DeviceName string `json:"deviceName"`
	Platform   string `json:"platform"`
}

func (a *App) bindDevice(w http.ResponseWriter, r *http.Request) {
	userID, _, err := a.currentUser(r)
	if err != nil {
		a.handleErr(w, r, err)
		return
	}

	var req bindDeviceRequest
	if err := decodeJSON(r, &req); err != nil {
		a.handleErr(w, r, badRequest("请求体必须是合法 JSON。"))
		return
	}
	if strings.TrimSpace(req.DeviceID) == "" {
		a.handleErr(w, r, badRequest("deviceId 不能为空。"))
		return
	}

	if err := a.upsertDevice(r, userID, req.DeviceID, req.DeviceName, req.Platform); err != nil {
		a.handleErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *App) upsertDevice(r *http.Request, userID int64, deviceID string, deviceName string, platform string) error {
	if platform == "" {
		platform = "unknown"
	}
	now := unixNow()
	_, err := a.db.ExecContext(
		r.Context(),
		`INSERT INTO devices(user_id, device_id, device_name, platform, last_seen_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT(user_id, device_id)
		 DO UPDATE SET device_name = excluded.device_name, platform = excluded.platform, last_seen_at = excluded.last_seen_at`,
		userID,
		strings.TrimSpace(deviceID),
		strings.TrimSpace(deviceName),
		strings.TrimSpace(platform),
		now,
		now,
	)
	return err
}

type grantEntitlementRequest struct {
	UserID      int64  `json:"userId"`
	ToolCode    string `json:"toolCode"`
	ProductCode string `json:"productCode,omitempty"`
	Source      string `json:"source"`
	OrderNo     string `json:"orderNo"`
	OrderID     string `json:"orderId,omitempty"`
}

func (a *App) adminGrantEntitlement(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Admin-Key") != a.cfg.AdminKey {
		a.handleErr(w, r, forbidden("管理员密钥无效。"))
		return
	}

	var req grantEntitlementRequest
	if err := decodeJSON(r, &req); err != nil {
		a.handleErr(w, r, badRequest("请求体必须是合法 JSON。"))
		return
	}
	if req.ToolCode == "" {
		req.ToolCode = req.ProductCode
	}
	if req.OrderNo == "" {
		req.OrderNo = req.OrderID
	}
	if req.UserID <= 0 || strings.TrimSpace(req.ToolCode) == "" {
		a.handleErr(w, r, badRequest("userId 和 toolCode 必填。"))
		return
	}
	if req.Source == "" {
		req.Source = "admin"
	}

	now := unixNow()
	_, err := a.db.ExecContext(
		r.Context(),
		`INSERT INTO entitlements(user_id, tool_code, source, order_no, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, NULL, $5)
		 ON CONFLICT(user_id, tool_code)
		 DO UPDATE SET source = excluded.source, order_no = excluded.order_no, expires_at = NULL`,
		req.UserID,
		req.ToolCode,
		req.Source,
		req.OrderNo,
		now,
	)
	if err != nil {
		a.handleErr(w, r, err)
		return
	}

	detail, _ := json.Marshal(req)
	_ = a.audit(r, &req.UserID, "entitlement.grant", string(detail))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type confirmOrderRequest struct {
	OrderNo string `json:"orderNo"`
}

func (a *App) adminConfirmOrder(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Admin-Key") != a.cfg.AdminKey {
		a.handleErr(w, r, forbidden("管理员密钥无效。"))
		return
	}

	var req confirmOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		a.handleErr(w, r, badRequest("请求体必须是合法 JSON。"))
		return
	}
	req.OrderNo = strings.TrimSpace(req.OrderNo)
	if req.OrderNo == "" {
		a.handleErr(w, r, badRequest("orderNo 不能为空。"))
		return
	}

	var userID int64
	var toolCode string
	var status string
	err := a.db.QueryRowContext(
		r.Context(),
		`SELECT user_id, tool_code, status FROM orders WHERE order_no = $1`,
		req.OrderNo,
	).Scan(&userID, &toolCode, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.handleErr(w, r, badRequest("订单不存在。"))
			return
		}
		a.handleErr(w, r, err)
		return
	}
	if status == "paid" {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	now := unixNow()
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.handleErr(w, r, err)
		return
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(r.Context(), `UPDATE orders SET status = 'paid', paid_at = $1, updated_at = $2 WHERE order_no = $3`, now, now, req.OrderNo); err != nil {
		a.handleErr(w, r, err)
		return
	}
	if _, err := tx.ExecContext(
		r.Context(),
		`INSERT INTO entitlements(user_id, tool_code, source, order_no, expires_at, created_at)
		 VALUES ($1, $2, 'order', $3, NULL, $4)
		 ON CONFLICT(user_id, tool_code)
		 DO UPDATE SET source = excluded.source, order_no = excluded.order_no, expires_at = NULL`,
		userID,
		toolCode,
		req.OrderNo,
		now,
	); err != nil {
		a.handleErr(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.handleErr(w, r, err)
		return
	}

	_ = a.audit(r, &userID, "order.confirm", req.OrderNo)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *App) audit(r *http.Request, userID *int64, action string, detail string) error {
	var id any
	if userID != nil {
		id = *userID
	}
	_, err := a.db.ExecContext(
		r.Context(),
		`INSERT INTO audit_logs(user_id, action, ip, user_agent, detail, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id,
		action,
		clientIP(r),
		r.UserAgent(),
		detail,
		unixNow(),
	)
	return err
}

func newOrderNo() string {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "ord_" + fmtSprint(unixNow())
	}
	return "ord_" + hex.EncodeToString(bytes[:])
}
