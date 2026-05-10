// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUpdateAwarenessCycleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateAwarenessCycleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateAwarenessCycleLogic {
	return &AdminUpdateAwarenessCycleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateAwarenessCycleLogic) AdminUpdateAwarenessCycle(req *types.AwarenessCycleUpdateReq) (resp *types.SimpleResp, err error) {
	if err = requireAdminUser(l.ctx); err != nil {
		return nil, err
	}
	if err = updateAwarenessCycleSettings(l.ctx, l.svcCtx, req.StartDate, int(req.RestDays)); err != nil {
		return nil, err
	}

	return okSimple("ok"), nil
}
