package logic

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"api/internal/config"
	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type adminListTopicsDB struct {
	sqlx.SqlConn
	query string
}

func (db *adminListTopicsDB) QueryRowsCtx(_ context.Context, _ any, query string, _ ...any) error {
	db.query = query
	return nil
}

type adminListAwarenessModel struct {
	model.AwarenessModel
	points []model.Awareness
	called bool
}

func (m *adminListAwarenessModel) FindEligible(context.Context) ([]model.Awareness, error) {
	m.called = true
	return m.points, nil
}
func (m *adminListAwarenessModel) withSession(sqlx.Session) model.AwarenessModel { return m }

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
