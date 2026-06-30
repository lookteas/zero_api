package logic

import (
	"context"
	"fmt"
	"sort"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFreemodeChaptersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListFreemodeChaptersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFreemodeChaptersLogic {
	return &ListFreemodeChaptersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFreemodeChaptersLogic) ListFreemodeChapters() (*types.FreeModeChapterListResp, error) {
	if l.svcCtx.AwarenessModel == nil || l.svcCtx.ChaptersModel == nil {
		return &types.FreeModeChapterListResp{Code: 0, Message: "ok", Data: types.FreeModeChapterListData{}}, nil
	}

	chapters, err := l.svcCtx.ChaptersModel.FindAll(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("query chapters: %w", err)
	}
	points, err := l.svcCtx.AwarenessModel.FindEligible(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("query awareness: %w", err)
	}

	pointGroups := make(map[uint64][]types.FreeModeAwarenessInfo)
	for _, point := range points {
		pointGroups[point.ChapterId] = append(pointGroups[point.ChapterId], awarenessToFreeModeInfo(point))
	}
	for chapterID := range pointGroups {
		sort.SliceStable(pointGroups[chapterID], func(i, j int) bool {
			if pointGroups[chapterID][i].OrderNo == pointGroups[chapterID][j].OrderNo {
				return pointGroups[chapterID][i].AwarenessId < pointGroups[chapterID][j].AwarenessId
			}
			return pointGroups[chapterID][i].OrderNo < pointGroups[chapterID][j].OrderNo
		})
	}

	list := make([]types.FreeModeChapterInfo, 0, len(chapters))
	for _, chapter := range chapters {
		list = append(list, types.FreeModeChapterInfo{
			ChapterId:        chapter.ChapterId,
			ChapterNo:        chapter.ChapterNo,
			ChapterTitle:     chapter.ChapterTitle,
			ChapterFullTitle: chapter.ChapterFullTitle,
			Points:           pointGroups[chapter.ChapterId],
		})
	}

	return &types.FreeModeChapterListResp{
		Code:    0,
		Message: "ok",
		Data: types.FreeModeChapterListData{
			List: list,
		},
	}, nil
}

func awarenessToFreeModeInfo(point model.Awareness) types.FreeModeAwarenessInfo {
	info := types.FreeModeAwarenessInfo{
		AwarenessId:  point.AwarenessId,
		ChapterId:    point.ChapterId,
		SectionId:    point.SectionId,
		Title:        point.PointTitle,
		Summary:      nullableString(point.Summary),
		Details:      nullableString(point.Details),
		OrderNo:      point.SortOrderGlobal,
		ReferenceMin: nullDecimalString(point.ReferenceMin),
		ReferenceMax: nullDecimalString(point.ReferenceMax),
	}
	if point.Theme.Valid {
		info.Theme = point.Theme.String
	}
	return info
}
