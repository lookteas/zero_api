package logic

import (
	"context"
	"database/sql"
	"testing"

	"api/internal/config"
	"api/internal/svc"
	"api/internal/types"
	"api/model"
)

func TestAdminGetAwarenessCycleReturnsSummaryWithConfigDefaults(t *testing.T) {
	t.Parallel()

	awarenessModel := &adminListAwarenessModel{
		points: []model.Awareness{
			{
				AwarenessId:     201,
				PointTitle:      "觉察呼吸",
				Theme:           sql.NullString{String: "身体", Valid: true},
				Summary:         sql.NullString{String: "留意呼吸", Valid: true},
				SortOrderGlobal: 1,
				Status:          1,
			},
			{
				AwarenessId:     202,
				PointTitle:      "觉察念头",
				Summary:         sql.NullString{String: "看见念头", Valid: true},
				SortOrderGlobal: 2,
				Status:          1,
			},
		},
	}
	logic := NewAdminGetAwarenessCycleLogic(WithCurrentAdminID(context.Background(), 1), &svc.ServiceContext{
		Config: config.Config{
			AwarenessCycle: config.AwarenessCycleConf{
				StartDate: "2026-05-04",
				RestDays:  7,
			},
		},
		AwarenessModel: awarenessModel,
	})

	resp, err := logic.AdminGetAwarenessCycle()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Code != 0 || resp.Message != "ok" {
		t.Fatalf("expected ok response, got %+v", resp)
	}
	if resp.Data.StartDate != "2026-05-04" || resp.Data.RestDays != 7 {
		t.Fatalf("expected config cycle settings, got %+v", resp.Data)
	}
	if resp.Data.EligibleAwarenessCount != 2 {
		t.Fatalf("expected eligible count 2, got %d", resp.Data.EligibleAwarenessCount)
	}
	if resp.Data.NormalDayCount != 2 || resp.Data.RestDayCount != 5 {
		t.Fatalf("expected 2 normal and 5 rest days, got normal=%d rest=%d", resp.Data.NormalDayCount, resp.Data.RestDayCount)
	}
	if len(resp.Data.WeekDays) != 7 {
		t.Fatalf("expected 7 week days, got %d", len(resp.Data.WeekDays))
	}
	first := resp.Data.WeekDays[0]
	if first.Date != "2026-05-04" || first.AwarenessId != 201 || first.Title != "觉察呼吸" || first.IsRestDay {
		t.Fatalf("expected first awareness day, got %+v", first)
	}
	third := resp.Data.WeekDays[2]
	if third.Date != "2026-05-06" || !third.IsRestDay {
		t.Fatalf("expected third day rest, got %+v", third)
	}
}

func TestAdminUpdateAwarenessCycleRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	logic := NewAdminUpdateAwarenessCycleLogic(WithCurrentAdminID(context.Background(), 1), &svc.ServiceContext{})

	if _, err := logic.AdminUpdateAwarenessCycle(&types.AwarenessCycleUpdateReq{StartDate: "2026/05/04", RestDays: 7}); err == nil {
		t.Fatalf("expected invalid date error")
	}
	if _, err := logic.AdminUpdateAwarenessCycle(&types.AwarenessCycleUpdateReq{StartDate: "2026-05-04", RestDays: 0}); err == nil {
		t.Fatalf("expected invalid restDays error")
	}
}
