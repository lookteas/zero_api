package logic

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"testing"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type adminAwarenessMaintenanceModel struct {
	model.AwarenessModel
	points            []model.Awareness
	updatedID         uint64
	updatedTitle      string
	updatedSummary    string
	updatedDetails    string
	disabledID        uint64
	created           *model.Awareness
	reorderedID       uint64
	reorderedPosition int64
	nextID            uint64
}

func (m *adminAwarenessMaintenanceModel) FindEligible(context.Context) ([]model.Awareness, error) {
	result := make([]model.Awareness, 0, len(m.points))
	for _, point := range m.points {
		if point.Status == 1 && point.IsMeta == 0 {
			result = append(result, point)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].SortOrderGlobal == result[j].SortOrderGlobal {
			return result[i].AwarenessId < result[j].AwarenessId
		}
		return result[i].SortOrderGlobal < result[j].SortOrderGlobal
	})
	return result, nil
}

func (m *adminAwarenessMaintenanceModel) FindOne(_ context.Context, id uint64) (*model.Awareness, error) {
	for i := range m.points {
		if m.points[i].AwarenessId == id {
			return &m.points[i], nil
		}
	}
	return nil, model.ErrNotFound
}

func (m *adminAwarenessMaintenanceModel) CreateMinimal(_ context.Context, title, summary, details string) (*model.Awareness, error) {
	if m.nextID == 0 {
		m.nextID = 900
	}
	item := model.Awareness{
		AwarenessId:     m.nextID,
		PointTitle:      title,
		Summary:         sql.NullString{String: summary, Valid: summary != ""},
		Details:         sql.NullString{String: details, Valid: details != ""},
		Status:          1,
		IsMeta:          0,
		SortOrderGlobal: 999,
	}
	m.created = &item
	m.points = append(m.points, item)
	m.nextID++
	return &item, nil
}

func (m *adminAwarenessMaintenanceModel) MoveToPosition(_ context.Context, id uint64, position int64) error {
	m.reorderedID = id
	m.reorderedPosition = position

	eligible := make([]model.Awareness, 0, len(m.points))
	for _, point := range m.points {
		if point.Status == 1 && point.IsMeta == 0 {
			eligible = append(eligible, point)
		}
	}
	sort.Slice(eligible, func(i, j int) bool {
		if eligible[i].SortOrderGlobal == eligible[j].SortOrderGlobal {
			return eligible[i].AwarenessId < eligible[j].AwarenessId
		}
		return eligible[i].SortOrderGlobal < eligible[j].SortOrderGlobal
	})

	orderedIDs := make([]uint64, 0, len(eligible))
	for _, point := range eligible {
		if point.AwarenessId != id {
			orderedIDs = append(orderedIDs, point.AwarenessId)
		}
	}
	insertAt := int(position - 1)
	if insertAt < 0 {
		insertAt = 0
	}
	if insertAt > len(orderedIDs) {
		insertAt = len(orderedIDs)
	}
	orderedIDs = append(orderedIDs, 0)
	copy(orderedIDs[insertAt+1:], orderedIDs[insertAt:])
	orderedIDs[insertAt] = id

	for order, orderedID := range orderedIDs {
		for i := range m.points {
			if m.points[i].AwarenessId == orderedID {
				m.points[i].SortOrderGlobal = int64(order + 1)
			}
		}
	}
	return nil
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

func TestAdminInsertExistingAwarenessPlacesPointOnEffectiveDateAndReflows(t *testing.T) {
	t.Parallel()

	awareness := &adminAwarenessMaintenanceModel{points: []model.Awareness{
		{AwarenessId: 101, PointTitle: "周一", SortOrderGlobal: 1, Status: 1},
		{AwarenessId: 102, PointTitle: "周二", SortOrderGlobal: 2, Status: 1},
		{AwarenessId: 103, PointTitle: "原周三", SortOrderGlobal: 3, Status: 1},
		{AwarenessId: 104, PointTitle: "平常心", SortOrderGlobal: 99, Status: 1},
	}}
	schedule := &adminAwarenessScheduleModel{}
	logic := NewAdminInsertAwarenessLogic(WithCurrentAdminID(context.Background(), 1), &svc.ServiceContext{
		AwarenessModel:             awareness,
		AwarenessCyclesModel:       &adminAwarenessCycleModel{cycle: testMaintenanceCycle(t)},
		AwarenessScheduleDaysModel: schedule,
	})

	_, err := logic.AdminInsertAwareness(&types.AdminAwarenessInsertReq{
		ExistingAwarenessId: 104,
		EffectiveDate:       "2026-05-03",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if awareness.reorderedID != 104 || awareness.reorderedPosition != 3 {
		t.Fatalf("expected awareness 104 inserted at position 3, got id=%d position=%d", awareness.reorderedID, awareness.reorderedPosition)
	}
	if len(schedule.upserts) == 0 {
		t.Fatalf("expected schedule regeneration")
	}
	if schedule.upserts[0].AwarenessId.Int64 != 104 || schedule.upserts[0].AwarenessTitle.String != "平常心" {
		t.Fatalf("expected selected date to become 平常心, got %+v", schedule.upserts[0])
	}
	if len(schedule.upserts) > 1 && schedule.upserts[1].AwarenessId.Int64 != 103 {
		t.Fatalf("expected original third point to shift after inserted point, got %+v", schedule.upserts[1])
	}
}

func TestAdminInsertNewAwarenessCreatesPointAndPlacesItOnEffectiveDate(t *testing.T) {
	t.Parallel()

	awareness := &adminAwarenessMaintenanceModel{points: []model.Awareness{
		{AwarenessId: 101, PointTitle: "周一", SortOrderGlobal: 1, Status: 1},
		{AwarenessId: 102, PointTitle: "周二", SortOrderGlobal: 2, Status: 1},
		{AwarenessId: 103, PointTitle: "原周三", SortOrderGlobal: 3, Status: 1},
	}}
	schedule := &adminAwarenessScheduleModel{}
	logic := NewAdminInsertAwarenessLogic(WithCurrentAdminID(context.Background(), 1), &svc.ServiceContext{
		AwarenessModel:             awareness,
		AwarenessCyclesModel:       &adminAwarenessCycleModel{cycle: testMaintenanceCycle(t)},
		AwarenessScheduleDaysModel: schedule,
	})

	_, err := logic.AdminInsertAwareness(&types.AdminAwarenessInsertReq{
		Title:         "平常心",
		Summary:       "练习稳定地看见起伏",
		Description:   "把注意力放回当下。",
		EffectiveDate: "2026-05-03",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if awareness.created == nil || awareness.created.PointTitle != "平常心" {
		t.Fatalf("expected new awareness to be created, got %+v", awareness.created)
	}
	if awareness.reorderedID != awareness.created.AwarenessId || awareness.reorderedPosition != 3 {
		t.Fatalf("expected new awareness inserted at position 3, got id=%d position=%d", awareness.reorderedID, awareness.reorderedPosition)
	}
	if schedule.upserts[0].AwarenessTitle.String != "平常心" {
		t.Fatalf("expected selected date to use new point, got %+v", schedule.upserts[0])
	}
}

func TestAdminInsertExistingAwarenessFromEarlierPositionPlacesPointOnLaterDate(t *testing.T) {
	t.Parallel()

	awareness := &adminAwarenessMaintenanceModel{points: []model.Awareness{
		{AwarenessId: 101, PointTitle: "提前项", SortOrderGlobal: 1, Status: 1},
		{AwarenessId: 102, PointTitle: "第二项", SortOrderGlobal: 2, Status: 1},
		{AwarenessId: 103, PointTitle: "原目标日", SortOrderGlobal: 3, Status: 1},
		{AwarenessId: 104, PointTitle: "第四项", SortOrderGlobal: 4, Status: 1},
	}}
	schedule := &adminAwarenessScheduleModel{}
	logic := NewAdminInsertAwarenessLogic(WithCurrentAdminID(context.Background(), 1), &svc.ServiceContext{
		AwarenessModel:             awareness,
		AwarenessCyclesModel:       &adminAwarenessCycleModel{cycle: testMaintenanceCycle(t)},
		AwarenessScheduleDaysModel: schedule,
	})

	_, err := logic.AdminInsertAwareness(&types.AdminAwarenessInsertReq{
		ExistingAwarenessId: 101,
		EffectiveDate:       "2026-05-03",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schedule.upserts) < 2 {
		t.Fatalf("expected regenerated schedule, got %+v", schedule.upserts)
	}
	if schedule.upserts[0].AwarenessId.Int64 != 101 || schedule.upserts[0].AwarenessTitle.String != "提前项" {
		t.Fatalf("expected selected date to become 提前项, got %+v", schedule.upserts[0])
	}
	if schedule.upserts[1].AwarenessId.Int64 != 104 {
		t.Fatalf("expected point after inserted date to follow reordered sequence, got %+v", schedule.upserts[1])
	}
}

func TestAdminInsertDisabledExistingAwarenessReturnsErrorWithoutRegeneratingSchedule(t *testing.T) {
	t.Parallel()

	awareness := &adminAwarenessMaintenanceModel{points: []model.Awareness{
		{AwarenessId: 101, PointTitle: "周一", SortOrderGlobal: 1, Status: 1},
		{AwarenessId: 102, PointTitle: "停用项", SortOrderGlobal: 2, Status: 0},
		{AwarenessId: 103, PointTitle: "周三", SortOrderGlobal: 3, Status: 1},
	}}
	schedule := &adminAwarenessScheduleModel{}
	logic := NewAdminInsertAwarenessLogic(WithCurrentAdminID(context.Background(), 1), &svc.ServiceContext{
		AwarenessModel:             awareness,
		AwarenessCyclesModel:       &adminAwarenessCycleModel{cycle: testMaintenanceCycle(t)},
		AwarenessScheduleDaysModel: schedule,
	})

	_, err := logic.AdminInsertAwareness(&types.AdminAwarenessInsertReq{
		ExistingAwarenessId: 102,
		EffectiveDate:       "2026-05-02",
	})
	if err == nil {
		t.Fatalf("expected disabled awareness insert to fail")
	}
	if !strings.Contains(err.Error(), "disabled awareness cannot be inserted") {
		t.Fatalf("expected clear disabled awareness error, got %v", err)
	}
	if awareness.reorderedID != 0 {
		t.Fatalf("expected disabled awareness not to be reordered, got %d", awareness.reorderedID)
	}
	if len(schedule.upserts) != 0 {
		t.Fatalf("expected no schedule regeneration, got %+v", schedule.upserts)
	}
}

func TestAdminInsertNewAwarenessCalculatesPositionBeforeCreatingEligiblePoint(t *testing.T) {
	t.Parallel()

	awareness := &adminAwarenessMaintenanceModel{points: []model.Awareness{
		{AwarenessId: 101, PointTitle: "周一", SortOrderGlobal: 1, Status: 1},
		{AwarenessId: 102, PointTitle: "周二", SortOrderGlobal: 2, Status: 1},
		{AwarenessId: 103, PointTitle: "周三", SortOrderGlobal: 3, Status: 1},
	}}
	schedule := &adminAwarenessScheduleModel{}
	logic := NewAdminInsertAwarenessLogic(WithCurrentAdminID(context.Background(), 1), &svc.ServiceContext{
		AwarenessModel:             awareness,
		AwarenessCyclesModel:       &adminAwarenessCycleModel{cycle: testMaintenanceCycle(t)},
		AwarenessScheduleDaysModel: schedule,
	})

	_, err := logic.AdminInsertAwareness(&types.AdminAwarenessInsertReq{
		Title:         "新一轮首日",
		EffectiveDate: "2026-05-07",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if awareness.reorderedID != awareness.created.AwarenessId || awareness.reorderedPosition != 3 {
		t.Fatalf("expected new awareness inserted at original position 3, got id=%d position=%d", awareness.reorderedID, awareness.reorderedPosition)
	}
	if len(schedule.upserts) == 0 {
		t.Fatalf("expected schedule regeneration")
	}
	if schedule.upserts[0].AwarenessTitle.String != "新一轮首日" {
		t.Fatalf("expected selected date to use new point, got %+v", schedule.upserts[0])
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
