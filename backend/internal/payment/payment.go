package payment

import (
	"context"
	"net/url"
)

type PrecreateRequest struct {
	OrderNo     string
	Subject     string
	Body        string
	AmountCents int64
}

type PrecreateResult struct {
	QRCode string
}

type TradeResult struct {
	OrderNo         string
	ProviderTradeNo string
	Status          string
	Amount          string
	AppID           string
	SellerID        string
}

type Provider interface {
	Enabled() bool
	Precreate(context.Context, PrecreateRequest) (PrecreateResult, error)
	Query(context.Context, string) (TradeResult, error)
	DecodeNotification(context.Context, url.Values) (TradeResult, error)
}

type Disabled struct{}

func (Disabled) Enabled() bool { return false }

func (Disabled) Precreate(context.Context, PrecreateRequest) (PrecreateResult, error) {
	return PrecreateResult{}, ErrNotConfigured
}

func (Disabled) Query(context.Context, string) (TradeResult, error) {
	return TradeResult{}, ErrNotConfigured
}

func (Disabled) DecodeNotification(context.Context, url.Values) (TradeResult, error) {
	return TradeResult{}, ErrNotConfigured
}
