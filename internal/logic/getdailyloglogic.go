// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDailyLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDailyLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDailyLogLogic {
	return &GetDailyLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDailyLogLogic) GetDailyLog(req *types.IdPathReq) (resp *types.DailyLogResp, err error) {
	if l.svcCtx.DailyLogsModel == nil {
		return okDailyLog(), nil
	}

	item, err := l.svcCtx.DailyLogsModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return okDailyLog(), nil
	}

	return &types.DailyLogResp{Code: 0, Message: "ok", Data: dailyLogToInfo(item)}, nil
}
