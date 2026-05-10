// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

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
	if strings.TrimSpace(req.WeekStart) != "" {
		return l.adminListAwarenessSchedule(req)
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

func (l *AdminListTopicsLogic) adminListAwarenessSchedule(req *types.TopicQueryReq) (*types.TopicListResp, error) {
	weekStart, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(req.WeekStart), time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid weekStart: %w", err)
	}

	var points []model.Awareness
	if l.svcCtx.AwarenessModel != nil {
		points, err = l.svcCtx.AwarenessModel.FindEligible(l.ctx)
		if err != nil {
			return nil, fmt.Errorf("query awareness: %w", err)
		}
	}

	startDate, restDays, err := getAwarenessCycleSettings(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	list := make([]types.TopicInfo, 0, 7)
	for i := 0; i < 7; i++ {
		date := normalizeDate(weekStart).AddDate(0, 0, i)
		cycle := resolveAwarenessCycleDay(startDate, date, restDays, points)
		if cycle.Awareness == nil || cycle.IsRestDay || cycle.IsPreStart {
			list = append(list, restTopicInfo(date))
			continue
		}

		list = append(list, awarenessTopicInfo(cycle.Awareness, date))
	}

	return &types.TopicListResp{
		Code:    0,
		Message: "ok",
		Data: types.TopicListData{
			List:       list,
			Pagination: types.Pagination{Page: 1, PageSize: 7, Total: 7},
		},
	}, nil
}

func awarenessTopicInfo(item *model.Awareness, date time.Time) types.TopicInfo {
	info := types.TopicInfo{
		Id:           item.AwarenessId,
		Title:        item.PointTitle,
		Summary:      awarenessSummary(item),
		Description:  awarenessDetails(item),
		OrderNo:      item.SortOrderGlobal,
		Status:       item.Status,
		ScheduleDate: normalizeDate(date).Format("2006-01-02"),
		AwarenessId:  item.AwarenessId,
		ReferenceMin: nullDecimalString(item.ReferenceMin),
		ReferenceMax: nullDecimalString(item.ReferenceMax),
	}
	if item.Theme.Valid {
		info.AwarenessTheme = item.Theme.String
	}
	return info
}

func restTopicInfo(date time.Time) types.TopicInfo {
	return types.TopicInfo{
		Id:           0,
		Title:        "本轮结束，休息整合中",
		Summary:      "今天不生成新的练习任务，可以回看历史打卡和到期复盘，把这一轮练过的意识点整合一下。",
		OrderNo:      0,
		Status:       1,
		ScheduleDate: normalizeDate(date).Format("2006-01-02"),
		IsRestDay:    true,
	}
}
