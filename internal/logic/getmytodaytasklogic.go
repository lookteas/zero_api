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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetMyTodayTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyTodayTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyTodayTaskLogic {
	return &GetMyTodayTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyTodayTaskLogic) GetMyTodayTask() (resp *types.DailyTaskResp, err error) {
	if l.svcCtx.DailyTasksModel == nil {
		return okDailyTask(), nil
	}

	today := normalizeDate(time.Now())
	var matchedTopic *model.Topics
	item, err := l.svcCtx.DailyTasksModel.FindOneByUserIdTaskDate(l.ctx, currentUserID(l.ctx), today)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(codes.NotFound, "today task not found")
		}

		return nil, err
	}

	if l.svcCtx.TopicsModel != nil {
		topic, topicErr := l.svcCtx.TopicsModel.FindLatestActiveByScheduleDate(l.ctx, item.TaskDate)
		if topicErr == nil && shouldRefreshTodayTaskTopic(item, topic) {
			refreshed := refreshTodayTaskTopicSnapshot(item, topic)
			if updateErr := l.svcCtx.DailyTasksModel.Update(l.ctx, refreshed); updateErr == nil {
				item = refreshed
			}
		}
		if topicErr == nil {
			matchedTopic = topic
		}
	}

	return &types.DailyTaskResp{Code: 0, Message: "ok", Data: dailyTaskToInfoWithTopic(item, matchedTopic)}, nil
}
