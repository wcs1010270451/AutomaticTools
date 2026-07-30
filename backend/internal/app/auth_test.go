package app

import "testing"

func TestNormalizeContacts(t *testing.T) {
	if got := normalizeEmail("  User@Example.COM "); got != "user@example.com" {
		t.Fatalf("unexpected normalized email: %q", got)
	}
	if got := normalizePhone(" +86 138-0013-8000 "); got != "+8613800138000" {
		t.Fatalf("unexpected normalized phone: %q", got)
	}
}

func TestValidateContacts(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		phone   string
		wantErr bool
	}{
		{name: "empty contacts are allowed"},
		{name: "valid contacts", email: "user@example.com", phone: "+8613800138000"},
		{name: "invalid email", email: "not-an-email", wantErr: true},
		{name: "phone too short", phone: "12345", wantErr: true},
		{name: "phone contains letters", phone: "13800abc000", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContacts(tt.email, tt.phone)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateContacts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNullableString(t *testing.T) {
	if nullableString("") != nil {
		t.Fatal("empty string should be stored as NULL")
	}
	if got := nullableString("value"); got != "value" {
		t.Fatalf("unexpected non-empty value: %v", got)
	}
}
