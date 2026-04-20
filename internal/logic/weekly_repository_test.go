package logic

import (
	"testing"
	"time"

	"api/internal/types"
)

func TestFindTodayVoteRecordMatchesTodayOnly(t *testing.T) {
	now := time.Date(2026, 4, 16, 9, 0, 0, 0, time.Local)
	record, ok := findTodayVoteRecord([]types.VoteRecordInfo{
		{Id: 1, CandidateId: 2, CreatedAt: "2026-04-15 20:00:00"},
		{Id: 2, CandidateId: 3, CreatedAt: "2026-04-16 08:30:00"},
	}, now)

	if !ok {
		t.Fatal("expected today record")
	}
	if record.Id != 2 {
		t.Fatalf("expected record 2, got %d", record.Id)
	}
}
