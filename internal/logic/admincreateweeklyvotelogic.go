// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminCreateWeeklyVoteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminCreateWeeklyVoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreateWeeklyVoteLogic {
	return &AdminCreateWeeklyVoteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminCreateWeeklyVoteLogic) AdminCreateWeeklyVote(req *types.WeeklyVoteCreateReq) (resp *types.SimpleResp, err error) {
	return okSimple("周投票已创建（演示数据）"), nil
}
