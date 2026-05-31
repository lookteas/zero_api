package logic

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestListDailyTasksBuildQueryAddsAwarenessIdAndKeywordFilterAcrossHistoryFields(t *testing.T) {
	t.Parallel()

	query, args := buildDailyTaskListQuery(7, &types.DailyTaskQueryReq{
		StartDate: "2026-04-10",
		EndDate:   "2026-04-16",
		Keyword:   "复盘",
	})

	expectedQuery := "select id, user_id, community_id, task_date, schedule_day_id, topic_id, awareness_id, topic_order_no, topic_title, topic_summary, weakness, improvement_plan, verification_path, reflection_note, status, submitted_at, created_at, updated_at from daily_tasks where user_id = ? and task_date >= ? and task_date <= ? and (topic_title like ? or weakness like ? or improvement_plan like ? or verification_path like ? or reflection_note like ?) order by task_date desc limit 100"
	if query != expectedQuery {
		t.Fatalf("unexpected query:\n%s", query)
	}

	if len(args) != 8 {
		t.Fatalf("expected 8 args, got %d", len(args))
	}

	for i, arg := range []any{uint64(7), "2026-04-10", "2026-04-16", "%复盘%", "%复盘%", "%复盘%", "%复盘%", "%复盘%"} {
		if args[i] != arg {
			t.Fatalf("unexpected arg %d: %#v", i, args[i])
		}
	}
}

type listDailyTasksModel struct {
	model.DailyTasksModel
	inserted []*model.DailyTasks
	nextID   int64
}

func (m *listDailyTasksModel) Insert(_ context.Context, data *model.DailyTasks) (sql.Result, error) {
	m.inserted = append(m.inserted, data)
	m.nextID++
	return stubInsertResult{id: m.nextID}, nil
}
func (m *listDailyTasksModel) FindOne(context.Context, uint64) (*model.DailyTasks, error) {
	panic("unexpected call")
}
func (m *listDailyTasksModel) FindOneByUserIdTaskDate(context.Context, uint64, time.Time) (*model.DailyTasks, error) {
	panic("unexpected call")
}
func (m *listDailyTasksModel) Update(context.Context, *model.DailyTasks) error {
	panic("unexpected call")
}
func (m *listDailyTasksModel) Delete(context.Context, uint64) error { panic("unexpected call") }
func (m *listDailyTasksModel) withSession(sqlx.Session) model.DailyTasksModel {
	return m
}

type listAwarenessScheduleDaysModel struct {
	model.AwarenessScheduleDaysModel
	items []model.AwarenessScheduleDays
}

func (m *listAwarenessScheduleDaysModel) FindByCommunityDateRange(context.Context, uint64, time.Time, time.Time) ([]model.AwarenessScheduleDays, error) {
	return m.items, nil
}
func (m *listAwarenessScheduleDaysModel) FindOneByCycleIdScheduleDate(context.Context, uint64, time.Time) (*model.AwarenessScheduleDays, error) {
	panic("unexpected call")
}
func (m *listAwarenessScheduleDaysModel) Insert(context.Context, *model.AwarenessScheduleDays) (sql.Result, error) {
	panic("unexpected call")
}
func (m *listAwarenessScheduleDaysModel) Upsert(context.Context, *model.AwarenessScheduleDays) error {
	panic("unexpected call")
}
func (m *listAwarenessScheduleDaysModel) DeleteFutureByCycle(context.Context, uint64, time.Time) error {
	panic("unexpected call")
}
func (m *listAwarenessScheduleDaysModel) withSession(sqlx.Session) model.AwarenessScheduleDaysModel {
	return m
}

func TestFillMissingDatesSkipsPausedScheduleDays(t *testing.T) {
	t.Parallel()

	pausedDate := time.Date(2026, 5, 11, 0, 0, 0, 0, time.Local)
	nextDate := time.Date(2026, 5, 12, 0, 0, 0, 0, time.Local)
	dailyTasks := &listDailyTasksModel{}
	logic := NewListDailyTasksLogic(WithCurrentUserID(context.Background(), 7), &svc.ServiceContext{
		DailyTasksModel: dailyTasks,
		AwarenessScheduleDaysModel: &listAwarenessScheduleDaysModel{items: []model.AwarenessScheduleDays{
			{
				ScheduleDate: pausedDate,
				DayType:      scheduleDayPaused,
			},
			{
				ScheduleDayId:     42,
				CommunityId:       defaultCommunityID,
				ScheduleDate:      nextDate,
				DayType:           scheduleDayNormal,
				AwarenessId:       sql.NullInt64{Int64: 101, Valid: true},
				AwarenessTitle:    sql.NullString{String: "觉察边界", Valid: true},
				AwarenessSummary:  sql.NullString{String: "边界摘要", Valid: true},
				CycleDayIndex:     sql.NullInt64{Int64: 0, Valid: true},
				EffectiveDayIndex: sql.NullInt64{Int64: 0, Valid: true},
				UpdatedAt:         nextDate,
			},
		}},
	})
	existingDates := map[string]bool{}
	list := []types.DailyTaskInfo{}

	logic.fillMissingDates(pausedDate, nextDate, existingDates, &list)

	if len(list) != 1 {
		t.Fatalf("expected only normal schedule day to be listed, got %+v", list)
	}
	if list[0].TaskDate != "2026-05-12" || list[0].TopicTitle != "觉察边界" {
		t.Fatalf("expected next normal schedule task, got %+v", list[0])
	}
	if len(dailyTasks.inserted) != 1 {
		t.Fatalf("expected one inserted task, got %d", len(dailyTasks.inserted))
	}
	if dailyTasks.inserted[0].TaskDate.Format("2006-01-02") == "2026-05-11" {
		t.Fatalf("paused date should not create a draft task")
	}
	if !existingDates["2026-05-11"] || !existingDates["2026-05-12"] {
		t.Fatalf("expected both schedule dates to be marked handled, got %+v", existingDates)
	}
}

type listDailyTasksDB struct {
	sqlx.SqlConn
	items []model.DailyTasks
}

func (db *listDailyTasksDB) QueryRowsCtx(_ context.Context, v any, _ string, _ ...any) error {
	target := reflect.ValueOf(v)
	if target.Kind() != reflect.Pointer || target.IsNil() {
		return nil
	}
	elem := target.Elem()
	if elem.Kind() == reflect.Slice {
		elem.Set(reflect.ValueOf(db.items))
	}
	return nil
}

func TestListDailyTasksFiltersExistingTasksOnPausedScheduleDays(t *testing.T) {
	t.Parallel()

	pausedDate := time.Date(2026, 5, 11, 0, 0, 0, 0, time.Local)
	nextDate := time.Date(2026, 5, 12, 0, 0, 0, 0, time.Local)
	logic := NewListDailyTasksLogic(WithCurrentUserID(context.Background(), 7), &svc.ServiceContext{
		DB: &listDailyTasksDB{items: []model.DailyTasks{
			{
				Id:           1,
				UserId:       7,
				TaskDate:     pausedDate,
				AwarenessId:  sql.NullInt64{Int64: 101, Valid: true},
				TopicTitle:   "觉察边界",
				TopicSummary: "边界摘要",
				Status:       "draft",
				CreatedAt:    pausedDate,
				UpdatedAt:    pausedDate,
			},
			{
				Id:           2,
				UserId:       7,
				TaskDate:     nextDate,
				AwarenessId:  sql.NullInt64{Int64: 101, Valid: true},
				TopicTitle:   "觉察边界",
				TopicSummary: "边界摘要",
				Status:       "draft",
				CreatedAt:    nextDate,
				UpdatedAt:    nextDate,
			},
		}},
		DailyTasksModel: &listDailyTasksModel{},
		AwarenessScheduleDaysModel: &listAwarenessScheduleDaysModel{items: []model.AwarenessScheduleDays{
			{
				ScheduleDate: pausedDate,
				DayType:      scheduleDayPaused,
			},
			{
				ScheduleDate: nextDate,
				DayType:      scheduleDayNormal,
			},
		}},
	})

	resp, err := logic.ListDailyTasks(&types.DailyTaskQueryReq{StartDate: "2026-05-11", EndDate: "2026-05-12"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.List) != 1 {
		t.Fatalf("expected paused task to be hidden, got %+v", resp.Data.List)
	}
	if resp.Data.List[0].TaskDate != "2026-05-12" || resp.Data.List[0].Id != 2 {
		t.Fatalf("expected next normal task only, got %+v", resp.Data.List[0])
	}
}

func TestListDailyTasksUsesScheduleSnapshotForExistingHistoryTask(t *testing.T) {
	t.Parallel()

	taskDate := time.Date(2026, 5, 12, 0, 0, 0, 0, time.Local)
	logic := NewListDailyTasksLogic(WithCurrentUserID(context.Background(), 7), &svc.ServiceContext{
		DB: &listDailyTasksDB{items: []model.DailyTasks{
			{
				Id:           2,
				UserId:       7,
				TaskDate:     taskDate,
				AwarenessId:  sql.NullInt64{Int64: 101, Valid: true},
				TopicTitle:   "旧排期主题",
				TopicSummary: "旧摘要",
				Status:       "draft",
				CreatedAt:    taskDate,
				UpdatedAt:    taskDate,
			},
		}},
		DailyTasksModel: &listDailyTasksModel{},
		AwarenessScheduleDaysModel: &listAwarenessScheduleDaysModel{items: []model.AwarenessScheduleDays{
			{
				ScheduleDayId:     42,
				ScheduleDate:      taskDate,
				DayType:           scheduleDayNormal,
				AwarenessId:       sql.NullInt64{Int64: 202, Valid: true},
				AwarenessTitle:    sql.NullString{String: "新排期主题", Valid: true},
				AwarenessSummary:  sql.NullString{String: "新摘要", Valid: true},
				CycleDayIndex:     sql.NullInt64{Int64: 9, Valid: true},
				EffectiveDayIndex: sql.NullInt64{Int64: 19, Valid: true},
			},
		}},
	})

	resp, err := logic.ListDailyTasks(&types.DailyTaskQueryReq{StartDate: "2026-05-12", EndDate: "2026-05-12"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.List) != 1 {
		t.Fatalf("expected one history task, got %+v", resp.Data.List)
	}
	got := resp.Data.List[0]
	if got.TopicTitle != "新排期主题" || got.TopicSummary != "新摘要" || got.AwarenessId != 202 || got.TopicOrderNo != 9 {
		t.Fatalf("expected history task to use schedule snapshot, got %+v", got)
	}
}
