// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"fmt"
	"strings"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListTopicsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListTopicsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListTopicsLogic {
	return &AdminListTopicsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListTopicsLogic) AdminListTopics(req *types.TopicQueryReq) (resp *types.TopicListResp, err error) {
	if err = requireAdminUser(l.ctx); err != nil {
		return nil, err
	}
	if l.svcCtx.DB == nil {
		return okTopicList(), nil
	}

	query := "select id, title, summary, description, order_no, status, schedule_date, created_at, updated_at from topics where 1=1"
	args := make([]any, 0, 2)
	if strings.TrimSpace(req.Keyword) != "" {
		query += " and (title like ? or summary like ? or description like ?)"
		keyword := "%" + strings.TrimSpace(req.Keyword) + "%"
		args = append(args, keyword, keyword, keyword)
	}
	query += " order by schedule_date is null asc, schedule_date asc, order_no asc, id asc"

	var items []model.Topics
	if err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &items, query, args...); err != nil {
		return nil, fmt.Errorf("query topics: %w", err)
	}

	list := make([]types.TopicInfo, 0, len(items))
	for _, item := range items {
		copied := item
		list = append(list, topicToInfo(&copied))
	}

	return &types.TopicListResp{
		Code:    0,
		Message: "ok",
		Data: types.TopicListData{
			List:       list,
			Pagination: types.Pagination{Page: 1, PageSize: int64(len(list)), Total: int64(len(list))},
		},
	}, nil
}
