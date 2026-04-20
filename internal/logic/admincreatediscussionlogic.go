// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
    "context"
    "strings"

    "api/internal/svc"
    "api/internal/types"
    "api/model"

    "github.com/zeromicro/go-zero/core/logx"
)

type AdminCreateDiscussionLogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewAdminCreateDiscussionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreateDiscussionLogic {
    return &AdminCreateDiscussionLogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *AdminCreateDiscussionLogic) AdminCreateDiscussion(req *types.DiscussionCreateReq) (resp *types.SimpleResp, err error) {
    if err = requireAdminUser(l.ctx); err != nil {
        return nil, err
    }
    if l.svcCtx.DB == nil {
        return okSimple("讨论说明已创建（演示数据）"), nil
    }

    weekStart := parseTaskDate(req.WeekStartDate)
    meetingTime := parseDiscussionMeetingTime(req.MeetingTime, weekStart)
    topicTitle := strings.TrimSpace(req.TopicTitle)
    discussionTitle := strings.TrimSpace(req.DiscussionTitle)
    if discussionTitle == "" {
        discussionTitle = defaultDiscussionTitle(topicTitle)
    }

    existing, findErr := l.svcCtx.DiscussionInfosModel.FindOneByWeekStartDate(l.ctx, weekStart)
    if findErr == nil && existing != nil {
        existing.TopicId = req.TopicId
        existing.TopicTitle = topicTitle
        existing.DiscussionTitle = discussionTitle
        existing.Description = nullString(strings.TrimSpace(req.Description))
        existing.Goals = nullString(strings.TrimSpace(req.Goals))
        existing.MeetingTime = meetingTime
        existing.MeetingLink = strings.TrimSpace(req.MeetingLink)
        existing.ShareText = nullString(strings.TrimSpace(req.ShareText))
        if !existing.ShareText.Valid {
            existing.ShareText = nullString(buildDiscussionShareTextFromFields(existing.TopicTitle, existing.DiscussionTitle, existing.MeetingTime.Format("2006-01-02 15:04:05"), existing.MeetingLink))
        }
        existing.Status = normalizeDiscussionStatus(req.Status)
        existing.AdminRemark = nullString(strings.TrimSpace(req.AdminRemark))
        if err = l.svcCtx.DiscussionInfosModel.Update(l.ctx, existing); err != nil {
            return nil, err
        }
        return okSimple("本周讨论已更新"), nil
    }
    if findErr != nil && findErr != model.ErrNotFound {
        return nil, findErr
    }

    meetingLink := strings.TrimSpace(req.MeetingLink)
    shareText := strings.TrimSpace(req.ShareText)
    if shareText == "" {
        shareText = buildDiscussionShareTextFromFields(topicTitle, discussionTitle, meetingTime.Format("2006-01-02 15:04:05"), meetingLink)
    }

    _, err = l.svcCtx.DiscussionInfosModel.Insert(l.ctx, &model.DiscussionInfos{
        WeekStartDate:   weekStart,
        TopicId:         req.TopicId,
        TopicTitle:      topicTitle,
        DiscussionTitle: discussionTitle,
        Description:     nullString(strings.TrimSpace(req.Description)),
        Goals:           nullString(strings.TrimSpace(req.Goals)),
        MeetingTime:     meetingTime,
        MeetingLink:     meetingLink,
        ShareText:       nullString(shareText),
        Status:          normalizeDiscussionStatus(req.Status),
        AdminRemark:     nullString(strings.TrimSpace(req.AdminRemark)),
    })
    if err != nil {
        return nil, err
    }

    return okSimple("本周讨论已创建"), nil
}
