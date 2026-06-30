package logic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateFreemodePracticeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateFreemodePracticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFreemodePracticeLogic {
	return &CreateFreemodePracticeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateFreemodePracticeLogic) CreateFreemodePractice(req *types.FreeModePracticeCreateReq) (*types.FreeModePracticeResp, error) {
	if l.svcCtx.FreeModePracticesModel == nil || l.svcCtx.AwarenessModel == nil {
		return nil, fmt.Errorf("free mode practice storage not ready")
	}

	awareness, err := l.svcCtx.AwarenessModel.FindOne(l.ctx, req.AwarenessId)
	if err != nil {
		return nil, fmt.Errorf("query awareness: %w", err)
	}

	chapterID := req.ChapterId
	if chapterID == 0 {
		chapterID = awareness.ChapterId
	}
	if chapterID != awareness.ChapterId {
		return nil, fmt.Errorf("awareness chapter mismatch")
	}

	chapterTitle := ""
	chapterFullTitle := ""
	chapterNo := int64(0)
	if l.svcCtx.ChaptersModel != nil {
		chapter, chapterErr := l.svcCtx.ChaptersModel.FindOne(l.ctx, chapterID)
		if chapterErr != nil {
			return nil, fmt.Errorf("query chapter: %w", chapterErr)
		}
		chapterTitle = chapter.ChapterTitle
		chapterFullTitle = chapter.ChapterFullTitle
		chapterNo = chapter.ChapterNo
	}

	practiceDate := time.Now().In(time.Local)
	if strings.TrimSpace(req.PracticeDate) != "" {
		parsedDate, parseErr := time.ParseInLocation("2006-01-02", strings.TrimSpace(req.PracticeDate), time.Local)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid practiceDate: %w", parseErr)
		}
		practiceDate = parsedDate
	}

	record := &model.FreeModePractice{
		UserId:           currentUserID(l.ctx),
		PracticeDate:     practiceDate,
		ChapterId:        chapterID,
		ChapterNo:        chapterNo,
		ChapterTitle:     chapterTitle,
		ChapterFullTitle: chapterFullTitle,
		AwarenessId:      awareness.AwarenessId,
		SectionId:        awareness.SectionId,
		AwarenessTitle:   awareness.PointTitle,
		AwarenessSummary: awareness.Summary,
		AwarenessDetails: awareness.Details,
		PracticeNote:     sqlNullString(strings.TrimSpace(req.PracticeNote)),
	}

	result, err := l.svcCtx.FreeModePracticesModel.Insert(l.ctx, record)
	if err != nil {
		return nil, fmt.Errorf("insert free mode practice: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("resolve free mode practice id: %w", err)
	}

	created, err := l.svcCtx.FreeModePracticesModel.FindOne(l.ctx, uint64(id))
	if err != nil {
		return nil, fmt.Errorf("load free mode practice: %w", err)
	}

	return &types.FreeModePracticeResp{
		Code:    0,
		Message: "ok",
		Data:    freeModePracticeToInfo(created),
	}, nil
}

func sqlNullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func freeModePracticeToInfo(item *model.FreeModePractice) types.FreeModePracticeInfo {
	return types.FreeModePracticeInfo{
		PracticeId:       item.PracticeId,
		PracticeDate:     item.PracticeDate.Format("2006-01-02"),
		ChapterId:        item.ChapterId,
		ChapterNo:        item.ChapterNo,
		ChapterTitle:     item.ChapterTitle,
		ChapterFullTitle: item.ChapterFullTitle,
		AwarenessId:      item.AwarenessId,
		SectionId:        item.SectionId,
		AwarenessTitle:   item.AwarenessTitle,
		AwarenessSummary: nullableString(item.AwarenessSummary),
		AwarenessDetails: nullableString(item.AwarenessDetails),
		PracticeNote:     nullableString(item.PracticeNote),
		CreatedAt:        item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        item.UpdatedAt.Format(time.RFC3339),
	}
}
