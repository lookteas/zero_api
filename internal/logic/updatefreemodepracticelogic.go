package logic

import (
	"context"
	"fmt"
	"strings"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFreemodePracticeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateFreemodePracticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFreemodePracticeLogic {
	return &UpdateFreemodePracticeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFreemodePracticeLogic) UpdateFreemodePractice(practiceID uint64, req *types.FreeModePracticeUpdateReq) (*types.FreeModePracticeResp, error) {
	if l.svcCtx.FreeModePracticesModel == nil {
		return nil, fmt.Errorf("free mode practice storage not ready")
	}

	item, err := l.svcCtx.FreeModePracticesModel.FindOne(l.ctx, practiceID)
	if err != nil {
		return nil, fmt.Errorf("query free mode practice: %w", err)
	}
	if item.UserId != currentUserID(l.ctx) {
		return nil, fmt.Errorf("free mode practice not found")
	}

	item.PracticeNote = sqlNullString(strings.TrimSpace(req.PracticeNote))
	if err := l.svcCtx.FreeModePracticesModel.Update(l.ctx, item); err != nil {
		return nil, fmt.Errorf("update free mode practice: %w", err)
	}

	updated, err := l.svcCtx.FreeModePracticesModel.FindOne(l.ctx, practiceID)
	if err != nil {
		return nil, fmt.Errorf("load free mode practice: %w", err)
	}
	if updated.UserId != currentUserID(l.ctx) {
		return nil, fmt.Errorf("free mode practice not found")
	}

	return &types.FreeModePracticeResp{
		Code:    0,
		Message: "ok",
		Data:    freeModePracticeToInfo(updated),
	}, nil
}
