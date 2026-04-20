// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListReviewHistoryRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type reviewHistoryRow struct {
	Id              uint64         `db:"id"`
	ReviewItemId    uint64         `db:"review_item_id"`
	DailyTaskId     uint64         `db:"daily_task_id"`
	ReviewStage     string         `db:"review_stage"`
	TaskDate        time.Time      `db:"task_date"`
	TopicTitle      string         `db:"topic_title"`
	TopicSummary    string         `db:"topic_summary"`
	Result          string         `db:"result"`
	ActualSituation sql.NullString `db:"actual_situation"`
	Suggestion      sql.NullString `db:"suggestion"`
	SubmittedAt     time.Time      `db:"submitted_at"`
}

func NewListReviewHistoryRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListReviewHistoryRecordsLogic {
	return &ListReviewHistoryRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListReviewHistoryRecordsLogic) ListReviewHistoryRecords(req *types.ReviewHistoryQueryReq) (resp *types.ReviewHistoryListResp, err error) {
	if l.svcCtx.ReviewRecordsModel == nil {
		return okReviewHistoryList(), nil
	}

	query, args := buildReviewHistoryListQuery(currentUserID(l.ctx), req)

	var items []reviewHistoryRow
	if err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &items, query, args...); err != nil {
		return nil, fmt.Errorf("query review history: %w", err)
	}

	list := make([]types.ReviewHistoryRecordInfo, 0, len(items))
	for i := range items {
		list = append(list, reviewHistoryRowToInfo(&items[i]))
	}

	return &types.ReviewHistoryListResp{
		Code:    0,
		Message: "ok",
		Data: types.ReviewHistoryListData{
			List:       list,
			Pagination: types.Pagination{Page: 1, PageSize: int64(len(list)), Total: int64(len(list))},
		},
	}, nil
}

func buildReviewHistoryListQuery(userID uint64, req *types.ReviewHistoryQueryReq) (string, []any) {
	query := "select rr.id, rr.review_item_id, ri.daily_task_id, ri.review_stage, dt.task_date, dt.topic_title, dt.topic_summary, rr.result, rr.actual_situation, rr.suggestion, rr.submitted_at from review_records rr join review_items ri on rr.review_item_id = ri.id join daily_tasks dt on ri.daily_task_id = dt.id where rr.user_id = ?"
	args := []any{userID}

	if req.StartDate != "" {
		query += " and date(rr.submitted_at) >= ?"
		args = append(args, req.StartDate)
	}

	if req.EndDate != "" {
		query += " and date(rr.submitted_at) <= ?"
		args = append(args, req.EndDate)
	}

	if req.Keyword != "" {
		query += " and (dt.topic_title like ? or dt.topic_summary like ? or dt.weakness like ? or dt.improvement_plan like ? or dt.verification_path like ? or rr.actual_situation like ? or rr.suggestion like ?)"
		keyword := "%" + req.Keyword + "%"
		args = append(args, keyword, keyword, keyword, keyword, keyword, keyword, keyword)
	}

	query += " order by rr.submitted_at desc limit 100"
	return query, args
}

func reviewHistoryRowToInfo(item *reviewHistoryRow) types.ReviewHistoryRecordInfo {
	actualSituation := ""
	if item.ActualSituation.Valid {
		actualSituation = strings.TrimSpace(item.ActualSituation.String)
	}

	suggestion := ""
	if item.Suggestion.Valid {
		suggestion = strings.TrimSpace(item.Suggestion.String)
	}

	topicSummary := strings.TrimSpace(item.TopicSummary)

	return types.ReviewHistoryRecordInfo{
		Id:              item.Id,
		ReviewItemId:    item.ReviewItemId,
		DailyTaskId:     item.DailyTaskId,
		ReviewStage:     item.ReviewStage,
		TaskDate:        item.TaskDate.Format("2006-01-02"),
		TopicTitle:      item.TopicTitle,
		TopicSummary:    topicSummary,
		Summary:         buildReviewHistorySummary(actualSituation, suggestion, topicSummary),
		Result:          item.Result,
		ActualSituation: actualSituation,
		Suggestion:      suggestion,
		SubmittedAt:     item.SubmittedAt.Format("2006-01-02 15:04:05"),
	}
}

func buildReviewHistorySummary(actualSituation, suggestion, topicSummary string) string {
	for _, value := range []string{actualSituation, suggestion, topicSummary} {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}

	return "??????????????"
}

func okReviewHistoryList() *types.ReviewHistoryListResp {
	return &types.ReviewHistoryListResp{
		Code:    0,
		Message: "ok",
		Data: types.ReviewHistoryListData{
			List:       []types.ReviewHistoryRecordInfo{},
			Pagination: types.Pagination{Page: 1, PageSize: 0, Total: 0},
		},
	}
}
