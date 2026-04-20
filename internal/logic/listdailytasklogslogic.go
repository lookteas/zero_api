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

type ListDailyTaskLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListDailyTaskLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDailyTaskLogsLogic {
	return &ListDailyTaskLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListDailyTaskLogsLogic) ListDailyTaskLogs(taskId uint64, req *types.DailyLogQueryReq) (resp *types.DailyLogListResp, err error) {
	if l.svcCtx.DailyLogsModel == nil {
		return okDailyLogList(), nil
	}

	query := "select id, user_id, daily_task_id, log_time, action_text, status, remark, created_at, updated_at from daily_logs where daily_task_id = ? order by log_time desc limit 100"
	var items []model.DailyLogs
	if err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &items, query, taskId); err != nil {
		return nil, fmt.Errorf("query daily logs: %w", err)
	}

	list := make([]types.DailyLogInfo, 0, len(items))
	for i := range items {
		list = append(list, dailyLogToInfo(&items[i]))
	}

	return &types.DailyLogListResp{
		Code:    0,
		Message: "ok",
		Data: types.DailyLogListData{
			List: list,
			Pagination: types.Pagination{Page: 1, PageSize: int64(len(list)), Total: int64(len(list))},
		},
	}, nil
}
