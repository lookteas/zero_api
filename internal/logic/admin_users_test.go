package logic

import (
	"testing"

	"api/internal/types"
)

func TestNormalizeAdminUserQueryDefaults(t *testing.T) {
	query := normalizeAdminUserQuery(types.AdminUserQueryReq{})

	if query.Page != 1 {
		t.Fatalf("expected default page 1, got %d", query.Page)
	}
	if query.PageSize != 20 {
		t.Fatalf("expected default page size 20, got %d", query.PageSize)
	}
	if query.Status != "" {
		t.Fatalf("expected default status empty, got %q", query.Status)
	}
}

func TestNormalizeAdminUserProfileInputRejectsEmptyNickname(t *testing.T) {
	_, err := normalizeAdminUserProfileInput(&types.AdminUpdateUserReq{
		Nickname: "   ",
		Status:   1,
	})

	if err == nil {
		t.Fatal("expected empty nickname to be rejected")
	}
}

func TestNormalizeAdminUserProfileInputTrimsOptionalFields(t *testing.T) {
	input, err := normalizeAdminUserProfileInput(&types.AdminUpdateUserReq{
		Nickname: "  Max  ",
		Email:    "  demo@example.com  ",
		Mobile:   " 13800000000 ",
		Avatar:   "  https://example.com/avatar.png  ",
		Status:   1,
	})

	if err != nil {
		t.Fatalf("expected valid profile input, got %v", err)
	}
	if input.Nickname != "Max" {
		t.Fatalf("expected trimmed nickname, got %q", input.Nickname)
	}
	if input.Email != "demo@example.com" {
		t.Fatalf("expected trimmed email, got %q", input.Email)
	}
	if input.Mobile != "13800000000" {
		t.Fatalf("expected trimmed mobile, got %q", input.Mobile)
	}
	if input.Avatar != "https://example.com/avatar.png" {
		t.Fatalf("expected trimmed avatar, got %q", input.Avatar)
	}
}
