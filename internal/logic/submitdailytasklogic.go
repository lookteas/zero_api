// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"database/sql"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SubmitDailyTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitDailyTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitDailyTaskLogic {
	return &SubmitDailyTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitDailyTaskLogic) SubmitDailyTask(req *types.IdPathReq) (resp *types.DailyTaskResp, err error) {
	if l.svcCtx.DailyTasksModel == nil || l.svcCtx.ReviewItemsModel == nil {
		return okDailyTask(), nil
	}

	item, err := l.svcCtx.DailyTasksModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(codes.NotFound, "daily task not found")
		}
		return nil, err
	}

	if item.UserId != currentUserID(l.ctx) {
		return nil, status.Error(codes.NotFound, "daily task not found")
	}

	if !item.Weakness.Valid || item.Weakness.String == "" || !item.ImprovementPlan.Valid || item.ImprovementPlan.String == "" || !item.VerificationPath.Valid || item.VerificationPath.String == "" {
		return nil, status.Error(codes.InvalidArgument, "weakness, improvementPlan and verificationPath are required")
	}

	now := time.Now()
	item.Status = "submitted"
	item.SubmittedAt = sql.NullTime{Time: now, Valid: true}

	if err = l.svcCtx.DailyTasksModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	for _, stage := range buildReviewStagePlans(now) {
		_, findErr := l.svcCtx.ReviewItemsModel.FindOneByDailyTaskIdReviewStage(l.ctx, item.Id, stage.Name)
		if findErr == nil {
			continue
		}

		if findErr != model.ErrNotFound {
			return nil, findErr
		}

		_, err = l.svcCtx.ReviewItemsModel.Insert(l.ctx, &model.ReviewItems{
			UserId:      item.UserId,
			DailyTaskId: item.Id,
			ReviewStage: stage.Name,
			DueAt:       stage.DueAt,
			Status:      "pending",
			CompletedAt: sql.NullTime{},
		})
		if err != nil {
			return nil, err
		}
	}

	updated, err := l.svcCtx.DailyTasksModel.FindOne(l.ctx, item.Id)
	if err != nil {
		return nil, err
	}

	return &types.DailyTaskResp{Code: 0, Message: "ok", Data: dailyTaskToInfo(updated)}, nil
}
