package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"api/internal/svc"
	"api/internal/types"

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

func regenerateAwarenessScheduleFrom(ctx context.Context, svcCtx *svc.ServiceContext, effectiveDate string) error {
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
	return generateAndStoreAwarenessSchedule(ctx, svcCtx, cycle, fromDate)
}
