// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMeLogic {
	return &GetMeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMeLogic) GetMe() (resp *types.UserResp, err error) {
	if l.svcCtx.UsersModel == nil {
		return okUser(), nil
	}

	user, err := l.svcCtx.UsersModel.FindOne(l.ctx, currentUserID(l.ctx))
	if err != nil {
		return okUser(), nil
	}

	return &types.UserResp{Code: 0, Message: "ok", Data: userToInfo(user)}, nil
}
