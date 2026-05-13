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
	if l.svcCtx.DailyTasksModel == nil || l.svcCtx.AwarenessModel == nil {
		return okDailyTask(), nil
	}

	taskDate := parseTaskDate(req.TaskDate)
	userID := currentUserID(l.ctx)
	existing, findErr := l.svcCtx.DailyTasksModel.FindOneByUserIdTaskDate(l.ctx, userID, taskDate)
	if findErr != nil && findErr != model.ErrNotFound {
		return nil, findErr
	}

	if l.svcCtx.AwarenessScheduleDaysModel != nil {
		cycle, cycleErr := getActiveAwarenessCycle(l.ctx, l.svcCtx)
		if cycleErr != nil {
			return nil, cycleErr
		}
		scheduleDay, scheduleErr := l.svcCtx.AwarenessScheduleDaysModel.FindOneByCycleIdScheduleDate(l.ctx, cycle.CycleId, taskDate)
		if scheduleErr == nil {
			if scheduleDay.DayType == scheduleDayPaused || scheduleDay.DayType == scheduleDayRest {
				return &types.DailyTaskResp{Code: 0, Message: "ok", Data: scheduleDayToDailyTaskInfo(scheduleDay)}, nil
			}
			if findErr == model.ErrNotFound {
				return l.createDailyTaskFromScheduleDay(scheduleDay, userID, taskDate)
			}
		} else if scheduleErr != model.ErrNotFound {
			return nil, scheduleErr
		}
	}

	points, err := l.svcCtx.AwarenessModel.FindEligible(l.ctx)
	if err != nil {
		return nil, err
	}

	startDate, restDays, err := getAwarenessCycleSettings(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}

	cycle := resolveAwarenessCycleDay(startDate, taskDate, restDays, points)
	if findErr == nil {
		info := dailyTaskToInfo(existing)
		if existing.AwarenessId.Valid {
			info = applyAwarenessToDailyTaskInfo(info, findAwarenessByID(points, uint64(existing.AwarenessId.Int64)))
		} else if cycle.Awareness != nil {
			info = applyAwarenessToDailyTaskInfo(info, cycle.Awareness)
		}
		return &types.DailyTaskResp{Code: 0, Message: "ok", Data: info}, nil
	}

	if cycle.IsPreStart || cycle.IsRestDay || cycle.Awareness == nil {
		return &types.DailyTaskResp{Code: 0, Message: "ok", Data: restDailyTaskInfo(taskDate)}, nil
	}

	summary := awarenessSummary(cycle.Awareness)
	now := time.Now()
	data := &model.DailyTasks{
		UserId:           userID,
		CommunityId:      defaultCommunityID,
		TaskDate:         taskDate,
		TopicId:          0,
		AwarenessId:      sql.NullInt64{Int64: int64(cycle.Awareness.AwarenessId), Valid: true},
		TopicOrderNo:     cycle.Awareness.SortOrderGlobal,
		TopicTitle:       cycle.Awareness.PointTitle,
		TopicSummary:     summary,
		Weakness:         nullString(""),
		ImprovementPlan:  nullString(""),
		VerificationPath: nullString(""),
		ReflectionNote:   nullString(""),
		Status:           "draft",
		SubmittedAt:      sql.NullTime{},
		CreatedAt:        now,
		UpdatedAt:        now,
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

	info := applyAwarenessToDailyTaskInfo(dailyTaskToInfo(item), cycle.Awareness)
	return &types.DailyTaskResp{Code: 0, Message: "ok", Data: info}, nil
}

func (l *CreateDailyTaskLogic) createDailyTaskFromScheduleDay(scheduleDay *model.AwarenessScheduleDays, userID uint64, taskDate time.Time) (*types.DailyTaskResp, error) {
	now := time.Now()
	data := &model.DailyTasks{
		UserId:           userID,
		CommunityId:      scheduleDay.CommunityId,
		TaskDate:         taskDate,
		ScheduleDayId:    sql.NullInt64{Int64: int64(scheduleDay.ScheduleDayId), Valid: scheduleDay.ScheduleDayId > 0},
		TopicId:          0,
		AwarenessId:      sql.NullInt64{Int64: nullableInt64(scheduleDay.AwarenessId), Valid: scheduleDay.AwarenessId.Valid},
		TopicOrderNo:     nullableInt64(scheduleDay.CycleDayIndex),
		TopicTitle:       nullableString(scheduleDay.AwarenessTitle),
		TopicSummary:     nullableString(scheduleDay.AwarenessSummary),
		Weakness:         nullString(""),
		ImprovementPlan:  nullString(""),
		VerificationPath: nullString(""),
		ReflectionNote:   nullString(""),
		Status:           "draft",
		SubmittedAt:      sql.NullTime{},
		CreatedAt:        now,
		UpdatedAt:        now,
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
	info := dailyTaskToInfo(item)
	info = applyScheduleDayAwarenessToDailyTaskInfo(info, scheduleDay)
	return &types.DailyTaskResp{Code: 0, Message: "ok", Data: info}, nil
}
