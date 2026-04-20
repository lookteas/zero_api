package logic

import (
	"fmt"
	"strings"
	"time"

	"api/internal/types"
)

func currentWeekStart(now time.Time) time.Time {
	date := normalizeDate(now)
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	return date.AddDate(0, 0, -(weekday - 1))
}

func voteCandidateWindow(weekStart time.Time) (time.Time, time.Time) {
	start := normalizeDate(weekStart).AddDate(0, 0, 5)
	end := start.AddDate(0, 0, 6)
	return start, end
}

func voteCandidateDate(weekStart time.Time, sortNo int64) time.Time {
	start, _ := voteCandidateWindow(weekStart)
	offset := int(sortNo - 1)
	if offset < 0 {
		offset = 0
	}
	return start.AddDate(0, 0, offset)
}

func voteCandidateDateLabel(date time.Time) string {
	weekdayMap := map[time.Weekday]string{
		time.Sunday:    "周日",
		time.Monday:    "周一",
		time.Tuesday:   "周二",
		time.Wednesday: "周三",
		time.Thursday:  "周四",
		time.Friday:    "周五",
		time.Saturday:  "周六",
	}

	return fmt.Sprintf("%s %s", date.Format("2006-01-02"), weekdayMap[date.Weekday()])
}

func defaultDiscussionTime(weekStart time.Time) time.Time {
	start := normalizeDate(weekStart)
	return time.Date(start.Year(), start.Month(), start.Day()+5, 20, 0, 0, 0, start.Location())
}

func defaultVoteEndTime(weekStart time.Time) time.Time {
	start := normalizeDate(weekStart)
	return time.Date(start.Year(), start.Month(), start.Day()+4, 23, 59, 59, 0, start.Location())
}

func pickWinningCandidate(candidates []types.VoteCandidateInfo) (types.VoteCandidateInfo, bool) {
	if len(candidates) == 0 {
		return types.VoteCandidateInfo{}, false
	}

	winner := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.VoteCount > winner.VoteCount {
			winner = candidate
			continue
		}

		if candidate.VoteCount == winner.VoteCount {
			if candidate.SortNo < winner.SortNo || (candidate.SortNo == winner.SortNo && candidate.Id < winner.Id) {
				winner = candidate
			}
		}
	}

	return winner, true
}

func defaultDiscussionTitle(topicTitle string) string {
	if strings.TrimSpace(topicTitle) == "" {
		return "本周讨论"
	}

	return fmt.Sprintf("本周讨论：%s", strings.TrimSpace(topicTitle))
}

func defaultDiscussionDescription(topicTitle string) string {
	if strings.TrimSpace(topicTitle) == "" {
		return "这次围绕本周票数最高的话题，一起聊聊真实练习时最容易卡住的地方。"
	}

	return fmt.Sprintf("这次围绕“%s”展开，重点聊真实练习场景、最容易卡住的瞬间，以及下周还能怎么继续练。", strings.TrimSpace(topicTitle))
}

func defaultDiscussionGoals(topicTitle string) string {
	base := []string{
		"1. 讲一个这周最真实的练习场景",
		"2. 说清楚自己最容易卡住的点",
		"3. 带走一个下周继续练的小动作",
	}

	if strings.TrimSpace(topicTitle) == "" {
		return strings.Join(base, "\n")
	}

	base[1] = fmt.Sprintf("2. 说清楚在“%s”里自己最容易卡住的点", strings.TrimSpace(topicTitle))
	return strings.Join(base, "\n")
}

func discussionTimeLabel(meetingTime string) string {
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", meetingTime, time.Local)
	if err != nil {
		return "周六 20:00"
	}

	weekdayMap := map[time.Weekday]string{
		time.Sunday:    "周日",
		time.Monday:    "周一",
		time.Tuesday:   "周二",
		time.Wednesday: "周三",
		time.Thursday:  "周四",
		time.Friday:    "周五",
		time.Saturday:  "周六",
	}

	return fmt.Sprintf("%s %02d:%02d", weekdayMap[parsed.Weekday()], parsed.Hour(), parsed.Minute())
}

func buildDiscussionShareText(info types.DiscussionInfo) string {
	title := strings.TrimSpace(info.DiscussionTitle)
	if title == "" {
		title = defaultDiscussionTitle(info.TopicTitle)
	}

	lines := []string{
		title,
		fmt.Sprintf("主题：%s", strings.TrimSpace(info.TopicTitle)),
		fmt.Sprintf("时间：%s", discussionTimeLabel(info.MeetingTime)),
	}

	if strings.TrimSpace(info.MeetingLink) != "" {
		lines = append(lines, fmt.Sprintf("入口：%s", strings.TrimSpace(info.MeetingLink)))
	}

	return strings.Join(lines, "\n")
}
