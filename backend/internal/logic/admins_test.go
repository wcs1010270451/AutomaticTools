package logic

import "testing"

func TestValidateAdminWrite(t *testing.T) {
	tests := []struct {
		name             string
		username         string
		password         string
		status           string
		passwordRequired bool
		wantErr          bool
	}{
		{
			name:             "valid create",
			username:         "admin_ops",
			password:         "123456",
			status:           "active",
			passwordRequired: true,
		},
		{
			name:     "valid update without password",
			username: "admin_ops",
			status:   "disabled",
		},
		{
			name:             "create requires password",
			username:         "admin_ops",
			status:           "active",
			passwordRequired: true,
			wantErr:          true,
		},
		{
			name:     "password too short",
			username: "admin_ops",
			password: "12345",
			status:   "active",
			wantErr:  true,
		},
		{
			name:     "invalid username",
			username: "a",
			password: "123456",
			status:   "active",
			wantErr:  true,
		},
		{
			name:     "invalid status",
			username: "admin_ops",
			password: "123456",
			status:   "locked",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAdminWrite(
				tt.username,
				tt.password,
				tt.status,
				tt.passwordRequired,
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateAdminWrite() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeAdminStatus(t *testing.T) {
	if got := normalizeAdminStatus(""); got != "active" {
		t.Fatalf("unexpected default status: %s", got)
	}
	if got := normalizeAdminStatus(" DISABLED "); got != "disabled" {
		t.Fatalf("unexpected normalized status: %s", got)
	}
}
