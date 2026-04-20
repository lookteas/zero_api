// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"fmt"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCurrentWeeklyVoteRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCurrentWeeklyVoteRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCurrentWeeklyVoteRecordLogic {
	return &CreateCurrentWeeklyVoteRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCurrentWeeklyVoteRecordLogic) CreateCurrentWeeklyVoteRecord(req *types.VoteRecordCreateReq) (resp *types.SimpleResp, err error) {
	if l.svcCtx.DB == nil {
		return okSimple("投票成功（演示数据）"), nil
	}

	if req.CandidateId == 0 {
		return nil, fmt.Errorf("candidateId is required")
	}

	now := time.Now()
	userID := currentUserID(l.ctx)
	bundle, err := loadOrCreateCurrentWeeklyVote(l.ctx, l.svcCtx, now, userID)
	if err != nil {
		return nil, err
	}
	if bundle.TodayHasVoted {
		return nil, fmt.Errorf("今天已经投过票了，明天再来")
	}

	if now.After(bundle.Vote.VoteEndAt) {
		return nil, fmt.Errorf("本周投票已截止，讨论主题已经确定")
	}

	candidate, err := l.svcCtx.WeeklyTopicVoteCandidatesModel.FindOne(l.ctx, req.CandidateId)
	if err != nil {
		return nil, err
	}
	if candidate.WeeklyVoteId != bundle.Vote.Id {
		return nil, fmt.Errorf("candidate does not belong to current weekly vote")
	}

	_, err = l.svcCtx.WeeklyTopicVoteRecordsModel.Insert(l.ctx, &model.WeeklyTopicVoteRecords{
		WeeklyVoteId: bundle.Vote.Id,
		UserId:       userID,
		CandidateId:  req.CandidateId,
	})
	if err != nil {
		return nil, err
	}

	if _, err = loadOrCreateCurrentWeeklyVote(l.ctx, l.svcCtx, now, userID); err != nil {
		return nil, err
	}

	return okSimple("已记录这次投票"), nil
}
