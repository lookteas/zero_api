package logic

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type stubFreemodeInsertResult struct {
	id int64
}

func (s stubFreemodeInsertResult) LastInsertId() (int64, error) { return s.id, nil }
func (s stubFreemodeInsertResult) RowsAffected() (int64, error) { return 1, nil }

type freemodePracticeAwarenessModel struct {
	model.AwarenessModel
	awareness *model.Awareness
	called    bool
}

func (m *freemodePracticeAwarenessModel) FindOne(context.Context, uint64) (*model.Awareness, error) {
	m.called = true
	return m.awareness, nil
}

func (m *freemodePracticeAwarenessModel) FindEligible(context.Context) ([]model.Awareness, error) {
	return nil, nil
}

func (m *freemodePracticeAwarenessModel) CreateMinimal(context.Context, string, string, string) (*model.Awareness, error) {
	panic("unexpected call")
}

func (m *freemodePracticeAwarenessModel) MoveToPosition(context.Context, uint64, int64) error {
	panic("unexpected call")
}
func (m *freemodePracticeAwarenessModel) UpdateContent(context.Context, uint64, string, string, string) error {
	panic("unexpected call")
}
func (m *freemodePracticeAwarenessModel) Disable(context.Context, uint64) error {
	panic("unexpected call")
}
func (m *freemodePracticeAwarenessModel) withSession(sqlx.Session) model.AwarenessModel { return m }

type freemodePracticeChaptersModel struct {
	model.ChaptersModel
	chapter *model.Chapter
	called  bool
}

func (m *freemodePracticeChaptersModel) FindAll(context.Context) ([]model.Chapter, error) {
	return []model.Chapter{*m.chapter}, nil
}

func (m *freemodePracticeChaptersModel) FindOne(context.Context, uint64) (*model.Chapter, error) {
	m.called = true
	return m.chapter, nil
}

func (m *freemodePracticeChaptersModel) withSession(sqlx.Session) model.ChaptersModel { return m }

type freemodePracticeStoreModel struct {
	model.FreeModePracticesModel
	inserted *model.FreeModePractice
	stored   *model.FreeModePractice
	updated  *model.FreeModePractice
}

func (m *freemodePracticeStoreModel) Insert(_ context.Context, data *model.FreeModePractice) (sql.Result, error) {
	m.inserted = data
	return stubFreemodeInsertResult{id: 88}, nil
}

func (m *freemodePracticeStoreModel) FindOne(context.Context, uint64) (*model.FreeModePractice, error) {
	return m.stored, nil
}

func (m *freemodePracticeStoreModel) FindByUserID(context.Context, uint64, int64) ([]model.FreeModePractice, error) {
	return nil, nil
}

func (m *freemodePracticeStoreModel) Update(_ context.Context, data *model.FreeModePractice) error {
	m.updated = data
	m.stored = data
	return nil
}

func (m *freemodePracticeStoreModel) withSession(sqlx.Session) model.FreeModePracticesModel { return m }

func TestCreateFreemodePracticeStoresIndependentRecord(t *testing.T) {
	t.Parallel()

	awareness := &model.Awareness{
		AwarenessId:     301,
		ChapterId:       9,
		SectionId:       901,
		PointTitle:      "看见冲动",
		Summary:         sql.NullString{String: "先看见冲动冒出来", Valid: true},
		Details:         sql.NullString{String: "把冲动拆开看", Valid: true},
		SortOrderGlobal: 12,
		Theme:           sql.NullString{String: "冲动", Valid: true},
		ReferenceMin:    sql.NullFloat64{Float64: 1.50, Valid: true},
		ReferenceMax:    sql.NullFloat64{Float64: 4.20, Valid: true},
		BetterDirection: "higher",
	}
	chapter := &model.Chapter{
		ChapterId:        9,
		ChapterNo:        9,
		ChapterTitle:     "第九章",
		ChapterFullTitle: "第九章 自由练习",
		SortOrder:        9,
	}
	store := &freemodePracticeStoreModel{}
	store.stored = &model.FreeModePractice{
		PracticeId:       88,
		UserId:           7,
		PracticeDate:     time.Date(2026, 6, 30, 0, 0, 0, 0, time.Local),
		ChapterId:        chapter.ChapterId,
		ChapterNo:        chapter.ChapterNo,
		ChapterTitle:     chapter.ChapterTitle,
		ChapterFullTitle: chapter.ChapterFullTitle,
		AwarenessId:      awareness.AwarenessId,
		SectionId:        awareness.SectionId,
		AwarenessTitle:   awareness.PointTitle,
		AwarenessSummary: awareness.Summary,
		AwarenessDetails: awareness.Details,
		PracticeNote:     sql.NullString{String: "先做一次短暂停顿", Valid: true},
		CreatedAt:        time.Date(2026, 6, 30, 12, 0, 0, 0, time.Local),
		UpdatedAt:        time.Date(2026, 6, 30, 12, 0, 0, 0, time.Local),
	}

	logic := NewCreateFreemodePracticeLogic(WithCurrentUserID(context.Background(), 7), &svc.ServiceContext{
		AwarenessModel:         &freemodePracticeAwarenessModel{awareness: awareness},
		ChaptersModel:          &freemodePracticeChaptersModel{chapter: chapter},
		FreeModePracticesModel: store,
	})

	resp, err := logic.CreateFreemodePractice(&types.FreeModePracticeCreateReq{
		ChapterId:    9,
		AwarenessId:  301,
		PracticeNote: "先做一次短暂停顿",
		PracticeDate: "2026-06-30",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.inserted == nil {
		t.Fatalf("expected insert call")
	}
	if store.inserted.UserId != 7 || store.inserted.ChapterId != 9 || store.inserted.AwarenessId != 301 {
		t.Fatalf("expected independent free mode record, got %+v", store.inserted)
	}
	if !awareness.Summary.Valid || !awareness.Details.Valid {
		t.Fatalf("expected awareness source data")
	}
	if resp.Data.PracticeId != 88 || resp.Data.AwarenessId != 301 {
		t.Fatalf("expected stored record response, got %+v", resp.Data)
	}
	if resp.Data.PracticeNote != "先做一次短暂停顿" {
		t.Fatalf("expected note echoed from store, got %+v", resp.Data)
	}
	if resp.Data.ChapterFullTitle != chapter.ChapterFullTitle {
		t.Fatalf("expected chapter title in response, got %+v", resp.Data)
	}
}

type freemodePracticeListStoreModel struct {
	model.FreeModePracticesModel
	userID uint64
	limit  int64
	items  []model.FreeModePractice
}

func (m *freemodePracticeListStoreModel) Insert(context.Context, *model.FreeModePractice) (sql.Result, error) {
	panic("unexpected call")
}

func (m *freemodePracticeListStoreModel) FindOne(context.Context, uint64) (*model.FreeModePractice, error) {
	panic("unexpected call")
}

func (m *freemodePracticeListStoreModel) FindByUserID(_ context.Context, userID uint64, limit int64) ([]model.FreeModePractice, error) {
	m.userID = userID
	m.limit = limit
	return m.items, nil
}

func (m *freemodePracticeListStoreModel) withSession(sqlx.Session) model.FreeModePracticesModel {
	return m
}

func TestListFreemodePracticesReturnsOnlyCurrentUsersItems(t *testing.T) {
	t.Parallel()

	store := &freemodePracticeListStoreModel{
		items: []model.FreeModePractice{
			{
				PracticeId:       9,
				UserId:           7,
				PracticeDate:     time.Date(2026, 6, 30, 0, 0, 0, 0, time.Local),
				ChapterId:        1,
				ChapterNo:        1,
				ChapterTitle:     "第一章",
				ChapterFullTitle: "第一章 先稳住",
				AwarenessId:      11,
				SectionId:        101,
				AwarenessTitle:   "看见紧绷",
				AwarenessSummary: sql.NullString{String: "身体先紧起来", Valid: true},
				AwarenessDetails: sql.NullString{String: "肩颈和呼吸都能感觉到", Valid: true},
				PracticeNote:     sql.NullString{String: "先看再说", Valid: true},
				CreatedAt:        time.Date(2026, 6, 30, 10, 0, 0, 0, time.Local),
				UpdatedAt:        time.Date(2026, 6, 30, 10, 0, 0, 0, time.Local),
			},
		},
	}

	logic := NewListFreemodePracticesLogic(WithCurrentUserID(context.Background(), 7), &svc.ServiceContext{
		FreeModePracticesModel: store,
	})

	resp, err := logic.ListFreemodePractices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.userID != 7 {
		t.Fatalf("expected current user id 7, got %d", store.userID)
	}
	if store.limit != 100 {
		t.Fatalf("expected list limit 100, got %d", store.limit)
	}
	if len(resp.Data.List) != 1 {
		t.Fatalf("expected one practice, got %d", len(resp.Data.List))
	}
	if resp.Data.List[0].AwarenessTitle != "看见紧绷" || resp.Data.List[0].PracticeNote != "先看再说" {
		t.Fatalf("expected stored practice in response, got %+v", resp.Data.List[0])
	}
}

func TestUpdateFreemodePracticeOnlyUpdatesCurrentUsersNote(t *testing.T) {
	t.Parallel()

	store := &freemodePracticeStoreModel{
		stored: &model.FreeModePractice{
			PracticeId:       88,
			UserId:           7,
			PracticeDate:     time.Date(2026, 7, 2, 0, 0, 0, 0, time.Local),
			ChapterId:        2,
			ChapterNo:        2,
			ChapterTitle:     "自主意识区",
			ChapterFullTitle: "第二章 自主意识区",
			AwarenessId:      201,
			SectionId:        20,
			AwarenessTitle:   "看见反应",
			AwarenessSummary: sql.NullString{String: "先看见反应", Valid: true},
			AwarenessDetails: sql.NullString{String: "详情保持不变", Valid: true},
			PracticeNote:     sql.NullString{String: "旧的觉察", Valid: true},
			CreatedAt:        time.Date(2026, 7, 2, 8, 0, 0, 0, time.Local),
			UpdatedAt:        time.Date(2026, 7, 2, 8, 0, 0, 0, time.Local),
		},
	}

	logic := NewUpdateFreemodePracticeLogic(WithCurrentUserID(context.Background(), 7), &svc.ServiceContext{
		FreeModePracticesModel: store,
	})

	resp, err := logic.UpdateFreemodePractice(88, &types.FreeModePracticeUpdateReq{
		PracticeNote: "新的觉察记录，保留和意识点对照",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.updated == nil {
		t.Fatalf("expected update call")
	}
	if store.updated.UserId != 7 || store.updated.PracticeId != 88 {
		t.Fatalf("expected original ownership retained, got %+v", store.updated)
	}
	if !store.updated.PracticeNote.Valid || store.updated.PracticeNote.String != "新的觉察记录，保留和意识点对照" {
		t.Fatalf("expected updated practice note, got %+v", store.updated.PracticeNote)
	}
	if resp.Data.PracticeNote != "新的觉察记录，保留和意识点对照" {
		t.Fatalf("expected response to include updated note, got %+v", resp.Data)
	}
}

func TestUpdateFreemodePracticeRejectsOtherUsersItem(t *testing.T) {
	t.Parallel()

	store := &freemodePracticeStoreModel{
		stored: &model.FreeModePractice{
			PracticeId:   89,
			UserId:       99,
			PracticeDate: time.Date(2026, 7, 2, 0, 0, 0, 0, time.Local),
		},
	}

	logic := NewUpdateFreemodePracticeLogic(WithCurrentUserID(context.Background(), 7), &svc.ServiceContext{
		FreeModePracticesModel: store,
	})

	if _, err := logic.UpdateFreemodePractice(89, &types.FreeModePracticeUpdateReq{PracticeNote: "不能改别人记录"}); err == nil {
		t.Fatalf("expected ownership error")
	}
	if store.updated != nil {
		t.Fatalf("did not expect update for another user's item")
	}
}
