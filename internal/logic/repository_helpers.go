package logic

import (
	"database/sql"
	"fmt"
	"time"

	"api/internal/types"
	"api/model"
)

func hasDatabase(topicModel model.TopicsModel, dailyTaskModel model.DailyTasksModel) bool {
	return topicModel != nil || dailyTaskModel != nil
}

func topicToInfo(item *model.Topics) types.TopicInfo {
	description := ""
	if item.Description.Valid {
		description = item.Description.String
	}

	scheduleDate := ""
	if item.ScheduleDate.Valid {
		scheduleDate = item.ScheduleDate.Time.Format("2006-01-02")
	}

	return types.TopicInfo{
		Id:           item.Id,
		Title:        item.Title,
		Summary:      item.Summary,
		Description:  description,
		OrderNo:      item.OrderNo,
		Status:       int64(item.Status),
		ScheduleDate: scheduleDate,
	}
}

func nullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}

	return sql.NullString{String: value, Valid: true}
}

func parseTopicScheduleDate(input string) (sql.NullTime, error) {
	value := input
	if value == "" {
		return sql.NullTime{}, nil
	}

	parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return sql.NullTime{}, fmt.Errorf("scheduleDate must be in YYYY-MM-DD format")
	}

	return sql.NullTime{Time: normalizeDate(parsed), Valid: true}, nil
}

type dailyTaskAccessInfo struct {
	CanEditContent      bool
	CanAppendReflection bool
}

func dailyTaskAccess(item *model.DailyTasks, now time.Time) dailyTaskAccessInfo {
	anchor := item.CreatedAt
	if !item.UpdatedAt.IsZero() {
		anchor = item.UpdatedAt
	}
	if item.SubmittedAt.Valid {
		anchor = item.SubmittedAt.Time
	}

	windowEndsAt := anchor.Add(24 * time.Hour)
	if !now.After(windowEndsAt) {
		return dailyTaskAccessInfo{
			CanEditContent: true,
		}
	}

	return dailyTaskAccessInfo{
		CanAppendReflection: true,
	}
}

func shouldRefreshTodayTaskTopic(item *model.DailyTasks, topic *model.Topics) bool {
	if item == nil || topic == nil {
		return false
	}
	if item.Status != "draft" {
		return false
	}
	if item.Weakness.Valid || item.ImprovementPlan.Valid || item.VerificationPath.Valid || item.ReflectionNote.Valid {
		return false
	}
	if item.TopicId == topic.Id && item.TopicOrderNo == topic.OrderNo && item.TopicTitle == topic.Title && item.TopicSummary == topic.Summary {
		return false
	}
	return true
}

func refreshTodayTaskTopicSnapshot(item *model.DailyTasks, topic *model.Topics) *model.DailyTasks {
	if item == nil || topic == nil {
		return item
	}
	cloned := *item
	cloned.TopicId = topic.Id
	cloned.TopicOrderNo = topic.OrderNo
	cloned.TopicTitle = topic.Title
	cloned.TopicSummary = topic.Summary
	cloned.UpdatedAt = time.Now()
	return &cloned
}

func topicDescription(topic *model.Topics) string {
	if topic == nil || !topic.Description.Valid {
		return ""
	}

	return topic.Description.String
}

func dailyTaskToInfo(item *model.DailyTasks) types.DailyTaskInfo {
	return dailyTaskToInfoWithTopic(item, nil)
}

func dailyTaskToInfoWithTopic(item *model.DailyTasks, topic *model.Topics) types.DailyTaskInfo {
	weakness := ""
	if item.Weakness.Valid {
		weakness = item.Weakness.String
	}

	improvementPlan := ""
	if item.ImprovementPlan.Valid {
		improvementPlan = item.ImprovementPlan.String
	}

	verificationPath := ""
	if item.VerificationPath.Valid {
		verificationPath = item.VerificationPath.String
	}

	reflectionNote := ""
	if item.ReflectionNote.Valid {
		reflectionNote = item.ReflectionNote.String
	}

	submittedAt := ""
	if item.SubmittedAt.Valid {
		submittedAt = item.SubmittedAt.Time.Format("2006-01-02 15:04:05")
	}

	access := dailyTaskAccess(item, time.Now())

	return types.DailyTaskInfo{
		Id:                  item.Id,
		TaskDate:            item.TaskDate.Format("2006-01-02"),
		TopicId:             item.TopicId,
		TopicOrderNo:        item.TopicOrderNo,
		TopicTitle:          item.TopicTitle,
		TopicSummary:        item.TopicSummary,
		TopicDescription:    topicDescription(topic),
		Weakness:            weakness,
		ImprovementPlan:     improvementPlan,
		VerificationPath:    verificationPath,
		ReflectionNote:      reflectionNote,
		Status:              item.Status,
		CanEditContent:      access.CanEditContent,
		CanAppendReflection: access.CanAppendReflection,
		SubmittedAt:         submittedAt,
		CreatedAt:           item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:           item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func parseTaskDate(input string) time.Time {
	if input == "" {
		return normalizeDate(time.Now())
	}

	parsed, err := time.ParseInLocation("2006-01-02", input, time.Local)
	if err != nil {
		return normalizeDate(time.Now())
	}

	return normalizeDate(parsed)
}

func normalizeDate(input time.Time) time.Time {
	return time.Date(input.Year(), input.Month(), input.Day(), 0, 0, 0, 0, input.Location())
}

func dailyLogToInfo(item *model.DailyLogs) types.DailyLogInfo {
	actionText := ""
	if item.ActionText.Valid {
		actionText = item.ActionText.String
	}

	remark := ""
	if item.Remark.Valid {
		remark = item.Remark.String
	}

	return types.DailyLogInfo{
		Id:          item.Id,
		DailyTaskId: item.DailyTaskId,
		LogTime:     item.LogTime.Format("2006-01-02 15:04:05"),
		ActionText:  actionText,
		Status:      item.Status,
		Remark:      remark,
		CreatedAt:   item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func reviewRecordToInfo(item *model.ReviewRecords) types.ReviewRecordInfo {
	actualSituation := ""
	if item.ActualSituation.Valid {
		actualSituation = item.ActualSituation.String
	}

	suggestion := ""
	if item.Suggestion.Valid {
		suggestion = item.Suggestion.String
	}

	return types.ReviewRecordInfo{
		Id:              item.Id,
		ReviewItemId:    item.ReviewItemId,
		Result:          item.Result,
		ActualSituation: actualSituation,
		Suggestion:      suggestion,
		SubmittedAt:     item.SubmittedAt.Format("2006-01-02 15:04:05"),
	}
}

func reviewItemToInfo(item *model.ReviewItems, task *types.DailyTaskInfo, record *types.ReviewRecordInfo) types.ReviewItemInfo {
	completedAt := ""
	if item.CompletedAt.Valid {
		completedAt = item.CompletedAt.Time.Format("2006-01-02 15:04:05")
	}

	info := types.ReviewItemInfo{
		Id:          item.Id,
		DailyTaskId: item.DailyTaskId,
		ReviewStage: item.ReviewStage,
		DueAt:       item.DueAt.Format("2006-01-02 15:04:05"),
		Status:      item.Status,
		CompletedAt: completedAt,
	}

	if task != nil {
		info.DailyTask = *task
	}

	if record != nil {
		info.LatestRecord = *record
	}

	return info
}
