package logic

import (
	"reflect"
	"strings"
	"testing"

	"api/internal/types"
)

func TestBuildReviewHistoryListQueryWithoutFilters(t *testing.T) {
	query, args := buildReviewHistoryListQuery(7, &types.ReviewHistoryQueryReq{})

	if !strings.Contains(query, "from review_records rr join review_items ri on rr.review_item_id = ri.id join daily_tasks dt on ri.daily_task_id = dt.id where rr.user_id = ?") {
		t.Fatalf("expected joined review history query, got %q", query)
	}
	if !strings.Contains(query, "order by rr.submitted_at desc limit 100") {
		t.Fatalf("expected submitted_at desc ordering, got %q", query)
	}
	if !reflect.DeepEqual(args, []any{uint64(7)}) {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildReviewHistoryListQueryWithFilters(t *testing.T) {
	req := &types.ReviewHistoryQueryReq{
		StartDate: "2026-04-01",
		EndDate:   "2026-04-19",
		Keyword:   "??",
	}

	query, args := buildReviewHistoryListQuery(9, req)

	expectedFragments := []string{
		"date(rr.submitted_at) >= ?",
		"date(rr.submitted_at) <= ?",
		"dt.topic_title like ?",
		"dt.topic_summary like ?",
		"rr.actual_situation like ?",
		"rr.suggestion like ?",
	}
	for _, fragment := range expectedFragments {
		if !strings.Contains(query, fragment) {
			t.Fatalf("expected query to contain %q, got %q", fragment, query)
		}
	}

	expectedArgs := []any{
		uint64(9),
		"2026-04-01",
		"2026-04-19",
		"%??%",
		"%??%",
		"%??%",
		"%??%",
		"%??%",
		"%??%",
		"%??%",
	}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Fatalf("unexpected args: %#v", args)
	}
}
