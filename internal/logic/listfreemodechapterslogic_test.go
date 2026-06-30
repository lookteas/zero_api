package logic

import (
	"context"
	"database/sql"
	"testing"

	"api/internal/svc"
	"api/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type recordingFreemodeAwarenessModel struct {
	model.AwarenessModel
	points []model.Awareness
	called bool
}

func (m *recordingFreemodeAwarenessModel) FindEligible(context.Context) ([]model.Awareness, error) {
	m.called = true
	return m.points, nil
}

func (m *recordingFreemodeAwarenessModel) withSession(sqlx.Session) model.AwarenessModel { return m }

type recordingFreemodeChaptersModel struct {
	model.ChaptersModel
	chapters []model.Chapter
	called   bool
}

func (m *recordingFreemodeChaptersModel) FindAll(context.Context) ([]model.Chapter, error) {
	m.called = true
	return m.chapters, nil
}

func (m *recordingFreemodeChaptersModel) withSession(sqlx.Session) model.ChaptersModel { return m }

func TestListFreemodeChaptersGroupsPointsByChapter(t *testing.T) {
	t.Parallel()

	awarenessModel := &recordingFreemodeAwarenessModel{
		points: []model.Awareness{
			{
				AwarenessId:     11,
				ChapterId:       1,
				SectionId:       101,
				PointTitle:      "看见紧绷",
				Summary:         sql.NullString{String: "身体先紧起来", Valid: true},
				Details:         sql.NullString{String: "肩颈和呼吸都能感觉到", Valid: true},
				SortOrderGlobal: 2,
				Status:          1,
			},
			{
				AwarenessId:     10,
				ChapterId:       1,
				SectionId:       100,
				PointTitle:      "辨认触发",
				Summary:         sql.NullString{String: "先看到自己被什么触发", Valid: true},
				SortOrderGlobal: 1,
				Status:          1,
			},
			{
				AwarenessId:     21,
				ChapterId:       2,
				SectionId:       201,
				PointTitle:      "停一下",
				Summary:         sql.NullString{String: "先暂停再反应", Valid: true},
				SortOrderGlobal: 1,
				Status:          1,
			},
		},
	}
	chaptersModel := &recordingFreemodeChaptersModel{
		chapters: []model.Chapter{
			{ChapterId: 1, ChapterNo: 1, ChapterTitle: "第一章", ChapterFullTitle: "第一章 先看见自己", SortOrder: 1},
			{ChapterId: 2, ChapterNo: 2, ChapterTitle: "第二章", ChapterFullTitle: "第二章 先稳住", SortOrder: 2},
		},
	}

	logic := NewListFreemodeChaptersLogic(context.Background(), &svc.ServiceContext{
		AwarenessModel: awarenessModel,
		ChaptersModel:  chaptersModel,
	})

	resp, err := logic.ListFreemodeChapters()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !awarenessModel.called {
		t.Fatalf("expected awareness lookup")
	}
	if !chaptersModel.called {
		t.Fatalf("expected chapter lookup")
	}
	if len(resp.Data.List) != 2 {
		t.Fatalf("expected 2 chapter groups, got %d", len(resp.Data.List))
	}
	if resp.Data.List[0].ChapterId != 1 || resp.Data.List[0].ChapterTitle != "第一章" {
		t.Fatalf("expected first chapter group, got %+v", resp.Data.List[0])
	}
	if len(resp.Data.List[0].Points) != 2 {
		t.Fatalf("expected two points in first chapter, got %+v", resp.Data.List[0].Points)
	}
	if resp.Data.List[0].Points[0].AwarenessId != 10 || resp.Data.List[0].Points[1].AwarenessId != 11 {
		t.Fatalf("expected chapter points ordered by global sort, got %+v", resp.Data.List[0].Points)
	}
	if resp.Data.List[1].ChapterId != 2 || len(resp.Data.List[1].Points) != 1 {
		t.Fatalf("expected second chapter group, got %+v", resp.Data.List[1])
	}
	if resp.Data.List[1].Points[0].Title != "停一下" {
		t.Fatalf("expected second chapter point, got %+v", resp.Data.List[1].Points[0])
	}
}

func TestListFreemodeChaptersKeepsEmptyChaptersVisible(t *testing.T) {
	t.Parallel()

	awarenessModel := &recordingFreemodeAwarenessModel{
		points: []model.Awareness{
			{AwarenessId: 11, ChapterId: 2, SectionId: 201, PointTitle: "点二", SortOrderGlobal: 2, Status: 1},
		},
	}
	chaptersModel := &recordingFreemodeChaptersModel{
		chapters: []model.Chapter{
			{ChapterId: 1, ChapterNo: 1, ChapterTitle: "第一章", ChapterFullTitle: "第一章 空章节", SortOrder: 1},
			{ChapterId: 2, ChapterNo: 2, ChapterTitle: "第二章", ChapterFullTitle: "第二章 有内容", SortOrder: 2},
		},
	}

	logic := NewListFreemodeChaptersLogic(context.Background(), &svc.ServiceContext{
		AwarenessModel: awarenessModel,
		ChaptersModel:  chaptersModel,
	})

	resp, err := logic.ListFreemodeChapters()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Data.List) != 2 {
		t.Fatalf("expected 2 chapter groups, got %d", len(resp.Data.List))
	}
	if len(resp.Data.List[0].Points) != 0 {
		t.Fatalf("expected empty first chapter, got %+v", resp.Data.List[0].Points)
	}
	if len(resp.Data.List[1].Points) != 1 || resp.Data.List[1].Points[0].AwarenessId != 11 {
		t.Fatalf("expected second chapter to keep its point, got %+v", resp.Data.List[1].Points)
	}
}
