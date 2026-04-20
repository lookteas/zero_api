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

type ListTopicsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTopicsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTopicsLogic {
	return &ListTopicsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTopicsLogic) ListTopics(req *types.TopicQueryReq) (resp *types.TopicListResp, err error) {
	if l.svcCtx.TopicsModel == nil {
		return okTopicList(), nil
	}

	query := "select id, title, summary, description, order_no, status, schedule_date, created_at, updated_at from topics where 1 = 1"
	args := make([]any, 0)

	if req.Status != 0 {
		query += " and status = ?"
		args = append(args, req.Status)
	}

	if req.Keyword != "" {
		query += " and (title like ? or summary like ?)"
		keyword := fmt.Sprintf("%%%s%%", req.Keyword)
		args = append(args, keyword, keyword)
	}

	query += " order by schedule_date is null asc, schedule_date asc, order_no asc limit 100"

	var items []model.Topics
	if err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &items, query, args...); err != nil {
		return nil, err
	}

	list := make([]types.TopicInfo, 0, len(items))
	for i := range items {
		list = append(list, topicToInfo(&items[i]))
	}

	return &types.TopicListResp{
		Code:    0,
		Message: "ok",
		Data: types.TopicListData{
			List: list,
			Pagination: types.Pagination{
				Page:     1,
				PageSize: int64(len(list)),
				Total:    int64(len(list)),
			},
		},
	}, nil
}
