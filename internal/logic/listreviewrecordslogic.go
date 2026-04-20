// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"fmt"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListReviewRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListReviewRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListReviewRecordsLogic {
	return &ListReviewRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListReviewRecordsLogic) ListReviewRecords(reviewItemId uint64, req *types.DailyLogQueryReq) (resp *types.ReviewRecordListResp, err error) {
	if l.svcCtx.ReviewRecordsModel == nil {
		return okReviewRecordList(), nil
	}

	query := "select id, review_item_id, user_id, result, actual_situation, suggestion, submitted_at, created_at, updated_at from review_records where review_item_id = ? order by submitted_at desc limit 100"
	var items []model.ReviewRecords
	if err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &items, query, reviewItemId); err != nil {
		return nil, fmt.Errorf("query review records: %w", err)
	}

	list := make([]types.ReviewRecordInfo, 0, len(items))
	for i := range items {
		list = append(list, reviewRecordToInfo(&items[i]))
	}

	return &types.ReviewRecordListResp{
		Code:    0,
		Message: "ok",
		Data: types.ReviewRecordListData{
			List: list,
			Pagination: types.Pagination{Page: 1, PageSize: int64(len(list)), Total: int64(len(list))},
		},
	}, nil
}
