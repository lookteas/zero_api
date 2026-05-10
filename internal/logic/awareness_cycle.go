package logic

import (
	"time"

	"api/model"
)

const defaultAwarenessCycleRestDays = 7

type awarenessCycleResult struct {
	Awareness    *model.Awareness
	IsRestDay    bool
	IsPreStart   bool
	RestDayIndex int
	CycleIndex   int
}

func resolveAwarenessCycleDay(startDate, targetDate time.Time, restDays int, points []model.Awareness) awarenessCycleResult {
	startDate = normalizeDate(startDate)
	targetDate = normalizeDate(targetDate)
	if restDays <= 0 {
		restDays = defaultAwarenessCycleRestDays
	}

	if targetDate.Before(startDate) {
		return awarenessCycleResult{
			IsPreStart: true,
		}
	}

	if len(points) == 0 {
		return awarenessCycleResult{
			IsRestDay:    true,
			RestDayIndex: 1,
		}
	}

	daysSinceStart := int(targetDate.Sub(startDate).Hours() / 24)
	cycleLength := len(points) + restDays
	cycleIndex := daysSinceStart / cycleLength
	dayInCycle := daysSinceStart % cycleLength

	if dayInCycle >= len(points) {
		return awarenessCycleResult{
			IsRestDay:    true,
			RestDayIndex: dayInCycle - len(points) + 1,
			CycleIndex:   cycleIndex,
		}
	}

	return awarenessCycleResult{
		Awareness:  &points[dayInCycle],
		CycleIndex: cycleIndex,
	}
}

func parseAwarenessCycleStart(input string) time.Time {
	parsed, err := time.ParseInLocation("2006-01-02", input, time.Local)
	if err != nil {
		parsed, _ = time.ParseInLocation("2006-01-02", "2026-05-01", time.Local)
	}

	return normalizeDate(parsed)
}
