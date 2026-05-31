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

type adminAwarenessMaintenanceModel struct {
	model.AwarenessModel
	points         []model.Awareness
	updatedID      uint64
	updatedTitle   string
	updatedSummary string
	updatedDetails string
	disabledID     uint64
}

func (m *adminAwarenessMaintenanceModel) FindEligible(context.Context) ([]model.Awareness, error) {
	result := make([]model.Awareness, 0, len(m.points))
	for _, point := range m.points {
		if point.Status == 1 && point.IsMeta == 0 {
			result = append(result, point)
		}
	}
	return result, nil
}

func (m *adminAwarenessMaintenanceModel) UpdateContent(_ context.Context, id uint64, title, summary, details string) error {
	m.updatedID = id
	m.updatedTitle = title
	m.updatedSummary = summary
	m.updatedDetails = details
	for i := range m.points {
		if m.points[i].AwarenessId == id {
			m.points[i].PointTitle = title
			m.points[i].Summary = sql.NullString{String: summary, Valid: summary != ""}
			m.points[i].Details = sql.NullString{String: details, Valid: details != ""}
		}
	}
	return nil
}

func (m *adminAwarenessMaintenanceModel) Disable(_ context.Context, id uint64) error {
	m.disabledID = id
	for i := range m.points {
		if m.points[i].AwarenessId == id {
			m.points[i].Status = 0
		}
	}
	return nil
}

func (m *adminAwarenessMaintenanceModel) withSession(sqlx.Session) model.AwarenessModel { return m }

type adminAwarenessCycleModel struct {
	model.AwarenessCyclesModel
	cycle *model.AwarenessCycles
}

func (m *adminAwarenessCycleModel) FindActiveByCommunity(context.Context, uint64) (*model.AwarenessCycles, error) {
	return m.cycle, nil
}

func (m *adminAwarenessCycleModel) Update(context.Context, *model.AwarenessCycles) error { return nil }
func (m *adminAwarenessCycleModel) withSession(sqlx.Session) model.AwarenessCyclesModel  { return m }

type adminAwarenessScheduleModel struct {
	model.AwarenessScheduleDaysModel
	upserts []model.AwarenessScheduleDays
}

func (m *adminAwarenessScheduleModel) Upsert(_ context.Context, item *model.AwarenessScheduleDays) error {
	m.upserts = append(m.upserts, *item)
	return nil
}

func (m *adminAwarenessScheduleModel) FindOneByCycleIdScheduleDate(context.Context, uint64, time.Time) (*model.AwarenessScheduleDays, error) {
	return nil, model.ErrNotFound
}

func (m *adminAwarenessScheduleModel) FindByCommunityDateRange(context.Context, uint64, time.Time, time.Time) ([]model.AwarenessScheduleDays, error) {
	return nil, nil
}

func (m *adminAwarenessScheduleModel) Insert(context.Context, *model.AwarenessScheduleDays) (sql.Result, error) {
	return nil, nil
}

func (m *adminAwarenessScheduleModel) DeleteFutureByCycle(context.Context, uint64, time.Time) error {
	return nil
}

func (m *adminAwarenessScheduleModel) withSession(sqlx.Session) model.AwarenessScheduleDaysModel {
	return m
}

func TestAdminUpdateAwarenessUpdatesSourceAndRegeneratesScheduleSnapshots(t *testing.T) {
	t.Parallel()

	awareness := &adminAwarenessMaintenanceModel{points: []model.Awareness{
		{AwarenessId: 101, PointTitle: "旧标题", Summary: sql.NullString{String: "旧摘要", Valid: true}, Details: sql.NullString{String: "旧详情", Valid: true}, SortOrderGlobal: 1, Status: 1},
		{AwarenessId: 102, PointTitle: "第二个", SortOrderGlobal: 2, Status: 1},
	}}
	schedule := &adminAwarenessScheduleModel{}
	logic := NewAdminUpdateAwarenessLogic(WithCurrentAdminID(context.Background(), 1), &svc.ServiceContext{
		AwarenessModel:             awareness,
		AwarenessCyclesModel:       &adminAwarenessCycleModel{cycle: testMaintenanceCycle(t)},
		AwarenessScheduleDaysModel: schedule,
	})

	_, err := logic.AdminUpdateAwareness(&types.AdminAwarenessUpdateReq{
		Id:            101,
		Title:         "新标题",
		Summary:       "新摘要",
		Description:   "新详情",
		EffectiveDate: "2026-05-01",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if awareness.updatedID != 101 || awareness.updatedTitle != "新标题" || awareness.updatedSummary != "新摘要" || awareness.updatedDetails != "新详情" {
		t.Fatalf("expected source content update, got id=%d title=%q summary=%q details=%q", awareness.updatedID, awareness.updatedTitle, awareness.updatedSummary, awareness.updatedDetails)
	}
	if len(schedule.upserts) == 0 {
		t.Fatalf("expected schedule regeneration")
	}
	if got, want := schedule.upserts[0].ScheduleDate.Format("2006-01-02"), "2026-05-01"; got != want {
		t.Fatalf("expected regeneration from %s, got %s", want, got)
	}
	if schedule.upserts[0].AwarenessTitle.String != "新标题" || schedule.upserts[0].AwarenessSummary.String != "新摘要" || schedule.upserts[0].AwarenessDetails.String != "新详情" {
		t.Fatalf("expected refreshed snapshot content, got %+v", schedule.upserts[0])
	}
}

func TestAdminExcludeAwarenessDisablesPointAndReflowsScheduleFromEffectiveDate(t *testing.T) {
	t.Parallel()

	awareness := &adminAwarenessMaintenanceModel{points: []model.Awareness{
		{AwarenessId: 101, PointTitle: "第一个", SortOrderGlobal: 1, Status: 1},
		{AwarenessId: 102, PointTitle: "重复项", SortOrderGlobal: 2, Status: 1},
		{AwarenessId: 103, PointTitle: "第三个", SortOrderGlobal: 3, Status: 1},
	}}
	schedule := &adminAwarenessScheduleModel{}
	logic := NewAdminExcludeAwarenessLogic(WithCurrentAdminID(context.Background(), 1), &svc.ServiceContext{
		AwarenessModel:             awareness,
		AwarenessCyclesModel:       &adminAwarenessCycleModel{cycle: testMaintenanceCycle(t)},
		AwarenessScheduleDaysModel: schedule,
	})

	_, err := logic.AdminExcludeAwareness(&types.AdminAwarenessExcludeReq{
		Id:            102,
		EffectiveDate: "2026-05-02",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if awareness.disabledID != 102 {
		t.Fatalf("expected awareness 102 to be disabled, got %d", awareness.disabledID)
	}
	if len(schedule.upserts) == 0 {
		t.Fatalf("expected schedule regeneration")
	}
	if got, want := schedule.upserts[0].ScheduleDate.Format("2006-01-02"), "2026-05-02"; got != want {
		t.Fatalf("expected regeneration from %s, got %s", want, got)
	}
	for _, item := range schedule.upserts {
		if item.AwarenessId.Valid && item.AwarenessId.Int64 == 102 {
			t.Fatalf("excluded awareness appeared in regenerated schedule: %+v", item)
		}
	}
	if !schedule.upserts[0].AwarenessId.Valid || schedule.upserts[0].AwarenessId.Int64 != 103 {
		t.Fatalf("expected 2026-05-02 to reflow to awareness 103, got %+v", schedule.upserts[0])
	}
}

func testMaintenanceCycle(t *testing.T) *model.AwarenessCycles {
	t.Helper()

	return &model.AwarenessCycles{
		CycleId:             1,
		CommunityId:         1,
		StartDate:           testAwarenessCycleDate(t, "2026-05-01"),
		RestDays:            1,
		ScheduleHorizonDays: 4,
		Status:              "active",
	}
}
