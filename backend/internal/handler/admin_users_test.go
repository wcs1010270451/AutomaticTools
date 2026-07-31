package handler

import "testing"

func TestParsePositiveQueryInt(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		defaultValue int
		want         int
		wantErr      bool
	}{
		{name: "default", defaultValue: 20, want: 20},
		{name: "value", raw: "5", defaultValue: 20, want: 5},
		{name: "zero", raw: "0", defaultValue: 20, wantErr: true},
		{name: "invalid", raw: "abc", defaultValue: 20, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePositiveQueryInt(tt.raw, "page", tt.defaultValue)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}
