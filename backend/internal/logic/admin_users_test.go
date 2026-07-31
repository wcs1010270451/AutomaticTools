package logic

import (
	"strings"
	"testing"
)

func TestNormalizeAdminUserListQueryDefaults(t *testing.T) {
	query, err := normalizeAdminUserListQuery(AdminUserListQuery{
		Search: "  test  ",
		Status: " ACTIVE ",
	})
	if err != nil {
		t.Fatalf("normalize query: %v", err)
	}
	if query.Page != 1 || query.PageSize != 20 {
		t.Fatalf("unexpected pagination: %+v", query)
	}
	if query.Search != "test" || query.Status != "active" {
		t.Fatalf("unexpected filters: %+v", query)
	}
}

func TestNormalizeAdminUserListQueryRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		query AdminUserListQuery
	}{
		{name: "negative page", query: AdminUserListQuery{Page: -1}},
		{name: "page size too large", query: AdminUserListQuery{PageSize: 101}},
		{name: "invalid status", query: AdminUserListQuery{Status: "pending"}},
		{name: "search too long", query: AdminUserListQuery{Search: strings.Repeat("a", 101)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := normalizeAdminUserListQuery(tt.query); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
