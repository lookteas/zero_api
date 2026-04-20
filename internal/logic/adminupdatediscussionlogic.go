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

type AdminUpdateDiscussionLogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewAdminUpdateDiscussionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateDiscussionLogic {
    return &AdminUpdateDiscussionLogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *AdminUpdateDiscussionLogic) AdminUpdateDiscussion(req *types.DiscussionUpdateReq) (resp *types.SimpleResp, err error) {
    if err = requireAdminUser(l.ctx); err != nil {
        return nil, err
    }
    if l.svcCtx.DB == nil {
        return okSimple("讨论说明已更新（演示数据）"), nil
    }
    if req.Id == 0 {
        return nil, fmt.Errorf("discussion id is required")
    }

    item, err := l.svcCtx.DiscussionInfosModel.FindOne(l.ctx, req.Id)
    if err != nil {
        return nil, err
    }

    discussionTitle := strings.TrimSpace(req.DiscussionTitle)
    if discussionTitle == "" {
        discussionTitle = item.DiscussionTitle
    }

    item.DiscussionTitle = discussionTitle
    item.Description = nullString(strings.TrimSpace(req.Description))
    item.Goals = nullString(strings.TrimSpace(req.Goals))
    item.MeetingTime = parseDiscussionMeetingTime(req.MeetingTime, item.WeekStartDate)
    item.MeetingLink = strings.TrimSpace(req.MeetingLink)
    item.ShareText = nullString(strings.TrimSpace(req.ShareText))
    if !item.ShareText.Valid {
        item.ShareText = nullString(buildDiscussionShareTextFromFields(item.TopicTitle, item.DiscussionTitle, item.MeetingTime.Format("2006-01-02 15:04:05"), item.MeetingLink))
    }
    item.Status = normalizeDiscussionStatus(req.Status)
    item.AdminRemark = nullString(strings.TrimSpace(req.AdminRemark))

    if err = l.svcCtx.DiscussionInfosModel.Update(l.ctx, item); err != nil {
        return nil, err
    }

    return okSimple("讨论说明已更新"), nil
}
