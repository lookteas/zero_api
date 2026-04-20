// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDailyTaskLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateDailyTaskLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDailyTaskLogLogic {
	return &CreateDailyTaskLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDailyTaskLogLogic) CreateDailyTaskLog(taskId uint64, req *types.DailyLogCreateReq) (resp *types.DailyLogResp, err error) {
	if l.svcCtx.DailyLogsModel == nil {
		return okDailyLog(), nil
	}

	logTime := time.Now()
	if req.LogTime != "" {
		parsed, parseErr := time.ParseInLocation("2006-01-02 15:04:05", req.LogTime, time.Local)
		if parseErr == nil {
			logTime = parsed
		}
	}

	userID := currentUserID(l.ctx)

	data := &model.DailyLogs{
		UserId:      userID,
		DailyTaskId: taskId,
		LogTime:     logTime,
		ActionText:  nullString(req.ActionText),
		Status:      req.Status,
		Remark:      nullString(req.Remark),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := l.svcCtx.DailyLogsModel.Insert(l.ctx, data)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	item, err := l.svcCtx.DailyLogsModel.FindOne(l.ctx, uint64(id))
	if err != nil {
		return nil, err
	}

	return &types.DailyLogResp{Code: 0, Message: "ok", Data: dailyLogToInfo(item)}, nil
}
