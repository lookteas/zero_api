// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDailyTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDailyTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDailyTaskLogic {
	return &GetDailyTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDailyTaskLogic) GetDailyTask(req *types.IdPathReq) (resp *types.DailyTaskResp, err error) {
	if l.svcCtx.DailyTasksModel == nil {
		return okDailyTask(), nil
	}

	item, err := l.svcCtx.DailyTasksModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return okDailyTask(), nil
	}

	return &types.DailyTaskResp{Code: 0, Message: "ok", Data: dailyTaskToInfo(item)}, nil
}
