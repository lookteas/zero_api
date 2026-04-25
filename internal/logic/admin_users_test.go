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

func TestApplyAdminUserTopicCountsAttachesCountsToMatchingUsers(t *testing.T) {
	users := []types.AdminUserInfo{
		{Id: 1, Account: "alice"},
		{Id: 2, Account: "bob"},
	}
	counts := map[uint64]int64{
		1: 3,
	}

	applyAdminUserTopicCounts(users, counts)

	if users[0].TopicCount != 3 {
		t.Fatalf("expected alice topic count 3, got %d", users[0].TopicCount)
	}
	if users[1].TopicCount != 0 {
		t.Fatalf("expected bob topic count 0, got %d", users[1].TopicCount)
	}
}
