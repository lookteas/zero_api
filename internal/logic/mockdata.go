package logic

import "api/internal/types"

func mockUser() types.UserInfo {
	return types.UserInfo{
		Id:       1,
		Account:  "demo_user",
		Email:    "demo@example.com",
		Mobile:   "13800000000",
		Nickname: "演示用户",
		Avatar:   "",
		Status:   1,
	}
}

func mockLoginData() types.LoginData {
	return types.LoginData{
		AccessToken:  "demo-access-token",
		RefreshToken: "demo-refresh-token",
		AccessExpire: 86400,
		User:         mockUser(),
	}
}

func mockTodayTask() types.DailyTaskInfo {
	return types.DailyTaskInfo{
		Id:               1001,
		TaskDate:         "2026-04-13",
		TopicId:          1,
		TopicOrderNo:     1,
		TopicTitle:       "灵活切换能力",
		TopicSummary:     "在不同情绪与场景之间更稳定地完成意识切换",
		Weakness:         "遇到冲突时容易持续陷入愤怒状态",
		ImprovementPlan:  "冲突发生后先深呼吸三次，再组织回应",
		VerificationPath: "记录今天两次冲突场景中的切换动作是否完成",
		Status:           "draft",
		CreatedAt:        "2026-04-13 09:00:00",
		UpdatedAt:        "2026-04-13 09:00:00",
	}
}

func mockDailyLog() types.DailyLogInfo {
	return types.DailyLogInfo{
		Id:          2001,
		DailyTaskId: 1001,
		LogTime:     "2026-04-13 11:30:00",
		ActionText:  "在争执发生后停顿 10 秒并做了 3 次呼吸",
		Status:      "done",
		Remark:      "切换比以往更快，情绪强度从 8 降到 5",
		CreatedAt:   "2026-04-13 11:31:00",
		UpdatedAt:   "2026-04-13 11:31:00",
	}
}

func mockReviewRecord() types.ReviewRecordInfo {
	return types.ReviewRecordInfo{
		Id:              3001,
		ReviewItemId:    4001,
		Result:          "partial",
		ActualSituation: "当天执行了 1 次，但在第二次冲突里没有做到",
		Suggestion:      "需要在上午提前提醒自己今天的验证动作",
		SubmittedAt:     "2026-04-13 21:10:00",
	}
}

func mockReviewItem() types.ReviewItemInfo {
	return types.ReviewItemInfo{
		Id:           4001,
		DailyTaskId:  1001,
		ReviewStage:  "day0",
		DueAt:        "2026-04-13 21:00:00",
		Status:       "pending",
		DailyTask:    mockTodayTask(),
		LatestRecord: mockReviewRecord(),
	}
}

func mockWeeklyVote() types.WeeklyVoteInfo {
	return types.WeeklyVoteInfo{
		Id:              5001,
		WeekStartDate:   "2026-04-13",
		VoteStartAt:     "2026-04-13 00:00:00",
		VoteEndAt:       "2026-04-17 23:59:59",
		Status:          "active",
		ResultTopicId:   0,
		HasVoted:        false,
		UserCandidateId: 0,
		Candidates: []types.VoteCandidateInfo{
			{Id: 1, TopicId: 1, TopicTitle: "灵活切换能力", TopicSummary: "提升情绪与状态切换能力", VoteCount: 12, SortNo: 1},
			{Id: 2, TopicId: 2, TopicTitle: "沉迷拔出能力", TopicSummary: "从沉迷状态中及时抽离", VoteCount: 18, SortNo: 2},
			{Id: 3, TopicId: 3, TopicTitle: "自我觉察力", TopicSummary: "提升对情绪与状态的感知", VoteCount: 9, SortNo: 3},
		},
	}
}

func mockDiscussion() types.DiscussionInfo {
	return types.DiscussionInfo{
		Id:              6001,
		WeekStartDate:   "2026-04-13",
		TopicId:         2,
		TopicTitle:      "沉迷拔出能力",
		DiscussionTitle: "本周意识强度讨论",
		Description:     "围绕本周高票主题，分享实践中的困难与有效切换方法。",
		Goals:           "1. 分享真实场景\n2. 总结切换方法\n3. 形成下周行动计划",
		MeetingTime:     "2026-04-19 20:00:00",
		MeetingLink:     "https://example.com/meeting-room",
		ShareText:       "欢迎参加本周意识强度讨论，主题：沉迷拔出能力。",
		Status:          "published",
	}
}

func mockTopicList() []types.TopicInfo {
	return []types.TopicInfo{
		{Id: 1, Title: "灵活切换能力", Summary: "提升情绪与状态切换能力", Description: "学会在情绪变化时快速恢复觉察。", OrderNo: 1, Status: 1},
		{Id: 2, Title: "沉迷拔出能力", Summary: "从沉迷状态中及时抽离", Description: "在刷手机、拖延等状态里练习抽离。", OrderNo: 2, Status: 1},
		{Id: 3, Title: "自我觉察力", Summary: "提升对情绪与状态的感知", Description: "让日常行为从自动化回到有意识。", OrderNo: 3, Status: 1},
	}
}
