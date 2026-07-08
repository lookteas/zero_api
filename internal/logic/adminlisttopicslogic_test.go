package logic

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"

	"api/internal/config"
	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type adminListTopicsDB struct {
	sqlx.SqlConn
	query    string
	settings map[string]string
	execArgs []any
}

func (db *adminListTopicsDB) QueryRowsCtx(_ context.Context, _ any, query string, _ ...any) error {
	db.query = query
	return nil
}

func (db *adminListTopicsDB) QueryRowCtx(_ context.Context, v any, query string, args ...any) error {
	db.query = query
	if len(args) == 0 {
		return sql.ErrNoRows
	}
	value, ok := db.settings[args[0].(string)]
	if !ok {
		return sql.ErrNoRows
	}

	target := reflect.ValueOf(v)
	if target.Kind() != reflect.Pointer || target.IsNil() {
		return nil
	}
	elem := target.Elem()
	if elem.Kind() == reflect.String {
		elem.SetString(value)
	}
	return nil
}

func (db *adminListTopicsDB) ExecCtx(_ context.Context, query string, args ...any) (sql.Result, error) {
	db.query = query
	db.execArgs = args
	return nil, nil
}

type adminListAwarenessModel struct {
	model.AwarenessModel
	points        []model.Awareness
	called        bool
	eligibleCalls int
	findOneCalls  int
}

func (m *adminListAwarenessModel) FindEligible(context.Context) ([]model.Awareness, error) {
	m.called = true
	m.eligibleCalls++
	return m.points, nil
}
func (m *adminListAwarenessModel) FindOne(_ context.Context, id uint64) (*model.Awareness, error) {
	m.called = true
	m.findOneCalls++
	for i := range m.points {
		if m.points[i].AwarenessId == id {
			return &m.points[i], nil
		}
	}
	return nil, model.ErrNotFound
}
func (m *adminListAwarenessModel) withSession(sqlx.Session) model.AwarenessModel { return m }

type adminScheduleModel struct {
	model.AwarenessScheduleDaysModel
	items []model.AwarenessScheduleDays
}

func (m *adminScheduleModel) FindByCommunityDateRange(context.Context, uint64, time.Time, time.Time) ([]model.AwarenessScheduleDays, error) {
	return m.items, nil
}
func (m *adminScheduleModel) FindOneByCycleIdScheduleDate(context.Context, uint64, time.Time) (*model.AwarenessScheduleDays, error) {
	panic("unexpected call")
}
func (m *adminScheduleModel) Insert(context.Context, *model.AwarenessScheduleDays) (sql.Result, error) {
	panic("unexpected call")
}
func (m *adminScheduleModel) Upsert(context.Context, *model.AwarenessScheduleDays) error {
	panic("unexpected call")
}
func (m *adminScheduleModel) DeleteFutureByCycle(context.Context, uint64, time.Time) error {
	panic("unexpected call")
}
func (m *adminScheduleModel) withSession(sqlx.Session) model.AwarenessScheduleDaysModel {
	return m
}

func TestAdminListTopicsWithoutWeekStartQueriesLegacyTopics(t *testing.T) {
	t.Parallel()

	db := &adminListTopicsDB{}
	logic := NewAdminListTopicsLogic(WithCurrentAdminID(context.Background(), 1), &svc.ServiceContext{
		DB: db,
	})

	if _, err := logic.AdminListTopics(&types.TopicQueryReq{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(db.query, "from topics") {
		t.Fatalf("expected legacy topics query, got %q", db.query)
	}
}

func TestAdminListTopicsWithWeekStartReturnsAwarenessSchedule(t *testing.T) {
	t.Parallel()

	awarenessModel := &adminListAwarenessModel{
		points: []model.Awareness{
			{
				AwarenessId:     101,
				PointTitle:      "觉察身体紧绷",
				Theme:           sql.NullString{String: "身体", Valid: true},
				Summary:         sql.NullString{String: "看见紧绷", Valid: true},
				Details:         sql.NullString{String: "记录肩颈和呼吸", Valid: true},
				ReferenceMin:    sql.NullFloat64{Float64: 1.25, Valid: true},
				ReferenceMax:    sql.NullFloat64{Float64: 3.5, Valid: true},
				SortOrderGlobal: 1,
				Status:          1,
			},
			{
				AwarenessId:     102,
				PointTitle:      "觉察念头",
				SortOrderGlobal: 2,
				Status:          1,
			},
		},
	}
	logic := NewAdminListTopicsLogic(WithCurrentAdminID(context.Background(), 1), &svc.ServiceContext{
		Config: config.Config{
			AwarenessCycle: config.AwarenessCycleConf{
				StartDate: "2026-05-04",
				RestDays:  7,
			},
		},
		AwarenessModel: awarenessModel,
	})

	resp, err := logic.AdminListTopics(&types.TopicQueryReq{WeekStart: "2026-05-04"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !awarenessModel.called {
		t.Fatalf("expected awareness lookup")
	}
	if len(resp.Data.List) != 7 {
		t.Fatalf("expected 7 schedule items, got %d", len(resp.Data.List))
	}
	if resp.Data.Pagination.Total != 7 || resp.Data.Pagination.PageSize != 7 {
		t.Fatalf("expected schedule pagination total/pageSize 7, got %+v", resp.Data.Pagination)
	}

	first := resp.Data.List[0]
	if first.Id != 101 || first.AwarenessId != 101 || first.Title != "觉察身体紧绷" {
		t.Fatalf("expected first awareness item, got %+v", first)
	}
	if first.ScheduleDate != "2026-05-04" || first.IsRestDay {
		t.Fatalf("expected first date normal awareness day, got %+v", first)
	}
	if first.Summary != "看见紧绷" || first.Description != "记录肩颈和呼吸" || first.AwarenessTheme != "身体" {
		t.Fatalf("expected awareness details copied, got %+v", first)
	}
	if first.ReferenceMin != "1.25" || first.ReferenceMax != "3.50" {
		t.Fatalf("expected reference range, got min=%q max=%q", first.ReferenceMin, first.ReferenceMax)
	}

	second := resp.Data.List[1]
	if second.Id != 102 || second.ScheduleDate != "2026-05-05" || second.IsRestDay {
		t.Fatalf("expected second awareness item, got %+v", second)
	}

	rest := resp.Data.List[2]
	if !rest.IsRestDay || rest.Id != 0 || rest.ScheduleDate != "2026-05-06" {
		t.Fatalf("expected third item rest day, got %+v", rest)
	}
	if rest.Title != "本轮结束，休息整合中" {
		t.Fatalf("expected rest title, got %q", rest.Title)
	}
}

func TestAdminListTopicsWithStoredScheduleLoadsAwarenessByID(t *testing.T) {
	t.Parallel()

	weekStart := time.Date(2026, 5, 4, 0, 0, 0, 0, time.Local)
	scheduleItems := make([]model.AwarenessScheduleDays, 0, 7)
	for i := 0; i < 7; i++ {
		dayType := scheduleDayRest
		awarenessID := sql.NullInt64{}
		if i == 0 {
			dayType = scheduleDayNormal
			awarenessID = sql.NullInt64{Int64: 303, Valid: true}
		}
		scheduleItems = append(scheduleItems, model.AwarenessScheduleDays{
			ScheduleDate:      weekStart.AddDate(0, 0, i),
			DayType:           dayType,
			AwarenessId:       awarenessID,
			CycleDayIndex:     sql.NullInt64{Int64: int64(i), Valid: dayType == scheduleDayNormal},
			EffectiveDayIndex: sql.NullInt64{Int64: int64(i), Valid: dayType == scheduleDayNormal},
		})
	}

	awarenessModel := &adminListAwarenessModel{
		points: []model.Awareness{
			{
				AwarenessId: 303,
				PointTitle:  "已停用但保留排程",
				Summary:     sql.NullString{String: "历史仍可显示", Valid: true},
				Status:      0,
			},
		},
	}
	logic := NewAdminListTopicsLogic(WithCurrentAdminID(context.Background(), 1), &svc.ServiceContext{
		AwarenessModel:             awarenessModel,
		AwarenessScheduleDaysModel: &adminScheduleModel{items: scheduleItems},
	})

	resp, err := logic.AdminListTopics(&types.TopicQueryReq{WeekStart: "2026-05-04"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if awarenessModel.eligibleCalls != 0 || awarenessModel.findOneCalls != 1 {
		t.Fatalf("expected stored schedule to load by id only, eligible=%d findOne=%d", awarenessModel.eligibleCalls, awarenessModel.findOneCalls)
	}
	if len(resp.Data.List) != 7 {
		t.Fatalf("expected stored week schedule, got %+v", resp.Data.List)
	}
	first := resp.Data.List[0]
	if first.AwarenessId != 303 || first.Title != "已停用但保留排程" || first.Summary != "历史仍可显示" {
		t.Fatalf("expected stored schedule awareness content, got %+v", first)
	}
}

func TestAdminListTopicsWithWeekStartUsesStoredAwarenessCycleSettings(t *testing.T) {
	t.Parallel()

	awarenessModel := &adminListAwarenessModel{
		points: []model.Awareness{
			{AwarenessId: 101, PointTitle: "第一天", SortOrderGlobal: 1, Status: 1},
			{AwarenessId: 102, PointTitle: "第二天", SortOrderGlobal: 2, Status: 1},
		},
	}
	logic := NewAdminListTopicsLogic(WithCurrentAdminID(context.Background(), 1), &svc.ServiceContext{
		Config: config.Config{
			AwarenessCycle: config.AwarenessCycleConf{
				StartDate: "2026-05-04",
				RestDays:  7,
			},
		},
		DB: &adminListTopicsDB{settings: map[string]string{
			awarenessCycleStartDateKey: "2026-05-05",
			awarenessCycleRestDaysKey:  "1",
		}},
		AwarenessModel: awarenessModel,
	})

	resp, err := logic.AdminListTopics(&types.TopicQueryReq{WeekStart: "2026-05-04"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first := resp.Data.List[0]; !first.IsRestDay || first.ScheduleDate != "2026-05-04" {
		t.Fatalf("expected pre-start rest item on 2026-05-04, got %+v", first)
	}
	if second := resp.Data.List[1]; second.IsRestDay || second.AwarenessId != 101 || second.ScheduleDate != "2026-05-05" {
		t.Fatalf("expected stored start date to make 2026-05-05 first awareness day, got %+v", second)
	}
	if fourth := resp.Data.List[3]; !fourth.IsRestDay || fourth.ScheduleDate != "2026-05-07" {
		t.Fatalf("expected stored restDays=1 to rest after two points, got %+v", fourth)
	}
	if fifth := resp.Data.List[4]; fifth.IsRestDay || fifth.AwarenessId != 101 || fifth.ScheduleDate != "2026-05-08" {
		t.Fatalf("expected cycle restart after stored one rest day, got %+v", fifth)
	}
}

func TestAdminListTopicsWithInvalidWeekStartReturnsError(t *testing.T) {
	t.Parallel()

	logic := NewAdminListTopicsLogic(WithCurrentAdminID(context.Background(), 1), &svc.ServiceContext{})

	_, err := logic.AdminListTopics(&types.TopicQueryReq{WeekStart: "bad-date"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "invalid weekStart") {
		t.Fatalf("expected clear weekStart error, got %v", err)
	}
}
