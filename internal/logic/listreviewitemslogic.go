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

type ListReviewItemsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListReviewItemsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListReviewItemsLogic {
	return &ListReviewItemsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListReviewItemsLogic) ListReviewItems(req *types.ReviewItemQueryReq) (resp *types.ReviewItemListResp, err error) {
	if l.svcCtx.ReviewItemsModel == nil {
		return okReviewItemList(), nil
	}

	query := "select id, user_id, daily_task_id, review_stage, due_at, status, completed_at, created_at, updated_at from review_items where user_id = ?"
	args := []any{currentUserID(l.ctx)}

	if req.Status != "" {
		query += " and status = ?"
		args = append(args, req.Status)
	}

	if req.Stage != "" {
		query += " and review_stage = ?"
		args = append(args, req.Stage)
	}

	query += " order by due_at asc limit 100"

	var items []model.ReviewItems
	if err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &items, query, args...); err != nil {
		return nil, fmt.Errorf("query review items: %w", err)
	}

	now := time.Now()
	lastActiveAt := resolveReviewLastActiveAt(l.ctx, l.svcCtx, currentUserID(l.ctx), now)
	tasks := loadReviewTaskInfoMap(l.ctx, l.svcCtx, items)
	records := loadReviewRecordInfoMap(l.ctx, l.svcCtx, items)
	data := BuildReviewItemListData(items, tasks, records, now, lastActiveAt)

	return &types.ReviewItemListResp{
		Code:    0,
		Message: "ok",
		Data:    data,
	}, nil
}
