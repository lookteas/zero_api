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

type recordingTopicsModel struct {
	model.TopicsModel
	seenDate time.Time
	item     *model.Topics
	err      error
}

func (m *recordingTopicsModel) Insert(context.Context, *model.Topics) (sql.Result, error) {
	panic("unexpected call")
}
func (m *recordingTopicsModel) FindOne(context.Context, uint64) (*model.Topics, error) {
	panic("unexpected call")
}
func (m *recordingTopicsModel) FindOneByOrderNo(context.Context, int64) (*model.Topics, error) {
	panic("unexpected legacy call")
}
func (m *recordingTopicsModel) FindLatestActiveByScheduleDate(_ context.Context, scheduleDate time.Time) (*model.Topics, error) {
	m.seenDate = scheduleDate
	return m.item, m.err
}
func (m *recordingTopicsModel) Update(context.Context, *model.Topics) error { panic("unexpected call") }
func (m *recordingTopicsModel) Delete(context.Context, uint64) error        { panic("unexpected call") }
func (m *recordingTopicsModel) withSession(sqlx.Session) model.TopicsModel  { return m }

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

func TestCreateDailyTaskUsesScheduledTopicForTaskDate(t *testing.T) {
	t.Parallel()

	taskDate := time.Date(2026, 4, 19, 0, 0, 0, 0, time.Local)
	description := "今天围绕这一点练，把情境、动作和判断标准都看清楚。"
	topic := &model.Topics{
		Id:          19,
		Title:       "scheduled topic",
		Summary:     "scheduled summary",
		Description: sql.NullString{String: description, Valid: true},
		OrderNo:     19,
		Status:      1,
		CreatedAt:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.Local),
		UpdatedAt:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.Local),
	}
	dailyTasks := &createDailyTasksModel{
		findErr: model.ErrNotFound,
		storedItem: &model.DailyTasks{
			Id:           9,
			UserId:       7,
			TaskDate:     taskDate,
			TopicId:      topic.Id,
			TopicOrderNo: topic.OrderNo,
			TopicTitle:   topic.Title,
			TopicSummary: topic.Summary,
			Status:       "draft",
			CreatedAt:    taskDate,
			UpdatedAt:    taskDate,
		},
	}
	topics := &recordingTopicsModel{item: topic}

	ctx := WithCurrentUserID(context.Background(), 7)
	logic := NewCreateDailyTaskLogic(ctx, &svc.ServiceContext{
		TopicsModel:     topics,
		DailyTasksModel: dailyTasks,
	})

	resp, err := logic.CreateDailyTask(&types.DailyTaskCreateReq{TaskDate: "2026-04-19"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !topics.seenDate.Equal(taskDate) {
		t.Fatalf("expected schedule date lookup %v, got %v", taskDate, topics.seenDate)
	}
	if dailyTasks.inserted == nil {
		t.Fatalf("expected daily task insert")
	}
	if dailyTasks.inserted.TopicId != topic.Id || dailyTasks.inserted.TopicOrderNo != topic.OrderNo {
		t.Fatalf("expected topic snapshot from scheduled topic, got topic_id=%d order_no=%d", dailyTasks.inserted.TopicId, dailyTasks.inserted.TopicOrderNo)
	}
	if resp.Data.TopicTitle != topic.Title {
		t.Fatalf("expected response topic title %q, got %q", topic.Title, resp.Data.TopicTitle)
	}
	if resp.Data.TopicDescription != description {
		t.Fatalf("expected topic description %q, got %q", description, resp.Data.TopicDescription)
	}
}
