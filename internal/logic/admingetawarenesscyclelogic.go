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

	weekStart := currentWeekStart(time.Now())
	data := buildAwarenessCycleAdminInfo(startDate, restDays, weekStart, points)
	return &types.AwarenessCycleAdminResp{
		Code:    0,
		Message: "ok",
		Data:    data,
	}, nil
}

func buildAwarenessCycleAdminInfo(startDate time.Time, restDays int, weekStart time.Time, points []model.Awareness) types.AwarenessCycleAdminInfo {
	weekDays := make([]types.AwarenessCycleDayInfo, 0, 7)
	var normalDayCount int64
	var restDayCount int64
	for i := 0; i < 7; i++ {
		date := normalizeDate(weekStart).AddDate(0, 0, i)
		cycle := resolveAwarenessCycleDay(startDate, date, restDays, points)
		if cycle.Awareness == nil || cycle.IsRestDay || cycle.IsPreStart {
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
		day := types.AwarenessCycleDayInfo{
			Date:        date.Format("2006-01-02"),
			Title:       cycle.Awareness.PointTitle,
			Summary:     awarenessSummary(cycle.Awareness),
			IsRestDay:   false,
			AwarenessId: cycle.Awareness.AwarenessId,
			OrderNo:     cycle.Awareness.SortOrderGlobal,
		}
		weekDays = append(weekDays, day)
	}

	return types.AwarenessCycleAdminInfo{
		StartDate:              normalizeDate(startDate).Format("2006-01-02"),
		RestDays:               int64(restDays),
		EligibleAwarenessCount: int64(len(points)),
		WeekStart:              normalizeDate(weekStart).Format("2006-01-02"),
		NormalDayCount:         normalDayCount,
		RestDayCount:           restDayCount,
		WeekDays:               weekDays,
	}
}
