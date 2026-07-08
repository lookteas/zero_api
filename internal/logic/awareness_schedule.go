package logic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"api/internal/svc"
	"api/model"
)

const (
	defaultCommunityID = 1
	scheduleDayNormal  = "normal"
	scheduleDayPaused  = "paused"
	scheduleDayRest    = "rest"
)

type scheduleDayPlan struct {
	Date              time.Time
	DayType           string
	Awareness         *model.Awareness
	Pause             *model.AwarenessCyclePauses
	CycleIndex        int64
	CycleDayIndex     sql.NullInt64
	EffectiveDayIndex sql.NullInt64
}

type schedulePlanAnchor struct {
	CycleDayIndex int64
	Valid         bool
}

func generateAwarenessScheduleDays(cycle *model.AwarenessCycles, points []model.Awareness, pauses []model.AwarenessCyclePauses, fromDate, untilDate time.Time) []model.AwarenessScheduleDays {
	plans := buildAwarenessSchedulePlanWithAnchor(cycle, points, pauses, fromDate, untilDate, schedulePlanAnchor{})
	items := make([]model.AwarenessScheduleDays, 0, len(plans))
	for _, plan := range plans {
		items = append(items, schedulePlanToModel(cycle, plan))
	}
	return items
}

func buildAwarenessSchedulePlan(cycle *model.AwarenessCycles, points []model.Awareness, pauses []model.AwarenessCyclePauses, fromDate, untilDate time.Time) []scheduleDayPlan {
	return buildAwarenessSchedulePlanWithAnchor(cycle, points, pauses, fromDate, untilDate, schedulePlanAnchor{})
}

func buildAwarenessSchedulePlanWithAnchor(cycle *model.AwarenessCycles, points []model.Awareness, pauses []model.AwarenessCyclePauses, fromDate, untilDate time.Time, anchor schedulePlanAnchor) []scheduleDayPlan {
	if cycle == nil || untilDate.Before(fromDate) {
		return nil
	}

	start := normalizeDate(cycle.StartDate)
	fromDate = normalizeDate(fromDate)
	untilDate = normalizeDate(untilDate)
	restDays := int(cycle.RestDays)
	if restDays <= 0 {
		restDays = defaultAwarenessCycleRestDays
	}
	cycleLength := len(points) + restDays

	effectiveDaysBeforeFrom := int64(0)
	for date := start; date.Before(fromDate); date = date.AddDate(0, 0, 1) {
		if !date.Before(start) && findPauseForDate(pauses, date) == nil {
			effectiveDaysBeforeFrom++
		}
	}

	effectiveDayIndex := effectiveDaysBeforeFrom
	if anchor.Valid && len(points) > 0 && cycleLength > 0 && anchor.CycleDayIndex >= 0 && int(anchor.CycleDayIndex) < len(points) {
		effectiveDayIndex = anchoredEffectiveDayIndex(effectiveDaysBeforeFrom, cycleLength, anchor.CycleDayIndex)
	}
	plans := make([]scheduleDayPlan, 0, int(untilDate.Sub(fromDate).Hours()/24)+1)
	for date := fromDate; !date.After(untilDate); date = date.AddDate(0, 0, 1) {
		plan := scheduleDayPlan{Date: date, DayType: scheduleDayRest}
		if date.Before(start) {
			plans = append(plans, plan)
			continue
		}

		if pause := findPauseForDate(pauses, date); pause != nil {
			plan.DayType = scheduleDayPaused
			plan.Pause = pause
			plans = append(plans, plan)
			continue
		}

		if len(points) == 0 || cycleLength <= 0 {
			plans = append(plans, plan)
			effectiveDayIndex++
			continue
		}

		dayInCycle := int(effectiveDayIndex % int64(cycleLength))
		plan.CycleIndex = effectiveDayIndex / int64(cycleLength)
		if dayInCycle < len(points) {
			plan.DayType = scheduleDayNormal
			plan.Awareness = &points[dayInCycle]
			plan.CycleDayIndex = sql.NullInt64{Int64: int64(dayInCycle), Valid: true}
			plan.EffectiveDayIndex = sql.NullInt64{Int64: effectiveDayIndex, Valid: true}
		}
		plans = append(plans, plan)
		effectiveDayIndex++
	}

	return plans
}

func anchoredEffectiveDayIndex(base int64, cycleLength int, cycleDayIndex int64) int64 {
	if cycleLength <= 0 {
		return base
	}
	length := int64(cycleLength)
	remainder := base % length
	offset := (cycleDayIndex - remainder + length) % length
	return base + offset
}

func findPauseForDate(pauses []model.AwarenessCyclePauses, date time.Time) *model.AwarenessCyclePauses {
	date = normalizeDate(date)
	for i := range pauses {
		start := normalizeDate(pauses[i].PauseStartDate)
		end := normalizeDate(pauses[i].PauseEndDate)
		if !date.Before(start) && !date.After(end) {
			return &pauses[i]
		}
	}
	return nil
}

func schedulePlanToModel(cycle *model.AwarenessCycles, plan scheduleDayPlan) model.AwarenessScheduleDays {
	item := model.AwarenessScheduleDays{
		CycleId:           cycle.CycleId,
		CommunityId:       cycle.CommunityId,
		ScheduleDate:      normalizeDate(plan.Date),
		DayType:           plan.DayType,
		CycleIndex:        plan.CycleIndex,
		CycleDayIndex:     plan.CycleDayIndex,
		EffectiveDayIndex: plan.EffectiveDayIndex,
		GeneratedVersion:  1,
	}
	if item.CommunityId == 0 {
		item.CommunityId = defaultCommunityID
	}

	if plan.Awareness != nil {
		item.AwarenessId = sql.NullInt64{Int64: int64(plan.Awareness.AwarenessId), Valid: true}
	}
	if plan.Pause != nil {
		item.PauseId = sql.NullInt64{Int64: int64(plan.Pause.PauseId), Valid: plan.Pause.PauseId > 0}
		item.PauseReason = plan.Pause.Reason
	}

	return item
}

func getActiveAwarenessCycle(ctx context.Context, svcCtx *svc.ServiceContext) (*model.AwarenessCycles, error) {
	if svcCtx.AwarenessCyclesModel != nil {
		cycle, err := svcCtx.AwarenessCyclesModel.FindActiveByCommunity(ctx, defaultCommunityID)
		if err == nil {
			return cycle, nil
		}
		if err != model.ErrNotFound {
			return nil, err
		}
	}

	startDate, restDays, err := getAwarenessCycleSettings(ctx, svcCtx)
	if err != nil {
		return nil, err
	}
	return &model.AwarenessCycles{
		CycleId:             1,
		CommunityId:         defaultCommunityID,
		CycleName:           "默认意识打卡轮次",
		StartDate:           startDate,
		RestDays:            int64(restDays),
		ScheduleHorizonDays: 365,
		Status:              "active",
	}, nil
}

func generateAndStoreAwarenessSchedule(ctx context.Context, svcCtx *svc.ServiceContext, cycle *model.AwarenessCycles, fromDate time.Time) error {
	return generateAndStoreAwarenessScheduleWithAnchor(ctx, svcCtx, cycle, fromDate, schedulePlanAnchor{})
}

func generateAndStoreAwarenessScheduleWithAnchor(ctx context.Context, svcCtx *svc.ServiceContext, cycle *model.AwarenessCycles, fromDate time.Time, anchor schedulePlanAnchor) error {
	if svcCtx.AwarenessScheduleDaysModel == nil || svcCtx.AwarenessModel == nil {
		return nil
	}
	points, err := svcCtx.AwarenessModel.FindEligible(ctx)
	if err != nil {
		return err
	}
	var pauses []model.AwarenessCyclePauses
	if svcCtx.AwarenessCyclePausesModel != nil {
		pauses, err = svcCtx.AwarenessCyclePausesModel.FindActiveByCycle(ctx, cycle.CycleId)
		if err != nil {
			return err
		}
	}
	horizon := int(cycle.ScheduleHorizonDays)
	if horizon <= 0 {
		horizon = 365
	}
	fromDate = normalizeDate(fromDate)
	untilDate := fromDate.AddDate(0, 0, horizon-1)
	plans := buildAwarenessSchedulePlanWithAnchor(cycle, points, pauses, fromDate, untilDate, anchor)
	items := make([]model.AwarenessScheduleDays, 0, len(plans))
	for _, plan := range plans {
		items = append(items, schedulePlanToModel(cycle, plan))
	}
	for i := range items {
		if err = svcCtx.AwarenessScheduleDaysModel.Upsert(ctx, &items[i]); err != nil {
			return fmt.Errorf("upsert schedule day %s: %w", items[i].ScheduleDate.Format("2006-01-02"), err)
		}
	}
	cycle.LastGeneratedUntil = sql.NullTime{Time: untilDate, Valid: true}
	if svcCtx.AwarenessCyclesModel != nil {
		return svcCtx.AwarenessCyclesModel.Update(ctx, cycle)
	}
	return nil
}
