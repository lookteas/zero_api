// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"fmt"
	"strings"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUpdateTopicLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateTopicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateTopicLogic {
	return &AdminUpdateTopicLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateTopicLogic) AdminUpdateTopic(req *types.TopicUpdateReq) (resp *types.SimpleResp, err error) {
	if err = requireAdminUser(l.ctx); err != nil {
		return nil, err
	}
	if l.svcCtx.DB == nil {
		return okSimple("主题已更新（演示数据）"), nil
	}
	if req.Id == 0 {
		return nil, fmt.Errorf("topic id is required")
	}

	item, err := l.svcCtx.TopicsModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}

	scheduleDate, err := parseTopicScheduleDate(strings.TrimSpace(req.ScheduleDate))
	if err != nil {
		return nil, err
	}

	item.Title = strings.TrimSpace(req.Title)
	item.Summary = strings.TrimSpace(req.Summary)
	item.Description = nullString(strings.TrimSpace(req.Description))
	item.OrderNo = req.OrderNo
	item.Status = uint64(req.Status)
	item.ScheduleDate = scheduleDate

	if err = l.svcCtx.TopicsModel.Update(l.ctx, item); err != nil {
		return nil, mysqlDuplicateMessage("主题", err)
	}

	return okSimple("主题已更新"), nil
}
