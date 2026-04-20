// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminStatsOverviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminStatsOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminStatsOverviewLogic {
	return &AdminStatsOverviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminStatsOverviewLogic) AdminStatsOverview() (resp *types.SimpleResp, err error) {
	return okSimple("统计概览接口待接数据库"), nil
}
