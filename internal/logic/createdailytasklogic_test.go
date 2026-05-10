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

type stubInsertResult struct {
	id int64
}

func (s stubInsertResult) LastInsertId() (int64, error) { return s.id, nil }
func (s stubInsertResult) RowsAffected() (int64, error) { return 1, nil }

type createDailyTasksModel struct {
	model.DailyTasksModel
	findErr    error
	inserted   *model.DailyTasks
	storedItem *model.DailyTasks
}

func (m *createDailyTasksModel) FindOneByUserIdTaskDate(context.Context, uint64, time.Time) (*model.DailyTasks, error) {
	return nil, m.findErr
}
func (m *createDailyTasksModel) Insert(_ context.Context, data *model.DailyTasks) (sql.Result, error) {
	m.inserted = data
	return stubInsertResult{id: 9}, nil
}
func (m *createDailyTasksModel) FindOne(context.Context, uint64) (*model.DailyTasks, error) {
	return m.storedItem, nil
}
func (m *createDailyTasksModel) Update(context.Context, *model.DailyTasks) error {
	panic("unexpected call")
}
func (m *createDailyTasksModel) Delete(context.Context, uint64) error           { panic("unexpected call") }
func (m *createDailyTasksModel) withSession(sqlx.Session) model.DailyTasksModel { return m }

type recordingAwarenessModel struct {
	model.AwarenessModel
	points []model.Awareness
	err    error
	called bool
}

func (m *recordingAwarenessModel) FindEligible(context.Context) ([]model.Awareness, error) {
	m.called = true
	return m.points, m.err
}
func (m *recordingAwarenessModel) withSession(sqlx.Session) model.AwarenessModel { return m }

func TestCreateDailyTaskUsesAwarenessCycleForTaskDate(t *testing.T) {
	t.Parallel()

	taskDate := time.Date(2026, 5, 2, 0, 0, 0, 0, time.Local)
	awareness := model.Awareness{
		AwarenessId:     102,
		PointTitle:      "觉察边界",
		Theme:           sql.NullString{String: "边界主题", Valid: true},
		Summary:         sql.NullString{String: "边界摘要", Valid: true},
		Details:         sql.NullString{String: "边界细节", Valid: true},
		ReferenceMin:    sql.NullFloat64{Float64: 1.2, Valid: true},
		ReferenceMax:    sql.NullFloat64{Float64: 3.4, Valid: true},
		BetterDirection: "higher",
		SortOrderGlobal: 2,
	}
	dailyTasks := &createDailyTasksModel{
		findErr: model.ErrNotFound,
		storedItem: &model.DailyTasks{
			Id:           9,
			UserId:       7,
			TaskDate:     taskDate,
			AwarenessId:  sql.NullInt64{Int64: int64(awareness.AwarenessId), Valid: true},
			TopicId:      0,
			TopicOrderNo: awareness.SortOrderGlobal,
			TopicTitle:   awareness.PointTitle,
			TopicSummary: awareness.Summary.String,
			Status:       "draft",
			CreatedAt:    taskDate,
			UpdatedAt:    taskDate,
		},
	}
	awarenessModel := &recordingAwarenessModel{
		points: []model.Awareness{
			{AwarenessId: 101, PointTitle: "first", SortOrderGlobal: 1},
			awareness,
		},
	}

	ctx := WithCurrentUserID(context.Background(), 7)
	logic := NewCreateDailyTaskLogic(ctx, &svc.ServiceContext{
		DailyTasksModel: dailyTasks,
		AwarenessModel:  awarenessModel,
	})

	resp, err := logic.CreateDailyTask(&types.DailyTaskCreateReq{TaskDate: "2026-05-02"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !awarenessModel.called {
		t.Fatalf("expected awareness lookup")
	}
	if dailyTasks.inserted == nil {
		t.Fatalf("expected daily task insert")
	}
	if !dailyTasks.inserted.AwarenessId.Valid || dailyTasks.inserted.AwarenessId.Int64 != int64(awareness.AwarenessId) {
		t.Fatalf("expected awareness id %d, got %+v", awareness.AwarenessId, dailyTasks.inserted.AwarenessId)
	}
	if dailyTasks.inserted.TopicId != 0 || dailyTasks.inserted.TopicOrderNo != awareness.SortOrderGlobal {
		t.Fatalf("expected awareness snapshot, got topic_id=%d order_no=%d", dailyTasks.inserted.TopicId, dailyTasks.inserted.TopicOrderNo)
	}
	if resp.Data.TopicTitle != awareness.PointTitle {
		t.Fatalf("expected response topic title %q, got %q", awareness.PointTitle, resp.Data.TopicTitle)
	}
	if resp.Data.AwarenessId != awareness.AwarenessId || resp.Data.AwarenessTitle != awareness.PointTitle {
		t.Fatalf("expected awareness response fields, got %+v", resp.Data)
	}
	if resp.Data.ReferenceMin != "1.20" || resp.Data.ReferenceMax != "3.40" {
		t.Fatalf("expected formatted reference range, got min=%q max=%q", resp.Data.ReferenceMin, resp.Data.ReferenceMax)
	}
}
