package logic

import (
	"context"
	"sort"
	"strings"
	"time"

	"api/internal/svc"
	"api/internal/types"
	"api/model"
)

const reviewRecoveryThresholdDays = 45

type ReviewPresentation struct {
	RecoveryMode        bool
	VisibleItems        []model.ReviewItems
	RecoveryGroups      []types.ReviewRecoveryGroupInfo
	VisibleItemTotal    int
	RecoveryGroupTotal  int
}

func BuildReviewItemListData(items []model.ReviewItems, tasks map[uint64]types.DailyTaskInfo, records map[uint64]types.ReviewRecordInfo, now, lastActiveAt time.Time) types.ReviewItemListData {
	presentation := BuildReviewPresentation(items, now, lastActiveAt, 1)
	list := make([]types.ReviewItemInfo, 0, len(presentation.VisibleItems))
	for _, item := range presentation.VisibleItems {
		list = append(list, reviewItemInfoFromMaps(item, tasks, records))
	}

	groups := attachRecoveryTaskInfo(presentation.RecoveryGroups, tasks)
	pendingRemainingCount := presentation.VisibleItemTotal - len(list)
	if pendingRemainingCount < 0 {
		pendingRemainingCount = 0
	}
	recoveryRemainingCount := presentation.RecoveryGroupTotal - len(groups)
	if recoveryRemainingCount < 0 {
		recoveryRemainingCount = 0
	}

	return types.ReviewItemListData{
		List:                   list,
		RecoveryGroups:         groups,
		PendingRemainingCount:  int64(pendingRemainingCount),
		RecoveryRemainingCount: int64(recoveryRemainingCount),
		Pagination: types.Pagination{
			Page:     1,
			PageSize: int64(len(list) + len(groups)),
			Total:    int64(len(list) + len(groups)),
		},
	}
}

func BuildHomeReviewPresentation(items []model.ReviewItems, tasks map[uint64]types.DailyTaskInfo, now, lastActiveAt time.Time) ([]types.ReviewItemInfo, []types.ReviewRecoveryGroupInfo) {
	presentation := BuildReviewPresentation(items, now, lastActiveAt, 1)
	pending := make([]types.ReviewItemInfo, 0, len(presentation.VisibleItems))
	for _, item := range presentation.VisibleItems {
		pending = append(pending, reviewItemInfoFromMaps(item, tasks, nil))
	}
	return pending, attachRecoveryTaskInfo(presentation.RecoveryGroups, tasks)
}

func BuildReviewPresentation(items []model.ReviewItems, now, lastActiveAt time.Time, limit int) ReviewPresentation {
	eligible := filterPresentableReviewItems(items)
	if !isReviewRecoveryMode(now, lastActiveAt) {
		allVisible := visibleDueReviewItems(eligible, now, 0)
		visible := allVisible
		if limit > 0 && len(visible) > limit {
			visible = visible[:limit]
		}
		return ReviewPresentation{
			RecoveryMode:     false,
			VisibleItems:     visible,
			VisibleItemTotal: len(allVisible),
		}
	}

	overdue := make([]model.ReviewItems, 0, len(eligible))
	for _, item := range eligible {
		if item.DueAt.After(now) {
			continue
		}
		overdue = append(overdue, item)
	}

	sortReviewItemsByDueAt(overdue)

	groups := buildReviewRecoveryGroups(overdue)
	totalGroups := len(groups)
	if totalGroups > 0 {
		if limit > 0 && len(groups) > limit {
			groups = groups[:limit]
		}
		return ReviewPresentation{
			RecoveryMode:       true,
			RecoveryGroups:     groups,
			RecoveryGroupTotal: totalGroups,
		}
	}

	return ReviewPresentation{
		RecoveryMode: true,
	}
}

func isReviewRecoveryMode(now, lastActiveAt time.Time) bool {
	if lastActiveAt.IsZero() {
		return false
	}

	return lastActiveAt.Before(now.AddDate(0, 0, -reviewRecoveryThresholdDays))
}

func filterPresentableReviewItems(items []model.ReviewItems) []model.ReviewItems {
	filtered := make([]model.ReviewItems, 0, len(items))
	for _, item := range items {
		if item.Status != "pending" {
			continue
		}
		if !isVisibleReviewStage(item.ReviewStage) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func sortReviewItemsByDueAt(items []model.ReviewItems) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].DueAt.Equal(items[j].DueAt) {
			return items[i].Id < items[j].Id
		}
		return items[i].DueAt.Before(items[j].DueAt)
	})
}

func buildReviewRecoveryGroups(items []model.ReviewItems) []types.ReviewRecoveryGroupInfo {
	grouped := make(map[uint64][]model.ReviewItems)
	order := make([]uint64, 0)
	for _, item := range items {
		if _, ok := grouped[item.DailyTaskId]; !ok {
			order = append(order, item.DailyTaskId)
		}
		grouped[item.DailyTaskId] = append(grouped[item.DailyTaskId], item)
	}

	groups := make([]types.ReviewRecoveryGroupInfo, 0, len(grouped))
	for _, taskID := range order {
		mergedItems := grouped[taskID]
		sortReviewItemsByDueAt(mergedItems)
		stages := uniqueSortedStages(mergedItems)
		itemIDs := make([]uint64, 0, len(mergedItems))
		for _, item := range mergedItems {
			itemIDs = append(itemIDs, item.Id)
		}
		groups = append(groups, types.ReviewRecoveryGroupInfo{
			DailyTaskId:      taskID,
			MergedStageLabel: strings.Join(stages, " / "),
			ReviewStageNames: stages,
			ReviewItemIds:    itemIDs,
			OldestDueAt:      mergedItems[0].DueAt.Format("2006-01-02 15:04:05"),
			LatestDueAt:      mergedItems[len(mergedItems)-1].DueAt.Format("2006-01-02 15:04:05"),
		})
	}
	return groups
}

func uniqueSortedStages(items []model.ReviewItems) []string {
	set := make(map[string]struct{})
	for _, item := range items {
		set[item.ReviewStage] = struct{}{}
	}
	stages := make([]string, 0, len(set))
	for stage := range set {
		stages = append(stages, stage)
	}
	sort.Slice(stages, func(i, j int) bool {
		return reviewStageSortKey(stages[i]) < reviewStageSortKey(stages[j])
	})
	return stages
}

func reviewStageSortKey(stage string) int {
	switch stage {
	case "day3":
		return 3
	case "day7":
		return 7
	case "day30":
		return 30
	default:
		return 999
	}
}

func reviewItemInfoFromMaps(item model.ReviewItems, tasks map[uint64]types.DailyTaskInfo, records map[uint64]types.ReviewRecordInfo) types.ReviewItemInfo {
	info := types.ReviewItemInfo{
		Id:          item.Id,
		DailyTaskId: item.DailyTaskId,
		ReviewStage: item.ReviewStage,
		DueAt:       item.DueAt.Format("2006-01-02 15:04:05"),
		Status:      item.Status,
	}
	if item.CompletedAt.Valid {
		info.CompletedAt = item.CompletedAt.Time.Format("2006-01-02 15:04:05")
	}
	if task, ok := tasks[item.DailyTaskId]; ok {
		info.DailyTask = task
	}
	if records != nil {
		if record, ok := records[item.Id]; ok {
			info.LatestRecord = record
		}
	}
	return info
}

func attachRecoveryTaskInfo(groups []types.ReviewRecoveryGroupInfo, tasks map[uint64]types.DailyTaskInfo) []types.ReviewRecoveryGroupInfo {
	attached := make([]types.ReviewRecoveryGroupInfo, 0, len(groups))
	for _, group := range groups {
		if task, ok := tasks[group.DailyTaskId]; ok {
			group.DailyTask = task
		}
		attached = append(attached, group)
	}
	return attached
}

func loadReviewTaskInfoMap(ctx context.Context, svcCtx *svc.ServiceContext, items []model.ReviewItems) map[uint64]types.DailyTaskInfo {
	tasks := make(map[uint64]types.DailyTaskInfo)
	for _, item := range items {
		if _, ok := tasks[item.DailyTaskId]; ok {
			continue
		}
		task, err := svcCtx.DailyTasksModel.FindOne(ctx, item.DailyTaskId)
		if err != nil || task == nil {
			continue
		}
		tasks[item.DailyTaskId] = dailyTaskToInfo(task)
	}
	return tasks
}

func loadReviewRecordInfoMap(ctx context.Context, svcCtx *svc.ServiceContext, items []model.ReviewItems) map[uint64]types.ReviewRecordInfo {
	records := make(map[uint64]types.ReviewRecordInfo)
	for _, item := range items {
		record, err := svcCtx.ReviewRecordsModel.FindOneByReviewItemId(ctx, item.Id)
		if err != nil || record == nil {
			continue
		}
		records[item.Id] = reviewRecordToInfo(record)
	}
	return records
}

func resolveReviewLastActiveAt(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64, fallback time.Time) time.Time {
	if svcCtx.UsersModel == nil {
		return fallback
	}
	user, err := svcCtx.UsersModel.FindOne(ctx, userID)
	if err != nil || user == nil || !user.LastLoginAt.Valid {
		return fallback
	}
	return user.LastLoginAt.Time
}
