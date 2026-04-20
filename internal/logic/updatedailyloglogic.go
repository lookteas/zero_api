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

type UpdateDailyLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateDailyLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDailyLogLogic {
	return &UpdateDailyLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDailyLogLogic) UpdateDailyLog(logId uint64, req *types.DailyLogUpdateReq) (resp *types.DailyLogResp, err error) {
	if l.svcCtx.DailyLogsModel == nil {
		return okDailyLog(), nil
	}

	item, err := l.svcCtx.DailyLogsModel.FindOne(l.ctx, logId)
	if err != nil {
		return okDailyLog(), nil
	}

	if req.LogTime != "" {
		if parsed, parseErr := time.ParseInLocation("2006-01-02 15:04:05", req.LogTime, time.Local); parseErr == nil {
			item.LogTime = parsed
		}
	}

	if req.ActionText != "" {
		item.ActionText = nullString(req.ActionText)
	}

	if req.Status != "" {
		item.Status = req.Status
	}

	if req.Remark != "" {
		item.Remark = nullString(req.Remark)
	}

	if err = l.svcCtx.DailyLogsModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	updated, err := l.svcCtx.DailyLogsModel.FindOne(l.ctx, item.Id)
	if err != nil {
		return nil, err
	}

	return &types.DailyLogResp{Code: 0, Message: "ok", Data: dailyLogToInfo(updated)}, nil
}
