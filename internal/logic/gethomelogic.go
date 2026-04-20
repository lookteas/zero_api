// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"fmt"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHomeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetHomeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHomeLogic {
	return &GetHomeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHomeLogic) GetHome() (resp *types.HomeResp, err error) {
	if l.svcCtx.DB == nil {
		return okHome(), nil
	}

	userID := currentUserID(l.ctx)
	now := time.Now()

	home := types.HomeData{
		Overview: types.HomeOverview{},
	}

	weeklyBundle, weeklyErr := loadOrCreateCurrentWeeklyVote(l.ctx, l.svcCtx, now, userID)
	if weeklyErr == nil && weeklyBundle != nil {
		home.CurrentVote = weeklyVoteToInfo(weeklyBundle)

		weekStart := currentWeekStart(now)
		discussionItem, discussionErr := l.svcCtx.DiscussionInfosModel.FindOneByWeekStartDate(l.ctx, weekStart)
		if discussionErr == nil && discussionItem != nil {
			home.CurrentDiscussion = discussionToInfo(discussionItem)
		} else {
			home.CurrentDiscussion = buildDerivedDiscussion(weeklyBundle)
		}
	}

	todayTask, taskErr := l.svcCtx.DailyTasksModel.FindOneByUserIdTaskDate(l.ctx, userID, normalizeDate(now))
	if taskErr == nil && todayTask != nil {
		mapped := dailyTaskToInfo(todayTask)
		home.TodayTask = &mapped
	}

	var pendingItems []model.ReviewItems
	query := "select id, user_id, daily_task_id, review_stage, due_at, status, completed_at, created_at, updated_at from review_items where user_id = ? and status = ? order by due_at asc limit 100"
	if err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &pendingItems, query, userID, "pending"); err != nil {
		return nil, fmt.Errorf("query home review items: %w", err)
	}

	lastActiveAt := resolveReviewLastActiveAt(l.ctx, l.svcCtx, userID, now)
	tasks := loadReviewTaskInfoMap(l.ctx, l.svcCtx, pendingItems)
	home.PendingReviews, home.RecoveryReviews = BuildHomeReviewPresentation(pendingItems, tasks, now, lastActiveAt)

	var totalTaskCount int64
	if err = l.svcCtx.DB.QueryRowCtx(l.ctx, &totalTaskCount, "select count(1) from daily_tasks where user_id = ? and status = 'submitted'", userID); err != nil {
		return nil, err
	}

	var totalReviewCount int64
	if err = l.svcCtx.DB.QueryRowCtx(l.ctx, &totalReviewCount, "select count(1) from review_records where user_id = ?", userID); err != nil {
		return nil, err
	}

	pendingCount := len(home.PendingReviews)
	if len(home.RecoveryReviews) > 0 {
		pendingCount = len(home.RecoveryReviews)
	}

	home.ReinforcementHints = BuildReinforcementHints(now, totalTaskCount, loadReinforcementSignals(l.ctx, l.svcCtx, userID, now))

	home.CycleSummary = BuildCycleSummary(configuredCycleTotalPoints(l.svcCtx.Config.Cycle.TotalPoints), totalTaskCount, resolveLatestCompletedTaskAt(home.TodayTask, now, l.ctx, l.svcCtx, userID))

	home.Overview = types.HomeOverview{
		ContinuousDays:     totalTaskCount,
		TotalTaskCount:     totalTaskCount,
		TotalReviewCount:   totalReviewCount,
		PendingReviewCount: int64(pendingCount),
	}

	return &types.HomeResp{Code: 0, Message: "ok", Data: home}, nil
}

func configuredCycleTotalPoints(value int64) int64 {
	if value > 0 {
		return value
	}
	return defaultCycleTotalPoints
}

func resolveLatestCompletedTaskAt(todayTask *types.DailyTaskInfo, fallback time.Time, ctx context.Context, svcCtx *svc.ServiceContext, userID uint64) time.Time {
	if todayTask != nil && todayTask.SubmittedAt != "" {
		if parsed, err := time.ParseInLocation("2006-01-02 15:04:05", todayTask.SubmittedAt, time.Local); err == nil {
			return parsed
		}
	}

	var latestSubmittedAt time.Time
	if err := svcCtx.DB.QueryRowCtx(ctx, &latestSubmittedAt, "select submitted_at from daily_tasks where user_id = ? and status = 'submitted' order by submitted_at desc limit 1", userID); err == nil && !latestSubmittedAt.IsZero() {
		return latestSubmittedAt
	}
	return fallback
}
