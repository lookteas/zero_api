package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUpdateAwarenessLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateAwarenessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateAwarenessLogic {
	return &AdminUpdateAwarenessLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateAwarenessLogic) AdminUpdateAwareness(req *types.AdminAwarenessUpdateReq) (*types.SimpleResp, error) {
	if err := requireAdminUser(l.ctx); err != nil {
		return nil, err
	}
	if l.svcCtx.AwarenessModel == nil {
		return nil, fmt.Errorf("awareness model unavailable")
	}
	title := strings.TrimSpace(req.Title)
	if req.Id == 0 || title == "" {
		return nil, fmt.Errorf("awareness id and title are required")
	}
	if err := l.svcCtx.AwarenessModel.UpdateContent(l.ctx, req.Id, title, strings.TrimSpace(req.Summary), strings.TrimSpace(req.Description)); err != nil {
		return nil, err
	}
	if err := regenerateAwarenessScheduleFrom(l.ctx, l.svcCtx, req.EffectiveDate); err != nil {
		return nil, err
	}
	return okSimple("ok"), nil
}

type AdminExcludeAwarenessLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminExcludeAwarenessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminExcludeAwarenessLogic {
	return &AdminExcludeAwarenessLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminExcludeAwarenessLogic) AdminExcludeAwareness(req *types.AdminAwarenessExcludeReq) (*types.SimpleResp, error) {
	if err := requireAdminUser(l.ctx); err != nil {
		return nil, err
	}
	if l.svcCtx.AwarenessModel == nil {
		return nil, fmt.Errorf("awareness model unavailable")
	}
	if req.Id == 0 {
		return nil, fmt.Errorf("awareness id is required")
	}
	if err := l.svcCtx.AwarenessModel.Disable(l.ctx, req.Id); err != nil {
		return nil, err
	}
	if err := regenerateAwarenessScheduleFrom(l.ctx, l.svcCtx, req.EffectiveDate); err != nil {
		return nil, err
	}
	return okSimple("ok"), nil
}

type AdminInsertAwarenessLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminInsertAwarenessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminInsertAwarenessLogic {
	return &AdminInsertAwarenessLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminInsertAwarenessLogic) AdminInsertAwareness(req *types.AdminAwarenessInsertReq) (*types.SimpleResp, error) {
	if err := requireAdminUser(l.ctx); err != nil {
		return nil, err
	}
	if l.svcCtx.AwarenessModel == nil {
		return nil, fmt.Errorf("awareness model unavailable")
	}
	effectiveDate, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(req.EffectiveDate), time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid effectiveDate, expected YYYY-MM-DD")
	}
	fromDate := normalizeDate(effectiveDate)
	position, err := l.insertionPositionForDate(fromDate)
	if err != nil {
		return nil, err
	}
	awarenessID, err := l.resolveInsertedAwarenessID(req)
	if err != nil {
		return nil, err
	}
	if err = l.svcCtx.AwarenessModel.MoveToPosition(l.ctx, awarenessID, position); err != nil {
		return nil, err
	}
	if err = regenerateAwarenessScheduleFromWithAnchor(l.ctx, l.svcCtx, fromDate.Format("2006-01-02"), schedulePlanAnchor{
		CycleDayIndex: position - 1,
		Valid:         true,
	}); err != nil {
		return nil, err
	}
	return okSimple("ok"), nil
}

func (l *AdminInsertAwarenessLogic) resolveInsertedAwarenessID(req *types.AdminAwarenessInsertReq) (uint64, error) {
	if req.ExistingAwarenessId > 0 {
		item, err := l.svcCtx.AwarenessModel.FindOne(l.ctx, req.ExistingAwarenessId)
		if err != nil {
			return 0, err
		}
		if item.Status != 1 {
			return 0, fmt.Errorf("disabled awareness cannot be inserted")
		}
		if item.IsMeta != 0 {
			return 0, fmt.Errorf("meta awareness cannot be inserted")
		}
		return item.AwarenessId, nil
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return 0, fmt.Errorf("title is required when creating awareness")
	}
	item, err := l.svcCtx.AwarenessModel.CreateMinimal(l.ctx, title, strings.TrimSpace(req.Summary), strings.TrimSpace(req.Description))
	if err != nil {
		return 0, err
	}
	return item.AwarenessId, nil
}

func (l *AdminInsertAwarenessLogic) insertionPositionForDate(date time.Time) (int64, error) {
	cycle, err := getActiveAwarenessCycle(l.ctx, l.svcCtx)
	if err != nil {
		return 0, err
	}
	points, err := l.svcCtx.AwarenessModel.FindEligible(l.ctx)
	if err != nil {
		return 0, err
	}
	var pauses []model.AwarenessCyclePauses
	if l.svcCtx.AwarenessCyclePausesModel != nil {
		pauses, err = l.svcCtx.AwarenessCyclePausesModel.FindActiveByCycle(l.ctx, cycle.CycleId)
		if err != nil {
			return 0, err
		}
	}
	plans := buildAwarenessSchedulePlan(cycle, points, pauses, date, date)
	if len(plans) == 0 || plans[0].DayType != scheduleDayNormal || !plans[0].CycleDayIndex.Valid {
		return 0, fmt.Errorf("selected date is not a practice day")
	}
	return plans[0].CycleDayIndex.Int64 + 1, nil
}

func regenerateAwarenessScheduleFrom(ctx context.Context, svcCtx *svc.ServiceContext, effectiveDate string) error {
	return regenerateAwarenessScheduleFromWithAnchor(ctx, svcCtx, effectiveDate, schedulePlanAnchor{})
}

func regenerateAwarenessScheduleFromWithAnchor(ctx context.Context, svcCtx *svc.ServiceContext, effectiveDate string, anchor schedulePlanAnchor) error {
	cycle, err := getActiveAwarenessCycle(ctx, svcCtx)
	if err != nil {
		return err
	}
	fromDate := normalizeDate(cycle.StartDate)
	if strings.TrimSpace(effectiveDate) != "" {
		parsed, parseErr := time.ParseInLocation("2006-01-02", strings.TrimSpace(effectiveDate), time.Local)
		if parseErr != nil {
			return fmt.Errorf("invalid effectiveDate, expected YYYY-MM-DD")
		}
		fromDate = normalizeDate(parsed)
	}
	return generateAndStoreAwarenessScheduleWithAnchor(ctx, svcCtx, cycle, fromDate, anchor)
}
