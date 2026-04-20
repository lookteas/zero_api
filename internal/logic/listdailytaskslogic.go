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

type ListDailyTasksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListDailyTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDailyTasksLogic {
	return &ListDailyTasksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListDailyTasksLogic) ListDailyTasks(req *types.DailyTaskQueryReq) (resp *types.DailyTaskListResp, err error) {
	if l.svcCtx.DailyTasksModel == nil {
		return okDailyTaskList(), nil
	}

	query, args := buildDailyTaskListQuery(currentUserID(l.ctx), req)

	var items []model.DailyTasks
	if err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &items, query, args...); err != nil {
		return nil, fmt.Errorf("query daily tasks: %w", err)
	}

	list := make([]types.DailyTaskInfo, 0, len(items))
	for i := range items {
		list = append(list, dailyTaskToInfo(&items[i]))
	}

	return &types.DailyTaskListResp{
		Code:    0,
		Message: "ok",
		Data: types.DailyTaskListData{
			List: list,
			Pagination: types.Pagination{Page: 1, PageSize: int64(len(list)), Total: int64(len(list))},
		},
	}, nil
}

func buildDailyTaskListQuery(userID uint64, req *types.DailyTaskQueryReq) (string, []any) {
	query := "select id, user_id, task_date, topic_id, topic_order_no, topic_title, topic_summary, weakness, improvement_plan, verification_path, reflection_note, status, submitted_at, created_at, updated_at from daily_tasks where user_id = ?"
	args := []any{userID}

	if req.Status != "" {
		query += " and status = ?"
		args = append(args, req.Status)
	}

	if req.StartDate != "" {
		query += " and task_date >= ?"
		args = append(args, req.StartDate)
	}

	if req.EndDate != "" {
		query += " and task_date <= ?"
		args = append(args, req.EndDate)
	}

	if req.Keyword != "" {
		query += " and (topic_title like ? or weakness like ? or improvement_plan like ? or verification_path like ? or reflection_note like ?)"
		keyword := "%" + req.Keyword + "%"
		args = append(args, keyword, keyword, keyword, keyword, keyword)
	}

	query += " order by task_date desc limit 100"
	return query, args
}
