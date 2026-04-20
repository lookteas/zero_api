// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CodeLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCodeLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CodeLoginLogic {
	return &CodeLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CodeLoginLogic) CodeLogin(req *types.CodeLoginReq) (resp *types.LoginResp, err error) {
	return okLogin(), nil
}
