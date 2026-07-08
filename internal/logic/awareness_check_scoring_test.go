package logic

import (
	"database/sql"
	"testing"
)

func TestScoreCheckPointHandlesHigherAndLower(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		selfScore  float64
		humanScore float64
		direction  string
		score      float64
		refScore   float64
		delta      float64
	}{
		{
			name:       "higher keeps self score",
			selfScore:  30,
			humanScore: 70,
			direction:  checkDirectionHigher,
			score:      30,
			refScore:   70,
			delta:      -40,
		},
		{
			name:       "lower reverses score",
			selfScore:  30,
			humanScore: 70,
			direction:  checkDirectionLower,
			score:      70,
			refScore:   30,
			delta:      40,
		},
		{
			name:       "empty direction defaults to higher",
			selfScore:  55,
			humanScore: 45,
			direction:  "",
			score:      55,
			refScore:   45,
			delta:      10,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			score, err := scoreCheckPoint(tt.selfScore, tt.humanScore, tt.direction)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertFloatEqual(t, score.Score, tt.score)
			assertFloatEqual(t, score.RefScore, tt.refScore)
			assertFloatEqual(t, score.Delta, tt.delta)
		})
	}
}

func TestScoreCheckPointRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	if _, err := scoreCheckPoint(-1, 50, checkDirectionHigher); err == nil {
		t.Fatalf("expected invalid selfScore error")
	}
	if _, err := scoreCheckPoint(50, 101, checkDirectionHigher); err == nil {
		t.Fatalf("expected invalid humanScore error")
	}
	if _, err := scoreCheckPoint(50, 50, "sideways"); err == nil {
		t.Fatalf("expected invalid direction error")
	}
}

func TestCheckHumanScoreUsesRangeOrFallback(t *testing.T) {
	t.Parallel()

	got := checkHumanScore(
		sql.NullFloat64{Float64: 20, Valid: true},
		sql.NullFloat64{Float64: 80, Valid: true},
	)
	assertFloatEqual(t, got, 50)

	got = checkHumanScore(sql.NullFloat64{}, sql.NullFloat64{Float64: 80, Valid: true})
	assertFloatEqual(t, got, 50)

	got = checkHumanScore(sql.NullFloat64{}, sql.NullFloat64{})
	assertFloatEqual(t, got, 50)
}

func TestAverageCheckScoresOnlyUsesDoneChapters(t *testing.T) {
	t.Parallel()

	higherScore, err := scoreCheckPoint(80, 60, checkDirectionHigher)
	if err != nil {
		t.Fatalf("unexpected higher score error: %v", err)
	}
	lowerScore, err := scoreCheckPoint(20, 70, checkDirectionLower)
	if err != nil {
		t.Fatalf("unexpected lower score error: %v", err)
	}

	chapter, ok := avgCheckPoints([]checkPointScore{higherScore, lowerScore})
	if !ok {
		t.Fatalf("expected chapter score")
	}
	assertFloatEqual(t, chapter.Score, 80)
	assertFloatEqual(t, chapter.RefScore, 45)
	assertFloatEqual(t, checkDelta(chapter), 35)

	overall, ok := avgDoneChapters([]checkChapterScore{
		{Status: checkStatusCompleted, Score: 70, RefScore: 60},
		{Status: "in_progress", Score: 100, RefScore: 100},
		{Status: checkStatusCompleted, Score: 50, RefScore: 55},
	})
	if !ok {
		t.Fatalf("expected overall score")
	}
	assertFloatEqual(t, overall.Score, 60)
	assertFloatEqual(t, overall.RefScore, 57.5)
	assertFloatEqual(t, checkDelta(overall), 2.5)

	if _, ok := avgDoneChapters([]checkChapterScore{{Status: "not_started"}}); ok {
		t.Fatalf("did not expect score without completed chapters")
	}
}

func assertFloatEqual(t *testing.T, got, want float64) {
	t.Helper()

	if got != want {
		t.Fatalf("got %.2f, want %.2f", got, want)
	}
}
