// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"time"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCurrentWeeklyVoteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCurrentWeeklyVoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrentWeeklyVoteLogic {
	return &GetCurrentWeeklyVoteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCurrentWeeklyVoteLogic) GetCurrentWeeklyVote() (resp *types.WeeklyVoteResp, err error) {
	if l.svcCtx.DB == nil {
		return okWeeklyVote(), nil
	}

	bundle, err := loadOrCreateCurrentWeeklyVote(l.ctx, l.svcCtx, time.Now(), currentUserID(l.ctx))
	if err != nil {
		return nil, err
	}

	return &types.WeeklyVoteResp{Code: 0, Message: "ok", Data: weeklyVoteToInfo(bundle)}, nil
}
