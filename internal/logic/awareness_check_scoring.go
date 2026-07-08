package logic

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
)

const (
	checkDirectionHigher = "higher"
	checkDirectionLower  = "lower"
	checkStatusCompleted = "completed"
)

type checkPointScore struct {
	SelfScore  float64
	HumanScore float64
	Score      float64
	RefScore   float64
	Delta      float64
}

type checkChapterScore struct {
	Status   string
	Score    float64
	RefScore float64
}

func checkHumanScore(referenceMin, referenceMax sql.NullFloat64) float64 {
	if !referenceMin.Valid || !referenceMax.Valid {
		return 50
	}
	return roundCheckScore(clampCheckScore((referenceMin.Float64 + referenceMax.Float64) / 2))
}

func scoreCheckPoint(selfScore, humanScore float64, direction string) (checkPointScore, error) {
	if err := validateCheckScore(selfScore, "selfScore"); err != nil {
		return checkPointScore{}, err
	}
	if err := validateCheckScore(humanScore, "humanScore"); err != nil {
		return checkPointScore{}, err
	}

	score, err := calcCheckScore(selfScore, direction)
	if err != nil {
		return checkPointScore{}, err
	}
	refScore, err := calcCheckScore(humanScore, direction)
	if err != nil {
		return checkPointScore{}, err
	}

	return checkPointScore{
		SelfScore:  roundCheckScore(selfScore),
		HumanScore: roundCheckScore(humanScore),
		Score:      score,
		RefScore:   refScore,
		Delta:      roundCheckScore(score - refScore),
	}, nil
}

func avgCheckPoints(scores []checkPointScore) (checkChapterScore, bool) {
	if len(scores) == 0 {
		return checkChapterScore{}, false
	}

	var scoreTotal float64
	var refTotal float64
	for _, score := range scores {
		scoreTotal += score.Score
		refTotal += score.RefScore
	}

	score := roundCheckScore(scoreTotal / float64(len(scores)))
	refScore := roundCheckScore(refTotal / float64(len(scores)))
	return checkChapterScore{
		Status:   checkStatusCompleted,
		Score:    score,
		RefScore: refScore,
	}, true
}

func avgDoneChapters(chapters []checkChapterScore) (checkChapterScore, bool) {
	var completed []checkChapterScore
	for _, chapter := range chapters {
		if chapter.Status == checkStatusCompleted {
			completed = append(completed, chapter)
		}
	}
	if len(completed) == 0 {
		return checkChapterScore{}, false
	}

	var scoreTotal float64
	var refTotal float64
	for _, chapter := range completed {
		scoreTotal += chapter.Score
		refTotal += chapter.RefScore
	}

	score := roundCheckScore(scoreTotal / float64(len(completed)))
	refScore := roundCheckScore(refTotal / float64(len(completed)))
	return checkChapterScore{
		Status:   checkStatusCompleted,
		Score:    score,
		RefScore: refScore,
	}, true
}

func checkDelta(score checkChapterScore) float64 {
	return roundCheckScore(score.Score - score.RefScore)
}

func calcCheckScore(value float64, direction string) (float64, error) {
	switch cleanCheckDirection(direction) {
	case checkDirectionHigher:
		return roundCheckScore(value), nil
	case checkDirectionLower:
		return roundCheckScore(100 - value), nil
	default:
		return 0, fmt.Errorf("unsupported direction: %s", direction)
	}
}

func cleanCheckDirection(value string) string {
	direction := strings.ToLower(strings.TrimSpace(value))
	if direction == "" {
		return checkDirectionHigher
	}
	return direction
}

func validateCheckScore(value float64, field string) error {
	if value < 0 || value > 100 {
		return fmt.Errorf("%s must be between 0 and 100", field)
	}
	return nil
}

func clampCheckScore(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func roundCheckScore(value float64) float64 {
	rounded := math.Round(value*100) / 100
	if rounded == 0 {
		return 0
	}
	return rounded
}
