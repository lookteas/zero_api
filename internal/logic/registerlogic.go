// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"
	"strings"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.LoginResp, err error) {
	if l.svcCtx.UsersModel == nil {
		return okLogin(), nil
	}

	account := strings.TrimSpace(req.Account)
	password := strings.TrimSpace(req.Password)
	nickname := strings.TrimSpace(req.Nickname)
	email := strings.TrimSpace(req.Email)
	mobile := strings.TrimSpace(req.Mobile)

	if len(password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	if isEmail(account) {
		email = account
	} else if isMobile(account) {
		mobile = account
	} else if len(account) <= 3 {
		return nil, errors.New("username length must be greater than 3")
	}

	_, err = l.svcCtx.UsersModel.FindOneByAccount(l.ctx, account)
	if err == nil {
		return nil, errors.New("account already exists")
	}
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}

	result, err := l.svcCtx.UsersModel.Insert(l.ctx, buildUser(account, password, nicknameOrAccount(nickname, account), email, mobile))
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	user, err := l.svcCtx.UsersModel.FindOne(l.ctx, uint64(id))
	if err != nil {
		return nil, err
	}

	return loginRespFromUser(user), nil
}

func nicknameOrAccount(nickname, account string) string {
	if nickname != "" {
		return nickname
	}
	return account
}