package logic

import (
	"context"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRecoveryReviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateRecoveryReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRecoveryReviewLogic {
	return &CreateRecoveryReviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateRecoveryReviewLogic) CreateRecoveryReview(req *types.RecoveryReviewCreateReq) (resp *types.SimpleResp, err error) {
	if l.svcCtx.ReviewRecordsModel == nil || l.svcCtx.ReviewItemsModel == nil {
		return okSimple("recovery review submitted"), nil
	}

	userID := currentUserID(l.ctx)
	now := time.Now()
	items := make([]*model.ReviewItems, 0, len(req.ReviewItemIds))
	for _, reviewItemID := range req.ReviewItemIds {
		item, findErr := l.svcCtx.ReviewItemsModel.FindOne(l.ctx, reviewItemID)
		if findErr != nil {
			return nil, findErr
		}
		items = append(items, item)
	}

	if err = validateRecoveryReviewItems(items, userID, now); err != nil {
		return nil, err
	}

	reviewReq := &types.ReviewRecordCreateReq{Result: req.Result, ActualSituation: req.ActualSituation, Suggestion: req.Suggestion}
	for _, item := range items {
		if err = saveReviewRecordAndCompleteItem(l.ctx, l.svcCtx, item, reviewReq, userID, now); err != nil {
			return nil, err
		}
	}

	return okSimple("recovery review submitted"), nil
}
