package payment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"automatictools/backend/internal/config"

	alipay "github.com/smartwalle/alipay/v3"
)

var ErrNotConfigured = errors.New("alipay is not configured")

type Alipay struct {
	client    *alipay.Client
	appID     string
	notifyURL string
	sellerID  string
}

func NewAlipay(cfg config.Config) (Provider, error) {
	if !cfg.AlipayEnabled {
		return Disabled{}, nil
	}
	if strings.TrimSpace(cfg.AlipayAppID) == "" ||
		strings.TrimSpace(cfg.AlipayPrivateKeyFile) == "" ||
		strings.TrimSpace(cfg.AlipayPublicKeyFile) == "" ||
		strings.TrimSpace(cfg.AlipayNotifyURL) == "" {
		return nil, errors.New("alipay app id, key files and notify URL are required")
	}
	notifyURL, err := url.ParseRequestURI(strings.TrimSpace(cfg.AlipayNotifyURL))
	if err != nil || notifyURL.Host == "" || (notifyURL.Scheme != "http" && notifyURL.Scheme != "https") {
		return nil, errors.New("alipay notify URL must be an absolute HTTP URL")
	}
	if cfg.AlipayProduction && notifyURL.Scheme != "https" {
		return nil, errors.New("production alipay notify URL must use HTTPS")
	}

	privateKey, err := os.ReadFile(cfg.AlipayPrivateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read alipay private key: %w", err)
	}
	publicKey, err := os.ReadFile(cfg.AlipayPublicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read alipay public key: %w", err)
	}
	client, err := alipay.New(
		strings.TrimSpace(cfg.AlipayAppID),
		string(privateKey),
		cfg.AlipayProduction,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize alipay client: %w", err)
	}
	if err := client.LoadAliPayPublicKey(string(publicKey)); err != nil {
		return nil, fmt.Errorf("load alipay public key: %w", err)
	}
	client.Client = &http.Client{Timeout: time.Duration(cfg.AlipayTimeoutSeconds) * time.Second}

	return &Alipay{
		client:    client,
		appID:     strings.TrimSpace(cfg.AlipayAppID),
		notifyURL: strings.TrimSpace(cfg.AlipayNotifyURL),
		sellerID:  strings.TrimSpace(cfg.AlipaySellerID),
	}, nil
}

func (a *Alipay) Enabled() bool { return true }

func (a *Alipay) Precreate(ctx context.Context, req PrecreateRequest) (PrecreateResult, error) {
	param := alipay.TradePreCreate{}
	param.NotifyURL = a.notifyURL
	param.Subject = req.Subject
	param.Body = req.Body
	param.OutTradeNo = req.OrderNo
	param.TotalAmount = formatAmount(req.AmountCents)
	param.ProductCode = "FACE_TO_FACE_PAYMENT"
	param.TimeoutExpress = "10m"

	response, err := a.client.TradePreCreate(ctx, param)
	if err != nil {
		return PrecreateResult{}, fmt.Errorf("alipay precreate request: %w", err)
	}
	if response == nil || response.IsFailure() {
		if response == nil {
			return PrecreateResult{}, errors.New("alipay returned an empty precreate response")
		}
		return PrecreateResult{}, fmt.Errorf("alipay precreate failed: %s", response.Error.Error())
	}
	if strings.TrimSpace(response.QRCode) == "" {
		return PrecreateResult{}, errors.New("alipay precreate response has no QR code")
	}
	return PrecreateResult{QRCode: response.QRCode}, nil
}

func (a *Alipay) Query(ctx context.Context, orderNo string) (TradeResult, error) {
	response, err := a.client.TradeQuery(ctx, alipay.TradeQuery{OutTradeNo: orderNo})
	if err != nil {
		return TradeResult{}, fmt.Errorf("alipay trade query: %w", err)
	}
	if response == nil || response.IsFailure() {
		if response == nil {
			return TradeResult{}, errors.New("alipay returned an empty query response")
		}
		return TradeResult{}, fmt.Errorf("alipay query failed: %s", response.Error.Error())
	}
	return TradeResult{
		OrderNo:         response.OutTradeNo,
		ProviderTradeNo: response.TradeNo,
		Status:          string(response.TradeStatus),
		Amount:          response.TotalAmount,
		AppID:           a.appID,
		SellerID:        a.sellerID,
	}, nil
}

func (a *Alipay) DecodeNotification(ctx context.Context, values url.Values) (TradeResult, error) {
	notification, err := a.client.DecodeNotification(ctx, values)
	if err != nil {
		return TradeResult{}, fmt.Errorf("verify alipay notification: %w", err)
	}
	if notification.AppId != a.appID {
		return TradeResult{}, errors.New("alipay notification app id mismatch")
	}
	if a.sellerID != "" && notification.SellerId != a.sellerID {
		return TradeResult{}, errors.New("alipay notification seller id mismatch")
	}
	return TradeResult{
		OrderNo:         notification.OutTradeNo,
		ProviderTradeNo: notification.TradeNo,
		Status:          string(notification.TradeStatus),
		Amount:          notification.TotalAmount,
		AppID:           notification.AppId,
		SellerID:        notification.SellerId,
	}, nil
}

func formatAmount(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}
