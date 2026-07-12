package logic

import (
	"context"
	"testing"

	"api/internal/svc"
	"api/internal/types"
	"api/model"
)

type passwordChangeUsersModel struct {
	model.UsersModel
	user    *model.Users
	updated bool
}

func (m *passwordChangeUsersModel) FindOne(_ context.Context, id uint64) (*model.Users, error) {
	if m.user == nil || m.user.Id != id {
		return nil, model.ErrNotFound
	}
	return m.user, nil
}

func (m *passwordChangeUsersModel) Update(_ context.Context, user *model.Users) error {
	m.user = user
	m.updated = true
	return nil
}

func TestChangePasswordValidatesAndUpdatesCurrentUser(t *testing.T) {
	store := &passwordChangeUsersModel{user: &model.Users{Id: 7, PasswordHash: hashPassword("old-pass")}}
	logic := NewChangePasswordLogic(WithCurrentUserID(context.Background(), 7), &svc.ServiceContext{UsersModel: store})

	if _, err := logic.ChangePassword(&types.ChangePasswordReq{CurrentPassword: "wrong-pass", NewPassword: "new-pass"}); err == nil {
		t.Fatal("expected current password validation error")
	}
	if _, err := logic.ChangePassword(&types.ChangePasswordReq{CurrentPassword: "old-pass", NewPassword: "old-pass"}); err == nil {
		t.Fatal("expected duplicate password validation error")
	}

	resp, err := logic.ChangePassword(&types.ChangePasswordReq{CurrentPassword: "old-pass", NewPassword: "new-pass"})
	if err != nil {
		t.Fatalf("change password: %v", err)
	}
	if resp.Message != "密码已更新" || !store.updated {
		t.Fatalf("unexpected response or update state: %#v, updated=%v", resp, store.updated)
	}
	if store.user.PasswordHash != hashPassword("new-pass") {
		t.Fatal("expected password hash to be replaced")
	}
}
