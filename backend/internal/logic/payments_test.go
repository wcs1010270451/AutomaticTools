package logic

import "testing"

func TestParseAmountCents(t *testing.T) {
	tests := []struct {
		value   string
		want    int64
		wantErr bool
	}{
		{value: "10", want: 1000},
		{value: "10.5", want: 1050},
		{value: "10.50", want: 1050},
		{value: "0.01", want: 1},
		{value: "10.001", wantErr: true},
		{value: "-1.00", wantErr: true},
		{value: "", wantErr: true},
		{value: "abc", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := parseAmountCents(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseAmountCents(%q) error = %v", tt.value, err)
			}
			if err == nil && got != tt.want {
				t.Fatalf("parseAmountCents(%q) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}
