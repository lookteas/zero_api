package logic

import (
	"testing"
	"time"
)

func TestBuildReinforcementHintsOnlyAppearsAtThirtyDayBoundary(t *testing.T) {
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	signals := []ReinforcementSignal{{TopicTitle: "Topic A", ObservedAt: now.AddDate(0, 0, -2), Source: "log"}}

	hints := BuildReinforcementHints(now, 29, signals)
	if len(hints) != 0 {
		t.Fatalf("expected no hints before 30-day boundary, got %d", len(hints))
	}
}

func TestBuildReinforcementHintsUsesOnlyLatestThirtyDays(t *testing.T) {
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	signals := []ReinforcementSignal{
		{TopicTitle: "Old Topic", ObservedAt: now.AddDate(0, 0, -50), Source: "review"},
		{TopicTitle: "Topic A", ObservedAt: now.AddDate(0, 0, -5), Source: "log"},
	}

	hints := BuildReinforcementHints(now, 30, signals)
	if len(hints) != 1 {
		t.Fatalf("expected 1 hint from latest 30-day window, got %d", len(hints))
	}
	if hints[0].TopicTitle != "Topic A" {
		t.Fatalf("expected latest-window topic to win, got %s", hints[0].TopicTitle)
	}
}

func TestBuildReinforcementHintsReturnsAtMostThreePrompts(t *testing.T) {
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	signals := []ReinforcementSignal{
		{TopicTitle: "A", ObservedAt: now.AddDate(0, 0, -1), Source: "log"},
		{TopicTitle: "A", ObservedAt: now.AddDate(0, 0, -2), Source: "review"},
		{TopicTitle: "B", ObservedAt: now.AddDate(0, 0, -1), Source: "log"},
		{TopicTitle: "B", ObservedAt: now.AddDate(0, 0, -2), Source: "review"},
		{TopicTitle: "C", ObservedAt: now.AddDate(0, 0, -1), Source: "log"},
		{TopicTitle: "D", ObservedAt: now.AddDate(0, 0, -1), Source: "review"},
	}

	hints := BuildReinforcementHints(now, 60, signals)
	if len(hints) != 3 {
		t.Fatalf("expected max 3 hints, got %d", len(hints))
	}
}

func TestBuildReinforcementHintsReturnsPromptTextWithoutScores(t *testing.T) {
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	signals := []ReinforcementSignal{{TopicTitle: "Topic A", ObservedAt: now.AddDate(0, 0, -2), Source: "review"}}

	hints := BuildReinforcementHints(now, 30, signals)
	if len(hints) != 1 {
		t.Fatalf("expected 1 hint, got %d", len(hints))
	}
	if hints[0].Prompt == "" {
		t.Fatalf("expected prompt text to be present")
	}
	if hints[0].SourceSummary == "" {
		t.Fatalf("expected source summary to be present")
	}
}
