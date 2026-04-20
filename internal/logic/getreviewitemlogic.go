// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReviewItemLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetReviewItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReviewItemLogic {
	return &GetReviewItemLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetReviewItemLogic) GetReviewItem(req *types.IdPathReq) (resp *types.ReviewItemResp, err error) {
	if l.svcCtx.ReviewItemsModel == nil {
		return okReviewItem(), nil
	}

	item, err := l.svcCtx.ReviewItemsModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return okReviewItem(), nil
	}

	task, _ := l.svcCtx.DailyTasksModel.FindOne(l.ctx, item.DailyTaskId)
	record, _ := l.svcCtx.ReviewRecordsModel.FindOneByReviewItemId(l.ctx, item.Id)

	var taskInfo *types.DailyTaskInfo
	if task != nil {
		mapped := dailyTaskToInfo(task)
		taskInfo = &mapped
	}

	var recordInfo *types.ReviewRecordInfo
	if record != nil {
		mapped := reviewRecordToInfo(record)
		recordInfo = &mapped
	}

	return &types.ReviewItemResp{Code: 0, Message: "ok", Data: reviewItemToInfo(item, taskInfo, recordInfo)}, nil
}
