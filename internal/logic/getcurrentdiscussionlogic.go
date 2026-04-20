// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCurrentDiscussionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCurrentDiscussionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrentDiscussionLogic {
	return &GetCurrentDiscussionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCurrentDiscussionLogic) GetCurrentDiscussion() (resp *types.DiscussionResp, err error) {
	if l.svcCtx.DB == nil {
		return okDiscussion(), nil
	}

	bundle, err := loadOrCreateCurrentWeeklyVote(l.ctx, l.svcCtx, time.Now(), currentUserID(l.ctx))
	if err != nil {
		return nil, err
	}

	weekStart := currentWeekStart(time.Now())
	item, err := l.svcCtx.DiscussionInfosModel.FindOneByWeekStartDate(l.ctx, weekStart)
	if err == nil && item != nil {
		info := discussionToInfo(item)
		if bundle.Winner != nil {
			info.TopicId = bundle.Winner.TopicId
			info.TopicTitle = bundle.Winner.TopicTitle
			if info.DiscussionTitle == "" || info.DiscussionTitle == defaultDiscussionTitle(item.TopicTitle) {
				info.DiscussionTitle = defaultDiscussionTitle(bundle.Winner.TopicTitle)
			}
			if info.ShareText == "" {
				info.ShareText = buildDiscussionShareText(info)
			}
		}
		return &types.DiscussionResp{Code: 0, Message: "ok", Data: info}, nil
	}
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}

	info := buildDerivedDiscussion(bundle)
	return &types.DiscussionResp{Code: 0, Message: "ok", Data: info}, nil
}
