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

	windowEndsAt := anchor.Add(48 * time.Hour)
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

func shouldRefreshTodayTaskAwareness(item *model.DailyTasks, awareness *model.Awareness) bool {
	if item == nil || awareness == nil {
		return false
	}
	if item.Status != "draft" {
		return false
	}
	if item.AwarenessId.Valid && uint64(item.AwarenessId.Int64) == awareness.AwarenessId {
		return false
	}
	return true
}

func refreshTodayTaskAwarenessSnapshot(item *model.DailyTasks, awareness *model.Awareness) *model.DailyTasks {
	if item == nil || awareness == nil {
		return item
	}
	cloned := *item
	cloned.TopicId = 0
	cloned.AwarenessId = sql.NullInt64{Int64: int64(awareness.AwarenessId), Valid: true}
	cloned.TopicOrderNo = awareness.SortOrderGlobal
	cloned.TopicTitle = awareness.PointTitle
	cloned.TopicSummary = awarenessSummary(awareness)
	cloned.UpdatedAt = time.Now()
	return &cloned
}

func shouldRefreshTodayTaskScheduleDay(item *model.DailyTasks, scheduleDay *model.AwarenessScheduleDays) bool {
	if item == nil || scheduleDay == nil || scheduleDay.DayType != scheduleDayNormal {
		return false
	}
	if item.Status != "draft" {
		return false
	}
	if item.ScheduleDayId.Valid && uint64(item.ScheduleDayId.Int64) == scheduleDay.ScheduleDayId {
		return false
	}
	if item.AwarenessId.Valid && scheduleDay.AwarenessId.Valid && item.AwarenessId.Int64 == scheduleDay.AwarenessId.Int64 {
		return false
	}
	return true
}

func refreshTodayTaskScheduleDaySnapshot(item *model.DailyTasks, scheduleDay *model.AwarenessScheduleDays) *model.DailyTasks {
	if item == nil || scheduleDay == nil {
		return item
	}
	cloned := *item
	cloned.CommunityId = scheduleDay.CommunityId
	cloned.ScheduleDayId = sql.NullInt64{Int64: int64(scheduleDay.ScheduleDayId), Valid: scheduleDay.ScheduleDayId > 0}
	cloned.TopicId = 0
	cloned.AwarenessId = sql.NullInt64{Int64: nullableInt64(scheduleDay.AwarenessId), Valid: scheduleDay.AwarenessId.Valid}
	cloned.TopicOrderNo = nullableInt64(scheduleDay.CycleDayIndex)
	cloned.TopicTitle = nullableString(scheduleDay.AwarenessTitle)
	cloned.TopicSummary = nullableString(scheduleDay.AwarenessSummary)
	cloned.UpdatedAt = time.Now()
	return &cloned
}

func topicDescription(topic *model.Topics) string {
	if topic == nil || !topic.Description.Valid {
		return ""
	}

	return topic.Description.String
}

func awarenessSummary(item *model.Awareness) string {
	if item == nil {
		return ""
	}
	if item.Summary.Valid {
		return item.Summary.String
	}
	if item.Theme.Valid {
		return item.Theme.String
	}
	return ""
}

func awarenessDetails(item *model.Awareness) string {
	if item == nil || !item.Details.Valid {
		return ""
	}
	return item.Details.String
}

func nullDecimalString(value sql.NullFloat64) string {
	if !value.Valid {
		return ""
	}
	return fmt.Sprintf("%.2f", value.Float64)
}

func applyAwarenessToDailyTaskInfo(info types.DailyTaskInfo, item *model.Awareness) types.DailyTaskInfo {
	if item == nil {
		return info
	}

	info.AwarenessId = item.AwarenessId
	info.AwarenessTitle = item.PointTitle
	if item.Theme.Valid {
		info.AwarenessTheme = item.Theme.String
	}
	info.AwarenessSummary = awarenessSummary(item)
	info.AwarenessDetails = awarenessDetails(item)
	info.ReferenceMin = nullDecimalString(item.ReferenceMin)
	info.ReferenceMax = nullDecimalString(item.ReferenceMax)
	info.BetterDirection = item.BetterDirection
	return info
}

func restDailyTaskInfo(taskDate time.Time) types.DailyTaskInfo {
	now := time.Now().Format("2006-01-02 15:04:05")
	return types.DailyTaskInfo{
		TaskDate:        normalizeDate(taskDate).Format("2006-01-02"),
		IsRestDay:       true,
		RestTitle:       "本轮结束，休息整合中",
		RestDescription: "今天不生成新的练习任务，可以回看历史打卡和到期复盘，把这一轮练过的意识点整合一下。",
		Status:          "rest",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func pausedDailyTaskInfo(taskDate time.Time, reason string) types.DailyTaskInfo {
	title := "今日暂停打卡"
	description := "今天暂停生成新的练习任务，暂停结束后会继续原来的意识点，不会跳过进度。到期复盘仍然可以照常完成。"
	if reason != "" {
		title = "今日暂停打卡：" + reason
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	return types.DailyTaskInfo{
		TaskDate:        normalizeDate(taskDate).Format("2006-01-02"),
		IsRestDay:       true,
		IsPausedDay:     true,
		RestTitle:       title,
		RestDescription: description,
		Status:          "paused",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func scheduleDayToDailyTaskInfo(item *model.AwarenessScheduleDays) types.DailyTaskInfo {
	if item == nil {
		return types.DailyTaskInfo{}
	}
	switch item.DayType {
	case scheduleDayPaused:
		return pausedDailyTaskInfo(item.ScheduleDate, nullableString(item.PauseReason))
	case scheduleDayRest:
		return restDailyTaskInfo(item.ScheduleDate)
	}

	info := types.DailyTaskInfo{
		TaskDate:         item.ScheduleDate.Format("2006-01-02"),
		TopicId:          0,
		TopicOrderNo:     nullableInt64(item.CycleDayIndex),
		TopicTitle:       nullableString(item.AwarenessTitle),
		TopicSummary:     nullableString(item.AwarenessSummary),
		AwarenessId:      uint64(nullableInt64(item.AwarenessId)),
		AwarenessTitle:   nullableString(item.AwarenessTitle),
		AwarenessTheme:   nullableString(item.AwarenessTheme),
		AwarenessSummary: nullableString(item.AwarenessSummary),
		AwarenessDetails: nullableString(item.AwarenessDetails),
		ReferenceMin:     nullDecimalString(item.ReferenceMin),
		ReferenceMax:     nullDecimalString(item.ReferenceMax),
		BetterDirection:  nullableString(item.BetterDirection),
		Status:           "draft",
		CanEditContent:   true,
		CreatedAt:        item.UpdatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	return info
}

func applyScheduleDayAwarenessToDailyTaskInfo(info types.DailyTaskInfo, item *model.AwarenessScheduleDays) types.DailyTaskInfo {
	if item == nil || item.DayType != scheduleDayNormal {
		return info
	}
	info.AwarenessId = uint64(nullableInt64(item.AwarenessId))
	info.AwarenessTitle = nullableString(item.AwarenessTitle)
	info.AwarenessTheme = nullableString(item.AwarenessTheme)
	info.AwarenessSummary = nullableString(item.AwarenessSummary)
	info.AwarenessDetails = nullableString(item.AwarenessDetails)
	info.ReferenceMin = nullDecimalString(item.ReferenceMin)
	info.ReferenceMax = nullDecimalString(item.ReferenceMax)
	info.BetterDirection = nullableString(item.BetterDirection)
	return info
}

func scheduleDayToTopicInfo(item *model.AwarenessScheduleDays) types.TopicInfo {
	if item == nil {
		return types.TopicInfo{}
	}
	if item.DayType == scheduleDayPaused {
		return types.TopicInfo{
			Id:           0,
			Title:        "暂停打卡",
			Summary:      "今天暂停生成新的练习任务，暂停结束后继续原来的意识点。",
			OrderNo:      0,
			Status:       1,
			ScheduleDate: item.ScheduleDate.Format("2006-01-02"),
			IsRestDay:    true,
		}
	}
	if item.DayType == scheduleDayRest {
		return restTopicInfo(item.ScheduleDate)
	}
	return types.TopicInfo{
		Id:             uint64(nullableInt64(item.AwarenessId)),
		Title:          nullableString(item.AwarenessTitle),
		Summary:        nullableString(item.AwarenessSummary),
		Description:    nullableString(item.AwarenessDetails),
		OrderNo:        nullableInt64(item.CycleDayIndex),
		Status:         1,
		ScheduleDate:   item.ScheduleDate.Format("2006-01-02"),
		AwarenessId:    uint64(nullableInt64(item.AwarenessId)),
		AwarenessTheme: nullableString(item.AwarenessTheme),
		ReferenceMin:   nullDecimalString(item.ReferenceMin),
		ReferenceMax:   nullDecimalString(item.ReferenceMax),
		ProgressNo:     nullableInt64(item.EffectiveDayIndex) + 1,
	}
}

func nullableString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func nullableInt64(value sql.NullInt64) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
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

func findAwarenessByID(points []model.Awareness, awarenessID uint64) *model.Awareness {
	for i := range points {
		if points[i].AwarenessId == awarenessID {
			return &points[i]
		}
	}
	return nil
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
