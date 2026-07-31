package logic

import (
	"strings"
	"testing"
)

func TestValidateToolWrite(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		toolName    string
		description string
		priceCents  int64
		currency    string
		wantErr     bool
	}{
		{name: "valid", code: "auto_click", toolName: "自动点击", priceCents: 1000, currency: "CNY"},
		{name: "free tool", code: "timer", toolName: "计时器", priceCents: 0, currency: "CNY"},
		{name: "invalid code", code: "Auto-Click", toolName: "自动点击", priceCents: 1000, currency: "CNY", wantErr: true},
		{name: "empty name", code: "auto_click", toolName: "", priceCents: 1000, currency: "CNY", wantErr: true},
		{name: "long name", code: "auto_click", toolName: strings.Repeat("工", 101), priceCents: 1000, currency: "CNY", wantErr: true},
		{name: "long description", code: "auto_click", toolName: "自动点击", description: strings.Repeat("工", 2001), priceCents: 1000, currency: "CNY", wantErr: true},
		{name: "negative price", code: "auto_click", toolName: "自动点击", priceCents: -1, currency: "CNY", wantErr: true},
		{name: "invalid currency", code: "auto_click", toolName: "自动点击", priceCents: 1000, currency: "CN", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateToolWrite(
				tt.code,
				tt.toolName,
				tt.description,
				tt.priceCents,
				tt.currency,
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateToolWrite() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeCurrency(t *testing.T) {
	if got := normalizeCurrency(" cny "); got != "CNY" {
		t.Fatalf("normalizeCurrency() = %q", got)
	}
	if got := normalizeCurrency(""); got != "CNY" {
		t.Fatalf("normalizeCurrency() default = %q", got)
	}
}
