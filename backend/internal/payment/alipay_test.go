package payment

import (
	"testing"

	"automatictools/backend/internal/config"
)

func TestFormatAmount(t *testing.T) {
	if got := formatAmount(1000); got != "10.00" {
		t.Fatalf("formatAmount(1000) = %q", got)
	}
	if got := formatAmount(1); got != "0.01" {
		t.Fatalf("formatAmount(1) = %q", got)
	}
}

func TestAlipayDisabledByDefault(t *testing.T) {
	provider, err := NewAlipay(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if provider.Enabled() {
		t.Fatal("provider must stay disabled without explicit configuration")
	}
}

func TestAlipayEnabledRequiresConfiguration(t *testing.T) {
	if _, err := NewAlipay(config.Config{AlipayEnabled: true}); err == nil {
		t.Fatal("enabled provider must reject missing merchant configuration")
	}
}

func TestAlipayProductionRequiresHTTPSNotifyURL(t *testing.T) {
	_, err := NewAlipay(config.Config{
		AlipayEnabled:        true,
		AlipayAppID:          "app-id",
		AlipayPrivateKeyFile: "private.pem",
		AlipayPublicKeyFile:  "public.pem",
		AlipayNotifyURL:      "http://example.com/api/payments/alipay/notify",
		AlipayProduction:     true,
	})
	if err == nil {
		t.Fatal("production provider must reject a non-HTTPS notify URL")
	}
}
