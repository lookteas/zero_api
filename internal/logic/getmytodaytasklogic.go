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
	var points []model.Awareness
	var cycle awarenessCycleResult
	hasAwarenessCycle := false
	if l.svcCtx.AwarenessModel != nil {
		eligible, awarenessErr := l.svcCtx.AwarenessModel.FindEligible(l.ctx)
		if awarenessErr == nil {
			points = eligible
			startDate, restDays, settingsErr := getAwarenessCycleSettings(l.ctx, l.svcCtx)
			if settingsErr != nil {
				return nil, settingsErr
			}
			cycle = resolveAwarenessCycleDay(startDate, today, restDays, points)
			hasAwarenessCycle = true
		}
	}

	item, err := l.svcCtx.DailyTasksModel.FindOneByUserIdTaskDate(l.ctx, currentUserID(l.ctx), today)
	if err != nil {
		if err == model.ErrNotFound {
			if hasAwarenessCycle && (cycle.IsPreStart || cycle.IsRestDay || cycle.Awareness == nil) {
				return &types.DailyTaskResp{Code: 0, Message: "ok", Data: restDailyTaskInfo(today)}, nil
			}
			return nil, status.Error(codes.NotFound, "today task not found")
		}

		return nil, err
	}

	info := dailyTaskToInfo(item)
	if hasAwarenessCycle {
		var matched *model.Awareness
		storedAwarenessMissing := false
		if item.AwarenessId.Valid {
			info.AwarenessId = uint64(item.AwarenessId.Int64)
			matched = findAwarenessByID(points, uint64(item.AwarenessId.Int64))
			storedAwarenessMissing = matched == nil
		}
		if !storedAwarenessMissing && shouldRefreshTodayTaskAwareness(item, cycle.Awareness) {
			refreshed := refreshTodayTaskAwarenessSnapshot(item, cycle.Awareness)
			if updateErr := l.svcCtx.DailyTasksModel.Update(l.ctx, refreshed); updateErr != nil {
				return nil, updateErr
			}
			item = refreshed
			info = dailyTaskToInfo(item)
			matched = cycle.Awareness
			if item.AwarenessId.Valid {
				info.AwarenessId = uint64(item.AwarenessId.Int64)
			}
		} else if !item.AwarenessId.Valid && cycle.Awareness != nil {
			matched = cycle.Awareness
		}
		info = applyAwarenessToDailyTaskInfo(info, matched)
	}

	return &types.DailyTaskResp{Code: 0, Message: "ok", Data: info}, nil
}
