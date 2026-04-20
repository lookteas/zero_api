package logic

import (
	"testing"

	"api/internal/types"
)

func TestBuildDailyTaskListQueryAddsKeywordFilterAcrossHistoryFields(t *testing.T) {
	t.Parallel()

	query, args := buildDailyTaskListQuery(7, &types.DailyTaskQueryReq{
		StartDate: "2026-04-10",
		EndDate:   "2026-04-16",
		Keyword:   "复盘",
	})

	expectedQuery := "select id, user_id, task_date, topic_id, topic_order_no, topic_title, topic_summary, weakness, improvement_plan, verification_path, reflection_note, status, submitted_at, created_at, updated_at from daily_tasks where user_id = ? and task_date >= ? and task_date <= ? and (topic_title like ? or weakness like ? or improvement_plan like ? or verification_path like ? or reflection_note like ?) order by task_date desc limit 100"
	if query != expectedQuery {
		t.Fatalf("unexpected query:\n%s", query)
	}

	if len(args) != 8 {
		t.Fatalf("expected 8 args, got %d", len(args))
	}

	for i, arg := range []any{uint64(7), "2026-04-10", "2026-04-16", "%复盘%", "%复盘%", "%复盘%", "%复盘%", "%复盘%"} {
		if args[i] != arg {
			t.Fatalf("unexpected arg %d: %#v", i, args[i])
		}
	}
}
