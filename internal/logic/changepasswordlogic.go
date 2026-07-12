package logic

import (
	"context"
	"errors"
	"strings"

	"api/internal/svc"
	"api/internal/types"
)

type ChangePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordReq) (*types.SimpleResp, error) {
	if l.svcCtx.UsersModel == nil {
		return okSimple("密码已更新"), nil
	}

	currentPassword := strings.TrimSpace(req.CurrentPassword)
	newPassword := strings.TrimSpace(req.NewPassword)
	if len(currentPassword) < 6 || len(newPassword) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}
	if currentPassword == newPassword {
		return nil, errors.New("new password must differ from current password")
	}

	user, err := l.svcCtx.UsersModel.FindOne(l.ctx, currentUserID(l.ctx))
	if err != nil {
		return nil, err
	}
	if user.PasswordHash != hashPassword(currentPassword) {
		return nil, errors.New("current password is incorrect")
	}

	user.PasswordHash = hashPassword(newPassword)
	if err := l.svcCtx.UsersModel.Update(l.ctx, user); err != nil {
		return nil, err
	}

	return okSimple("密码已更新"), nil
}
