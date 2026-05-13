// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUpdateAwarenessCycleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateAwarenessCycleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateAwarenessCycleLogic {
	return &AdminUpdateAwarenessCycleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateAwarenessCycleLogic) AdminUpdateAwarenessCycle(req *types.AwarenessCycleUpdateReq) (resp *types.SimpleResp, err error) {
	if err = requireAdminUser(l.ctx); err != nil {
		return nil, err
	}
	if err = updateAwarenessCycleSettings(l.ctx, l.svcCtx, req.StartDate, int(req.RestDays)); err != nil {
		return nil, err
	}
	cycle, err := getActiveAwarenessCycle(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if err = l.replaceCyclePauses(cycle, req.PausedDates); err != nil {
		return nil, err
	}
	if err = generateAndStoreAwarenessSchedule(l.ctx, l.svcCtx, cycle, cycle.StartDate); err != nil {
		return nil, err
	}

	return okSimple("ok"), nil
}

func (l *AdminUpdateAwarenessCycleLogic) replaceCyclePauses(cycle *model.AwarenessCycles, pausedDates []string) error {
	if l.svcCtx.AwarenessCyclePausesModel == nil {
		return updateAwarenessCyclePausedDatesSetting(l.ctx, l.svcCtx, pausedDates)
	}
	if err := l.svcCtx.AwarenessCyclePausesModel.DeleteByCycle(l.ctx, cycle.CycleId); err != nil {
		return err
	}
	normalized, err := normalizePausedDates(pausedDates)
	if err != nil {
		return err
	}
	for _, date := range normalized {
		if _, err = l.svcCtx.AwarenessCyclePausesModel.Insert(l.ctx, &model.AwarenessCyclePauses{
			CycleId:        cycle.CycleId,
			CommunityId:    cycle.CommunityId,
			PauseStartDate: date,
			PauseEndDate:   date,
			Reason:         sql.NullString{String: "后台暂停", Valid: true},
			Status:         1,
		}); err != nil {
			return err
		}
	}
	return updateAwarenessCyclePausedDatesSetting(l.ctx, l.svcCtx, dateListToStrings(normalized))
}

func normalizePausedDates(inputs []string) ([]time.Time, error) {
	seen := make(map[string]bool)
	result := make([]time.Time, 0, len(inputs))
	for _, input := range inputs {
		value := strings.TrimSpace(input)
		if value == "" {
			continue
		}
		parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
		if err != nil {
			return nil, err
		}
		key := normalizeDate(parsed).Format("2006-01-02")
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, normalizeDate(parsed))
	}
	return result, nil
}

func dateListToStrings(dates []time.Time) []string {
	result := make([]string, 0, len(dates))
	for _, date := range dates {
		result = append(result, normalizeDate(date).Format("2006-01-02"))
	}
	return result
}
