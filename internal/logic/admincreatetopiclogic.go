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

type AdminCreateTopicLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminCreateTopicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreateTopicLogic {
	return &AdminCreateTopicLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminCreateTopicLogic) AdminCreateTopic(req *types.TopicCreateReq) (resp *types.SimpleResp, err error) {
	if err = requireAdminUser(l.ctx); err != nil {
		return nil, err
	}
	if l.svcCtx.DB == nil {
		return okSimple("主题已创建（演示数据）"), nil
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	orderNo := req.OrderNo
	if orderNo <= 0 {
		var items []model.Topics
		if err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &items, "select id, title, summary, description, order_no, status, schedule_date, created_at, updated_at from topics order by order_no asc"); err != nil {
			return nil, err
		}
		orderNo = nextTopicOrderNo(items)
	}

	scheduleDate, err := parseTopicScheduleDate(strings.TrimSpace(req.ScheduleDate))
	if err != nil {
		return nil, err
	}

	status := uint64(1)
	if req.Status == 0 {
		status = 0
	}

	_, err = l.svcCtx.TopicsModel.Insert(l.ctx, &model.Topics{
		Title:        title,
		Summary:      strings.TrimSpace(req.Summary),
		Description:  nullString(strings.TrimSpace(req.Description)),
		OrderNo:      orderNo,
		Status:       status,
		ScheduleDate: scheduleDate,
	})
	if err != nil {
		return nil, mysqlDuplicateMessage("主题", err)
	}

	return okSimple("主题已创建"), nil
}
