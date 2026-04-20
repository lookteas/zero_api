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
)

type CreateDailyTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateDailyTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDailyTaskLogic {
	return &CreateDailyTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDailyTaskLogic) CreateDailyTask(req *types.DailyTaskCreateReq) (resp *types.DailyTaskResp, err error) {
	if l.svcCtx.DailyTasksModel == nil || l.svcCtx.TopicsModel == nil {
		return okDailyTask(), nil
	}

	taskDate := parseTaskDate(req.TaskDate)
	userID := currentUserID(l.ctx)
	existing, findErr := l.svcCtx.DailyTasksModel.FindOneByUserIdTaskDate(l.ctx, userID, taskDate)
	if findErr == nil {
		return &types.DailyTaskResp{Code: 0, Message: "ok", Data: dailyTaskToInfo(existing)}, nil
	}

	if findErr != nil && findErr != model.ErrNotFound {
		return nil, findErr
	}

	topic, err := l.svcCtx.TopicsModel.FindLatestActiveByScheduleDate(l.ctx, taskDate)
	if err != nil {
		return nil, err
	}

	data := &model.DailyTasks{
		UserId:           userID,
		TaskDate:         taskDate,
		TopicId:          topic.Id,
		TopicOrderNo:     topic.OrderNo,
		TopicTitle:       topic.Title,
		TopicSummary:     topic.Summary,
		Weakness:         nullString(""),
		ImprovementPlan:  nullString(""),
		VerificationPath: nullString(""),
		ReflectionNote:   nullString(""),
		Status:           "draft",
		SubmittedAt:      sql.NullTime{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	result, err := l.svcCtx.DailyTasksModel.Insert(l.ctx, data)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	item, err := l.svcCtx.DailyTasksModel.FindOne(l.ctx, uint64(id))
	if err != nil {
		return nil, err
	}

	return &types.DailyTaskResp{Code: 0, Message: "ok", Data: dailyTaskToInfoWithTopic(item, topic)}, nil
}
