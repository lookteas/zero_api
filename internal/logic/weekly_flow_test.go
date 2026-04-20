package logic

import (
	"testing"
	"time"

	"api/internal/types"
)

func TestCurrentWeekStartUsesMonday(t *testing.T) {
	input := time.Date(2026, 4, 16, 9, 30, 0, 0, time.Local)
	got := currentWeekStart(input)

	if got.Format("2006-01-02 15:04:05") != "2026-04-13 00:00:00" {
		t.Fatalf("expected monday start, got %s", got.Format("2006-01-02 15:04:05"))
	}
}

func TestDefaultDiscussionTimeIsSaturdayNightEight(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.Local)
	got := defaultDiscussionTime(weekStart)

	if got.Format("2006-01-02 15:04:05") != "2026-04-18 20:00:00" {
		t.Fatalf("expected saturday 20:00, got %s", got.Format("2006-01-02 15:04:05"))
	}
}

func TestDefaultVoteEndTimeIsFridayNight(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.Local)
	got := defaultVoteEndTime(weekStart)

	if got.Format("2006-01-02 15:04:05") != "2026-04-17 23:59:59" {
		t.Fatalf("expected friday 23:59:59, got %s", got.Format("2006-01-02 15:04:05"))
	}
}

func TestVoteCandidateWindowStartsSaturdayAndEndsFriday(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.Local)
	start, end := voteCandidateWindow(weekStart)

	if start.Format("2006-01-02") != "2026-04-18" {
		t.Fatalf("expected saturday start, got %s", start.Format("2006-01-02"))
	}
	if end.Format("2006-01-02") != "2026-04-24" {
		t.Fatalf("expected friday end, got %s", end.Format("2006-01-02"))
	}
}

func TestVoteCandidateDateUsesSaturdayToFridaySlots(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.Local)

	if got := voteCandidateDate(weekStart, 1).Format("2006-01-02"); got != "2026-04-18" {
		t.Fatalf("expected slot 1 to map to saturday, got %s", got)
	}
	if got := voteCandidateDate(weekStart, 7).Format("2006-01-02"); got != "2026-04-24" {
		t.Fatalf("expected slot 7 to map to friday, got %s", got)
	}
}

func TestPickWinningCandidateUsesVoteCountThenSortNo(t *testing.T) {
	winner, ok := pickWinningCandidate([]types.VoteCandidateInfo{
		{Id: 1, TopicTitle: "A", VoteCount: 3, SortNo: 2},
		{Id: 2, TopicTitle: "B", VoteCount: 8, SortNo: 3},
		{Id: 3, TopicTitle: "C", VoteCount: 8, SortNo: 1},
	})

	if !ok {
		t.Fatal("expected winner")
	}

	if winner.Id != 3 {
		t.Fatalf("expected candidate 3 to win tie by sort order, got %d", winner.Id)
	}
}

func TestBuildDiscussionShareTextIncludesTopicAndTime(t *testing.T) {
	text := buildDiscussionShareText(types.DiscussionInfo{
		TopicTitle:      "沉迷拔出能力",
		DiscussionTitle: "本周讨论：沉迷拔出能力",
		MeetingTime:     "2026-04-18 20:00:00",
	})

	if text == "" {
		t.Fatal("expected non-empty share text")
	}

	if want := "沉迷拔出能力"; !containsText(text, want) {
		t.Fatalf("expected share text to include %q, got %q", want, text)
	}

	if want := "周六 20:00"; !containsText(text, want) {
		t.Fatalf("expected share text to include %q, got %q", want, text)
	}
}

func containsText(got string, want string) bool {
	return len(got) >= len(want) && (got == want || containsText(got[1:], want) || got[:len(want)] == want)
}
