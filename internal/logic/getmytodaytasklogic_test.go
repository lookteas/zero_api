package logic

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"api/internal/config"
	"api/internal/svc"
	"api/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type recordingDailyTasksModel struct {
	model.DailyTasksModel
	seenTaskDate time.Time
	item         *model.DailyTasks
	err          error
	updatedItem  *model.DailyTasks
}

func (s *recordingDailyTasksModel) Insert(context.Context, *model.DailyTasks) (sql.Result, error) {
	panic("unexpected call")
}

func (s *recordingDailyTasksModel) FindOne(context.Context, uint64) (*model.DailyTasks, error) {
	panic("unexpected call")
}

func (s *recordingDailyTasksModel) FindOneByUserIdTaskDate(_ context.Context, _ uint64, taskDate time.Time) (*model.DailyTasks, error) {
	s.seenTaskDate = taskDate
	return s.item, s.err
}

func (s *recordingDailyTasksModel) Update(_ context.Context, item *model.DailyTasks) error {
	s.updatedItem = item
	s.item = item
	return nil
}

func (s *recordingDailyTasksModel) Delete(context.Context, uint64) error {
	panic("unexpected call")
}

func (s *recordingDailyTasksModel) withSession(sqlx.Session) model.DailyTasksModel {
	return s
}

func TestGetMyTodayTaskUsesDateOnlyLookup(t *testing.T) {
	t.Parallel()

	stub := &recordingDailyTasksModel{
		item: &model.DailyTasks{
			Id:           1,
			UserId:       1,
			TaskDate:     time.Date(2026, 4, 14, 0, 0, 0, 0, time.Local),
			TopicId:      1,
			TopicOrderNo: 1,
			TopicTitle:   "topic",
			TopicSummary: "summary",
			Status:       "draft",
			CreatedAt:    time.Date(2026, 4, 14, 10, 0, 0, 0, time.Local),
			UpdatedAt:    time.Date(2026, 4, 14, 10, 0, 0, 0, time.Local),
		},
	}

	logic := NewGetMyTodayTaskLogic(context.Background(), &svc.ServiceContext{DailyTasksModel: stub})
	_, err := logic.GetMyTodayTask()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stub.seenTaskDate.Hour() != 0 || stub.seenTaskDate.Minute() != 0 || stub.seenTaskDate.Second() != 0 {
		t.Fatalf("expected date-only lookup, got %v", stub.seenTaskDate)
	}
}

func TestNormalizeDateClearsTimePortion(t *testing.T) {
	t.Parallel()

	input := time.Date(2026, 4, 14, 22, 45, 33, 123, time.Local)
	got := normalizeDate(input)
	if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
		t.Fatalf("expected normalized date, got %v", got)
	}
	if got.Year() != 2026 || got.Month() != 4 || got.Day() != 14 {
		t.Fatalf("expected same date, got %v", got)
	}
}

func TestGetMyTodayTaskAttachesAwarenessInfoWithoutRefreshingSnapshot(t *testing.T) {
	t.Parallel()

	taskDate := normalizeDate(time.Now())
	awareness := model.Awareness{
		AwarenessId:     33,
		PointTitle:      "当日意识点",
		Theme:           sql.NullString{String: "今日主题", Valid: true},
		Summary:         sql.NullString{String: "今日摘要", Valid: true},
		Details:         sql.NullString{String: "今日细节", Valid: true},
		SortOrderGlobal: 1,
	}
	dailyTasks := &recordingDailyTasksModel{
		item: &model.DailyTasks{
			Id:           1,
			UserId:       1,
			TaskDate:     taskDate,
			AwarenessId:  sql.NullInt64{Int64: int64(awareness.AwarenessId), Valid: true},
			TopicId:      0,
			TopicOrderNo: 1,
			TopicTitle:   "snapshot title",
			TopicSummary: "snapshot summary",
			Status:       "draft",
			CreatedAt:    taskDate,
			UpdatedAt:    taskDate,
		},
	}
	awarenessModel := &recordingAwarenessModel{points: []model.Awareness{awareness}}

	logic := NewGetMyTodayTaskLogic(context.Background(), &svc.ServiceContext{
		Config:          config.Config{AwarenessCycle: config.AwarenessCycleConf{StartDate: taskDate.Format("2006-01-02"), RestDays: 7}},
		DailyTasksModel: dailyTasks,
		AwarenessModel:  awarenessModel,
	})
	resp, err := logic.GetMyTodayTask()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dailyTasks.updatedItem != nil {
		t.Fatalf("expected no task snapshot refresh")
	}
	if resp.Data.TopicTitle != "snapshot title" || resp.Data.TopicSummary != "snapshot summary" {
		t.Fatalf("expected existing snapshot fields preserved, got %+v", resp.Data)
	}
	if resp.Data.AwarenessId != awareness.AwarenessId || resp.Data.AwarenessTheme != awareness.Theme.String {
		t.Fatalf("expected attached awareness fields, got %+v", resp.Data)
	}
}

func TestGetMyTodayTaskDoesNotAttachCycleAwarenessWhenStoredAwarenessIdIsMissing(t *testing.T) {
	t.Parallel()

	taskDate := normalizeDate(time.Now())
	dailyTasks := &recordingDailyTasksModel{
		item: &model.DailyTasks{
			Id:           1,
			UserId:       1,
			TaskDate:     taskDate,
			AwarenessId:  sql.NullInt64{Int64: 99, Valid: true},
			TopicId:      0,
			TopicOrderNo: 99,
			TopicTitle:   "stored snapshot title",
			TopicSummary: "stored snapshot summary",
			Status:       "draft",
			CreatedAt:    taskDate,
			UpdatedAt:    taskDate,
		},
	}
	awarenessModel := &recordingAwarenessModel{
		points: []model.Awareness{
			{
				AwarenessId:  1,
				PointTitle:   "current cycle title",
				Theme:        sql.NullString{String: "current theme", Valid: true},
				Details:      sql.NullString{String: "current details", Valid: true},
				ReferenceMin: sql.NullFloat64{Float64: 1, Valid: true},
			},
		},
	}

	logic := NewGetMyTodayTaskLogic(context.Background(), &svc.ServiceContext{
		Config:          config.Config{AwarenessCycle: config.AwarenessCycleConf{StartDate: taskDate.Format("2006-01-02"), RestDays: 7}},
		DailyTasksModel: dailyTasks,
		AwarenessModel:  awarenessModel,
	})
	resp, err := logic.GetMyTodayTask()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Data.AwarenessId != 99 {
		t.Fatalf("expected stored awareness id 99, got %d", resp.Data.AwarenessId)
	}
	if resp.Data.AwarenessTitle != "" || resp.Data.AwarenessDetails != "" || resp.Data.ReferenceMin != "" {
		t.Fatalf("expected no unrelated awareness metadata, got %+v", resp.Data)
	}
	if resp.Data.TopicTitle != "stored snapshot title" || resp.Data.TopicSummary != "stored snapshot summary" {
		t.Fatalf("expected stored snapshot fields preserved, got %+v", resp.Data)
	}
}

func TestGetMyTodayTaskReturnsRestStateWhenNoTaskExists(t *testing.T) {
	t.Parallel()

	dailyTasks := &recordingDailyTasksModel{err: model.ErrNotFound}
	awarenessModel := &recordingAwarenessModel{
		points: []model.Awareness{
			{AwarenessId: 1, PointTitle: "only point", SortOrderGlobal: 1},
		},
	}

	logic := NewGetMyTodayTaskLogic(context.Background(), &svc.ServiceContext{
		Config:          config.Config{AwarenessCycle: config.AwarenessCycleConf{StartDate: normalizeDate(time.Now()).AddDate(0, 0, -1).Format("2006-01-02"), RestDays: 7}},
		DailyTasksModel: dailyTasks,
		AwarenessModel:  awarenessModel,
	})
	resp, err := logic.GetMyTodayTask()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Data.IsRestDay {
		t.Fatalf("expected rest daily task info, got %+v", resp.Data)
	}
	if resp.Data.RestTitle != "本轮结束，休息整合中" {
		t.Fatalf("unexpected rest title %q", resp.Data.RestTitle)
	}
}

func TestGetMyTodayTaskRefreshesDraftSnapshotFromStoredCycleSettings(t *testing.T) {
	t.Parallel()

	taskDate := normalizeDate(time.Now())
	expected := model.Awareness{
		AwarenessId:     302,
		PointTitle:      "生命合一指数",
		Summary:         sql.NullString{String: "生命合一摘要", Valid: true},
		SortOrderGlobal: 2,
	}
	dailyTasks := &recordingDailyTasksModel{
		item: &model.DailyTasks{
			Id:           1,
			UserId:       1,
			TaskDate:     taskDate,
			AwarenessId:  sql.NullInt64{Int64: 301, Valid: true},
			TopicOrderNo: 1,
			TopicTitle:   "性别偏见程度 vs. 性别平等程度",
			TopicSummary: "旧摘要",
			Weakness:     sql.NullString{String: "已经填写的内容", Valid: true},
			Status:       "draft",
			CreatedAt:    taskDate,
			UpdatedAt:    taskDate,
		},
	}

	logic := NewGetMyTodayTaskLogic(context.Background(), &svc.ServiceContext{
		DailyTasksModel: dailyTasks,
		AwarenessModel: &recordingAwarenessModel{points: []model.Awareness{
			{AwarenessId: 301, PointTitle: "性别偏见程度 vs. 性别平等程度", SortOrderGlobal: 1},
			expected,
		}},
		DB: createTaskStoredSettingsDB{settings: map[string]string{
			awarenessCycleStartDateKey: taskDate.AddDate(0, 0, -1).Format("2006-01-02"),
			awarenessCycleRestDaysKey:  "7",
		}},
	})

	resp, err := logic.GetMyTodayTask()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dailyTasks.updatedItem == nil {
		t.Fatalf("expected draft task snapshot refresh")
	}
	if resp.Data.AwarenessId != expected.AwarenessId || resp.Data.TopicTitle != expected.PointTitle {
		t.Fatalf("expected current cycle awareness %+v, got %+v", expected, resp.Data)
	}
	if resp.Data.Weakness != "已经填写的内容" {
		t.Fatalf("expected existing draft content preserved, got %+v", resp.Data)
	}
}
