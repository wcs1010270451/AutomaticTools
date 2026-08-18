package logic

import (
	"strings"
	"testing"
)

func TestNewAndNormalizeLicenseCode(t *testing.T) {
	code, err := newLicenseCode()
	if err != nil {
		t.Fatalf("newLicenseCode() error = %v", err)
	}
	if len(code) != 22 || !strings.HasPrefix(code, "AT-") {
		t.Fatalf("unexpected code format: %q", code)
	}

	normalized, err := normalizeLicenseCode(strings.ToLower(strings.ReplaceAll(code, "-", " ")))
	if err != nil {
		t.Fatalf("normalizeLicenseCode() error = %v", err)
	}
	if normalized != code {
		t.Fatalf("normalized = %q, want %q", normalized, code)
	}
	if hint := licenseCodeHint(code); !strings.HasSuffix(hint, code[len(code)-4:]) {
		t.Fatalf("unexpected hint: %q", hint)
	}
	if len(hashLicenseCode(code)) != 64 {
		t.Fatal("license code hash should be SHA-256 hex")
	}
}

func TestNormalizeLicenseCodeRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{
		"",
		"AT-1234",
		"XX-2345-6789-ABCD-EFGH",
		"AT-OOOO-OOOO-OOOO-OOOO",
	} {
		if _, err := normalizeLicenseCode(value); err == nil {
			t.Fatalf("normalizeLicenseCode(%q) should fail", value)
		}
	}
}
