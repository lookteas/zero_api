package logic

import (
	"sort"
	"time"

	"api/model"
)

type reviewStagePlan struct {
	Name  string
	DueAt time.Time
}

func buildReviewStagePlans(base time.Time) []reviewStagePlan {
	days := []struct {
		name string
		days int
	}{
		{name: "day3", days: 3},
		{name: "day7", days: 7},
		{name: "day30", days: 30},
	}

	plans := make([]reviewStagePlan, 0, len(days))
	for _, item := range days {
		plans = append(plans, reviewStagePlan{
			Name:  item.name,
			DueAt: base.AddDate(0, 0, item.days),
		})
	}

	return plans
}

func isVisibleReviewStage(stage string) bool {
	switch stage {
	case "day3", "day7", "day30":
		return true
	default:
		return false
	}
}

func visibleDueReviewItems(items []model.ReviewItems, now time.Time, limit int) []model.ReviewItems {
	visible := make([]model.ReviewItems, 0, len(items))
	for _, item := range items {
		if item.Status != "pending" {
			continue
		}
		if !isVisibleReviewStage(item.ReviewStage) {
			continue
		}
		if item.DueAt.After(now) {
			continue
		}
		visible = append(visible, item)
	}

	sort.Slice(visible, func(i, j int) bool {
		return visible[i].DueAt.Before(visible[j].DueAt)
	})

	if limit > 0 && len(visible) > limit {
		return visible[:limit]
	}

	return visible
}
