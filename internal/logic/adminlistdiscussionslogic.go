// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListDiscussionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListDiscussionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListDiscussionsLogic {
	return &AdminListDiscussionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListDiscussionsLogic) AdminListDiscussions() (resp *types.SimpleResp, err error) {
	return okSimple("讨论说明列表接口待接数据库"), nil
}
