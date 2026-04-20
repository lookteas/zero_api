// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"strings"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdateDailyTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateDailyTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDailyTaskLogic {
	return &UpdateDailyTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDailyTaskLogic) UpdateDailyTask(taskId uint64, req *types.DailyTaskUpdateReq) (resp *types.DailyTaskResp, err error) {
	if l.svcCtx.DailyTasksModel == nil {
		return okDailyTask(), nil
	}

	item, err := l.svcCtx.DailyTasksModel.FindOne(l.ctx, taskId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(codes.NotFound, "daily task not found")
		}
		return nil, err
	}

	if item.UserId != currentUserID(l.ctx) {
		return nil, status.Error(codes.NotFound, "daily task not found")
	}

	access := dailyTaskAccess(item, time.Now())
	if access.CanEditContent {
		item.Weakness = nullString(req.Weakness)
		item.ImprovementPlan = nullString(req.ImprovementPlan)
		item.VerificationPath = nullString(req.VerificationPath)
		item.ReflectionNote = nullString(req.ReflectionNote)
	} else {
		if strings.TrimSpace(req.Weakness) != "" || strings.TrimSpace(req.ImprovementPlan) != "" || strings.TrimSpace(req.VerificationPath) != "" {
			return nil, status.Error(codes.InvalidArgument, "daily task content can only be edited within 24 hours")
		}

		item.ReflectionNote = nullString(req.ReflectionNote)
	}

	if err = l.svcCtx.DailyTasksModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	updated, err := l.svcCtx.DailyTasksModel.FindOne(l.ctx, item.Id)
	if err != nil {
		return nil, err
	}

	return &types.DailyTaskResp{Code: 0, Message: "ok", Data: dailyTaskToInfo(updated)}, nil
}
