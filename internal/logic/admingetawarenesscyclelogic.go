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

type AdminGetAwarenessCycleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminGetAwarenessCycleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetAwarenessCycleLogic {
	return &AdminGetAwarenessCycleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminGetAwarenessCycleLogic) AdminGetAwarenessCycle() (resp *types.AwarenessCycleAdminResp, err error) {
	if err = requireAdminUser(l.ctx); err != nil {
		return nil, err
	}

	var points []model.Awareness
	if l.svcCtx.AwarenessModel != nil {
		points, err = l.svcCtx.AwarenessModel.FindEligible(l.ctx)
		if err != nil {
			return nil, fmt.Errorf("query awareness: %w", err)
		}
	}

	startDate, restDays, err := getAwarenessCycleSettings(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	var pauses []model.AwarenessCyclePauses
	if cycle, cycleErr := getActiveAwarenessCycle(l.ctx, l.svcCtx); cycleErr == nil && l.svcCtx.AwarenessCyclePausesModel != nil {
		pauses, _ = l.svcCtx.AwarenessCyclePausesModel.FindActiveByCycle(l.ctx, cycle.CycleId)
	}

	weekStart := currentWeekStart(time.Now())
	data := buildAwarenessCycleAdminInfo(startDate, restDays, weekStart, points, pauses)
	return &types.AwarenessCycleAdminResp{
		Code:    0,
		Message: "ok",
		Data:    data,
	}, nil
}

func buildAwarenessCycleAdminInfo(startDate time.Time, restDays int, weekStart time.Time, points []model.Awareness, pauses []model.AwarenessCyclePauses) types.AwarenessCycleAdminInfo {
	weekDays := make([]types.AwarenessCycleDayInfo, 0, 7)
	var normalDayCount int64
	var restDayCount int64
	var pausedDayCount int64
	var currentProgressNo int64
	var currentProgressTitle string
	today := normalizeDate(time.Now())
	cycle := &model.AwarenessCycles{CycleId: 1, CommunityId: defaultCommunityID, StartDate: startDate, RestDays: int64(restDays)}
	plans := buildAwarenessSchedulePlan(cycle, points, pauses, weekStart, normalizeDate(weekStart).AddDate(0, 0, 6))
	for _, plan := range plans {
		date := normalizeDate(plan.Date)
		if plan.DayType == scheduleDayPaused {
			pausedDayCount++
			weekDays = append(weekDays, types.AwarenessCycleDayInfo{
				Date:        date.Format("2006-01-02"),
				Title:       "暂停打卡",
				Summary:     nullableString(plan.Pause.Reason),
				IsRestDay:   true,
				IsPausedDay: true,
			})
			continue
		}
		if plan.DayType != scheduleDayNormal || plan.Awareness == nil {
			restDayCount++
			weekDays = append(weekDays, types.AwarenessCycleDayInfo{
				Date:      date.Format("2006-01-02"),
				Title:     "本轮结束，休息整合中",
				Summary:   "今天不生成新的练习任务，可以回看历史打卡和到期复盘，把这一轮练过的意识点整合一下。",
				IsRestDay: true,
			})
			continue
		}

		normalDayCount++
		progressNo := plan.EffectiveDayIndex.Int64 + 1
		if normalizeDate(plan.Date).Equal(today) {
			currentProgressNo = progressNo
			currentProgressTitle = plan.Awareness.PointTitle
		}
		day := types.AwarenessCycleDayInfo{
			Date:        date.Format("2006-01-02"),
			Title:       plan.Awareness.PointTitle,
			Summary:     awarenessSummary(plan.Awareness),
			IsRestDay:   false,
			AwarenessId: plan.Awareness.AwarenessId,
			OrderNo:     plan.Awareness.SortOrderGlobal,
			ProgressNo:  progressNo,
		}
		weekDays = append(weekDays, day)
	}

	return types.AwarenessCycleAdminInfo{
		StartDate:              normalizeDate(startDate).Format("2006-01-02"),
		RestDays:               int64(restDays),
		PausedDates:            pausesToDateStrings(pauses),
		EligibleAwarenessCount: int64(len(points)),
		WeekStart:              normalizeDate(weekStart).Format("2006-01-02"),
		NormalDayCount:         normalDayCount,
		RestDayCount:           restDayCount,
		PausedDayCount:         pausedDayCount,
		CurrentProgressNo:      currentProgressNo,
		CurrentProgressTitle:   currentProgressTitle,
		WeekDays:               weekDays,
	}
}

func pausesToDateStrings(pauses []model.AwarenessCyclePauses) []string {
	result := make([]string, 0)
	for _, pause := range pauses {
		for date := normalizeDate(pause.PauseStartDate); !date.After(normalizeDate(pause.PauseEndDate)); date = date.AddDate(0, 0, 1) {
			result = append(result, date.Format("2006-01-02"))
		}
	}
	return result
}
