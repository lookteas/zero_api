package logic

import (
	"math"
	"time"

	"api/internal/types"
)

const defaultCycleTotalPoints int64 = 200

func BuildCycleSummary(totalPoints, completedTaskCount int64, completedAt time.Time) *types.CycleSummaryInfo {
	if totalPoints <= 0 {
		totalPoints = defaultCycleTotalPoints
	}
	if completedTaskCount < totalPoints || completedAt.IsZero() {
		return nil
	}

	restDays := calculateCycleRestDays(totalPoints)
	nextCycleStart := normalizeDate(completedAt).AddDate(0, 0, int(restDays)+1)
	return &types.CycleSummaryInfo{
		TotalPoints:        totalPoints,
		CompletedTaskCount: completedTaskCount,
		RestDays:           restDays,
		CycleCompletedDate: normalizeDate(completedAt).Format("2006-01-02"),
		NextCycleStartDate: nextCycleStart.Format("2006-01-02"),
	}
}

func calculateCycleRestDays(totalPoints int64) int64 {
	if totalPoints <= 0 {
		totalPoints = defaultCycleTotalPoints
	}
	restDays := int64(math.Round(float64(totalPoints) * 0.035))
	if restDays < 5 {
		return 5
	}
	if restDays > 10 {
		return 10
	}
	return restDays
}
