// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLoginLogic {
	return &AdminLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminLoginLogic) AdminLogin(req *types.AdminLoginReq) (resp *types.LoginResp, err error) {
	if l.svcCtx.AdminUsersModel == nil {
		return okLogin(), nil
	}

	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if username == "" || len(password) < 6 {
		return nil, model.ErrNotFound
	}

	adminUser, err := l.svcCtx.AdminUsersModel.FindOneByUsername(l.ctx, username)
	if err != nil {
		return nil, err
	}
	if adminUser.Status != 1 {
		return nil, model.ErrNotFound
	}
	if adminUser.PasswordHash != hashPassword(password) {
		return nil, model.ErrNotFound
	}

	adminUser.LastLoginAt = sql.NullTime{Time: time.Now(), Valid: true}
	if err = l.svcCtx.AdminUsersModel.Update(l.ctx, adminUser); err != nil {
		return nil, err
	}
	if strings.TrimSpace(adminUser.Nickname) == "" {
		adminUser.Nickname = adminUser.Username
	}

	return &types.LoginResp{
		Code:    0,
		Message: "ok",
		Data: types.LoginData{
			AccessToken:  "dev-token-admin-1",
			RefreshToken: "dev-refresh-token-admin-1",
			AccessExpire: 86400,
			User: types.UserInfo{
				Id:       adminUser.Id,
				Account:  adminUser.Username,
				Nickname: adminUser.Nickname,
				Avatar:   "",
				Status:   int64(adminUser.Status),
			},
		},
	}, nil
}
