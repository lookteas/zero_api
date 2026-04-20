// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListWeeklyVotesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListWeeklyVotesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListWeeklyVotesLogic {
	return &AdminListWeeklyVotesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListWeeklyVotesLogic) AdminListWeeklyVotes() (resp *types.SimpleResp, err error) {
	return okSimple("周投票列表接口待接数据库"), nil
}
