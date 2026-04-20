package logic

import (
    "fmt"
    "strings"
    "time"

    "api/internal/types"
    "api/model"
)

func nextTopicOrderNo(items []model.Topics) int64 {
    var maxOrder int64
    for _, item := range items {
        if item.OrderNo > maxOrder {
            maxOrder = item.OrderNo
        }
    }
    return maxOrder + 1
}

func parseDiscussionMeetingTime(input string, weekStart time.Time) time.Time {
    trimmed := strings.TrimSpace(input)
    if trimmed == "" {
        return defaultDiscussionTime(weekStart)
    }

    parsed, err := time.ParseInLocation("2006-01-02 15:04:05", trimmed, time.Local)
    if err != nil {
        return defaultDiscussionTime(weekStart)
    }
    return parsed
}

func normalizeDiscussionStatus(input string) string {
    status := strings.TrimSpace(input)
    if status == "" {
        return "published"
    }
    return status
}

func buildDiscussionShareTextFromFields(topicTitle, discussionTitle, meetingTime, meetingLink string) string {
    return buildDiscussionShareText(types.DiscussionInfo{
        TopicTitle:      topicTitle,
        DiscussionTitle: discussionTitle,
        MeetingTime:     meetingTime,
        MeetingLink:     meetingLink,
    })
}

func mysqlDuplicateMessage(entity string, err error) error {
    if err == nil {
        return nil
    }
    message := err.Error()
    if strings.Contains(message, "Duplicate entry") {
        return fmt.Errorf("%s已存在，请调整顺序或唯一值后再试", entity)
    }
    return err
}

