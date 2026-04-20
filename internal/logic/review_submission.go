package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"
)

var errInvalidRecoveryReviewItems = errors.New("recovery review items must belong to the same task and already be due")

func validateReviewItemSubmission(item *model.ReviewItems, userID uint64, now time.Time) error {
	if item == nil || item.UserId != userID {
		return model.ErrNotFound
	}
	if item.Status != "pending" {
		return errors.New("review item has already been completed")
	}
	if item.DueAt.After(now) {
		return errors.New("review item is not due yet")
	}
	return nil
}

func validateRecoveryReviewItems(items []*model.ReviewItems, userID uint64, now time.Time) error {
	if len(items) == 0 {
		return errInvalidRecoveryReviewItems
	}
	var dailyTaskID uint64
	for _, item := range items {
		if err := validateReviewItemSubmission(item, userID, now); err != nil {
			return err
		}
		if dailyTaskID == 0 {
			dailyTaskID = item.DailyTaskId
			continue
		}
		if item.DailyTaskId != dailyTaskID {
			return errInvalidRecoveryReviewItems
		}
	}
	return nil
}

func saveReviewRecordAndCompleteItem(ctx context.Context, svcCtx *svc.ServiceContext, item *model.ReviewItems, req *types.ReviewRecordCreateReq, userID uint64, now time.Time) error {
	existing, findErr := svcCtx.ReviewRecordsModel.FindOneByReviewItemId(ctx, item.Id)
	if findErr == nil {
		existing.Result = req.Result
		existing.ActualSituation = nullString(req.ActualSituation)
		existing.Suggestion = nullString(req.Suggestion)
		existing.SubmittedAt = now
		if err := svcCtx.ReviewRecordsModel.Update(ctx, existing); err != nil {
			return err
		}
	} else {
		if findErr != model.ErrNotFound {
			return findErr
		}
		_, err := svcCtx.ReviewRecordsModel.Insert(ctx, &model.ReviewRecords{
			ReviewItemId:    item.Id,
			UserId:          userID,
			Result:          req.Result,
			ActualSituation: nullString(req.ActualSituation),
			Suggestion:      nullString(req.Suggestion),
			SubmittedAt:     now,
		})
		if err != nil {
			return err
		}
	}

	item.Status = "completed"
	item.CompletedAt = sql.NullTime{Time: now, Valid: true}
	return svcCtx.ReviewItemsModel.Update(ctx, item)
}
