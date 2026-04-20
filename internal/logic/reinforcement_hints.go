package logic

import (
	"context"
	"sort"
	"strings"
	"time"

	"api/internal/svc"
	"api/internal/types"
)

type ReinforcementSignal struct {
	TopicTitle string
	ObservedAt time.Time
	Source     string
}

func BuildReinforcementHints(now time.Time, completedTaskCount int64, signals []ReinforcementSignal) []types.ReinforcementHintInfo {
	if completedTaskCount == 0 || completedTaskCount%30 != 0 {
		return nil
	}

	windowStart := now.AddDate(0, 0, -30)
	topicCounts := make(map[string]int)
	topicSources := make(map[string]map[string]struct{})
	for _, signal := range signals {
		if signal.TopicTitle == "" || signal.ObservedAt.Before(windowStart) {
			continue
		}
		topicCounts[signal.TopicTitle]++
		if _, ok := topicSources[signal.TopicTitle]; !ok {
			topicSources[signal.TopicTitle] = make(map[string]struct{})
		}
		topicSources[signal.TopicTitle][signal.Source] = struct{}{}
	}

	topics := make([]string, 0, len(topicCounts))
	for topic := range topicCounts {
		topics = append(topics, topic)
	}
	sort.Slice(topics, func(i, j int) bool {
		if topicCounts[topics[i]] == topicCounts[topics[j]] {
			return topics[i] < topics[j]
		}
		return topicCounts[topics[i]] > topicCounts[topics[j]]
	})
	if len(topics) > 3 {
		topics = topics[:3]
	}

	hints := make([]types.ReinforcementHintInfo, 0, len(topics))
	for _, topic := range topics {
		hints = append(hints, types.ReinforcementHintInfo{
			TopicTitle:    topic,
			Prompt:        "这个点最近反复出现，下阶段建议继续盯住。",
			SourceSummary: buildHintSourceSummary(topicSources[topic]),
		})
	}
	return hints
}

func buildHintSourceSummary(sources map[string]struct{}) string {
	if len(sources) == 0 {
		return "来自最近 30 天的复盘与觉察记录"
	}
	parts := make([]string, 0, len(sources))
	if _, ok := sources["review"]; ok {
		parts = append(parts, "复盘")
	}
	if _, ok := sources["log"]; ok {
		parts = append(parts, "觉察")
	}
	if len(parts) == 0 {
		return "来自最近 30 天的记录"
	}
	return "来自最近 30 天的" + strings.Join(parts, " + ")
}

type reinforcementSignalRow struct {
	TopicTitle string    `db:"topic_title"`
	ObservedAt time.Time `db:"observed_at"`
}

func loadReinforcementSignals(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64, now time.Time) []ReinforcementSignal {
	windowStart := now.AddDate(0, 0, -30)
	signals := make([]ReinforcementSignal, 0)

	var logRows []reinforcementSignalRow
	logQuery := "select dt.topic_title as topic_title, dl.created_at as observed_at from daily_logs dl join daily_tasks dt on dt.id = dl.daily_task_id where dl.user_id = ? and dl.created_at >= ? order by dl.created_at desc limit 100"
	if err := svcCtx.DB.QueryRowsCtx(ctx, &logRows, logQuery, userID, windowStart); err == nil {
		for _, row := range logRows {
			signals = append(signals, ReinforcementSignal{TopicTitle: row.TopicTitle, ObservedAt: row.ObservedAt, Source: "log"})
		}
	}

	var reviewRows []reinforcementSignalRow
	reviewQuery := "select dt.topic_title as topic_title, rr.submitted_at as observed_at from review_records rr join review_items ri on ri.id = rr.review_item_id join daily_tasks dt on dt.id = ri.daily_task_id where rr.user_id = ? and rr.submitted_at >= ? order by rr.submitted_at desc limit 100"
	if err := svcCtx.DB.QueryRowsCtx(ctx, &reviewRows, reviewQuery, userID, windowStart); err == nil {
		for _, row := range reviewRows {
			signals = append(signals, ReinforcementSignal{TopicTitle: row.TopicTitle, ObservedAt: row.ObservedAt, Source: "review"})
		}
	}

	return signals
}
