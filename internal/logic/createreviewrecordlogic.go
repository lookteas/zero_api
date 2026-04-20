// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"time"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateReviewRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateReviewRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateReviewRecordLogic {
	return &CreateReviewRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateReviewRecordLogic) CreateReviewRecord(reviewItemId uint64, req *types.ReviewRecordCreateReq) (resp *types.SimpleResp, err error) {
	if l.svcCtx.ReviewRecordsModel == nil || l.svcCtx.ReviewItemsModel == nil {
		return okSimple("review submitted"), nil
	}

	item, err := l.svcCtx.ReviewItemsModel.FindOne(l.ctx, reviewItemId)
	if err != nil {
		return nil, err
	}

	userID := currentUserID(l.ctx)
	now := time.Now()
	if err = validateReviewItemSubmission(item, userID, now); err != nil {
		return nil, err
	}
	if err = saveReviewRecordAndCompleteItem(l.ctx, l.svcCtx, item, req, userID, now); err != nil {
		return nil, err
	}

	return okSimple("review submitted"), nil
}
