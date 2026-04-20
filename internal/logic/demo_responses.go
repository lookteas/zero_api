package logic

import "api/internal/types"

func okSimple(message string) *types.SimpleResp {
	return &types.SimpleResp{
		Code:    0,
		Message: message,
	}
}

func okLogin() *types.LoginResp {
	return &types.LoginResp{
		Code:    0,
		Message: "ok",
		Data:    mockLoginData(),
	}
}

func okUser() *types.UserResp {
	return &types.UserResp{
		Code:    0,
		Message: "ok",
		Data:    mockUser(),
	}
}

func okHome() *types.HomeResp {
	return &types.HomeResp{
		Code:    0,
		Message: "ok",
		Data: types.HomeData{
			TodayTask: todayTaskPtr(mockTodayTask()),
			Overview: types.HomeOverview{
				ContinuousDays:    5,
				TotalTaskCount:    12,
				TotalReviewCount:  21,
				PendingReviewCount: 3,
			},
			PendingReviews:   []types.ReviewItemInfo{mockReviewItem()},
			CurrentVote:      mockWeeklyVote(),
			CurrentDiscussion: mockDiscussion(),
		},
	}
}

func todayTaskPtr(task types.DailyTaskInfo) *types.DailyTaskInfo {
	return &task
}

func okDailyTask() *types.DailyTaskResp {
	return &types.DailyTaskResp{Code: 0, Message: "ok", Data: mockTodayTask()}
}

func okDailyTaskList() *types.DailyTaskListResp {
	return &types.DailyTaskListResp{
		Code:    0,
		Message: "ok",
		Data: types.DailyTaskListData{
			List: []types.DailyTaskInfo{mockTodayTask()},
			Pagination: types.Pagination{Page: 1, PageSize: 20, Total: 1},
		},
	}
}

func okDailyLog() *types.DailyLogResp {
	return &types.DailyLogResp{Code: 0, Message: "ok", Data: mockDailyLog()}
}

func okDailyLogList() *types.DailyLogListResp {
	return &types.DailyLogListResp{
		Code:    0,
		Message: "ok",
		Data: types.DailyLogListData{
			List: []types.DailyLogInfo{mockDailyLog()},
			Pagination: types.Pagination{Page: 1, PageSize: 20, Total: 1},
		},
	}
}

func okReviewItem() *types.ReviewItemResp {
	return &types.ReviewItemResp{Code: 0, Message: "ok", Data: mockReviewItem()}
}

func okReviewItemList() *types.ReviewItemListResp {
	return &types.ReviewItemListResp{
		Code:    0,
		Message: "ok",
		Data: types.ReviewItemListData{
			List: []types.ReviewItemInfo{mockReviewItem()},
			Pagination: types.Pagination{Page: 1, PageSize: 20, Total: 1},
		},
	}
}

func okReviewRecordList() *types.ReviewRecordListResp {
	return &types.ReviewRecordListResp{
		Code:    0,
		Message: "ok",
		Data: types.ReviewRecordListData{
			List: []types.ReviewRecordInfo{mockReviewRecord()},
			Pagination: types.Pagination{Page: 1, PageSize: 20, Total: 1},
		},
	}
}

func okWeeklyVote() *types.WeeklyVoteResp {
	return &types.WeeklyVoteResp{Code: 0, Message: "ok", Data: mockWeeklyVote()}
}

func okDiscussion() *types.DiscussionResp {
	return &types.DiscussionResp{Code: 0, Message: "ok", Data: mockDiscussion()}
}

func okTopicList() *types.TopicListResp {
	return &types.TopicListResp{
		Code:    0,
		Message: "ok",
		Data: types.TopicListData{
			List: mockTopicList(),
			Pagination: types.Pagination{Page: 1, PageSize: 20, Total: int64(len(mockTopicList()))},
		},
	}
}
