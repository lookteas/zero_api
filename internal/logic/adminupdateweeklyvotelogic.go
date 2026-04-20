// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUpdateWeeklyVoteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateWeeklyVoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateWeeklyVoteLogic {
	return &AdminUpdateWeeklyVoteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateWeeklyVoteLogic) AdminUpdateWeeklyVote(req *types.WeeklyVoteUpdateReq) (resp *types.SimpleResp, err error) {
	return okSimple("周投票已更新（演示数据）"), nil
}
