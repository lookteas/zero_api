// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"database/sql"
		"errors"
	"regexp"
	"strings"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type PasswordLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasswordLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PasswordLoginLogic {
	return &PasswordLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PasswordLoginLogic) PasswordLogin(req *types.PasswordLoginReq) (resp *types.LoginResp, err error) {
	if l.svcCtx.UsersModel == nil {
		return okLogin(), nil
	}

	account := strings.TrimSpace(req.Account)
	password := strings.TrimSpace(req.Password)
	if len(password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	user, err := findLoginUser(l.ctx, l.svcCtx.UsersModel, account)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, err
		}
		return nil, err
	}

	if user.PasswordHash != hashPassword(password) {
		return nil, model.ErrNotFound
	}

	user.LastLoginAt = sql.NullTime{Time: time.Now(), Valid: true}
	if err = l.svcCtx.UsersModel.Update(l.ctx, user); err != nil {
		return nil, err
	}

	return loginRespFromUser(user), nil
}

func findLoginUser(ctx context.Context, usersModel model.UsersModel, account string) (*model.Users, error) {
	if isMobile(account) {
		return usersModel.FindOneByMobileString(ctx, account)
	}

	if isEmail(account) {
		return usersModel.FindOneByEmailString(ctx, account)
	}

	if len(account) <= 3 {
		return nil, errors.New("username length must be greater than 3")
	}

	return usersModel.FindOneByAccount(ctx, account)
}

func isEmail(value string) bool {
	return strings.Contains(value, "@")
}

func isMobile(value string) bool {
	matched, _ := regexp.MatchString(`^1\d{10}$`, value)
	return matched
}
