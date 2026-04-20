// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"api/internal/config"
	"api/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                         config.Config
	DB                             sqlx.SqlConn
	TopicsModel                    model.TopicsModel
	DailyTasksModel                model.DailyTasksModel
	DailyLogsModel                 model.DailyLogsModel
	ReviewItemsModel               model.ReviewItemsModel
	ReviewRecordsModel             model.ReviewRecordsModel
	UsersModel                     model.UsersModel
	AdminUsersModel                model.AdminUsersModel
	WeeklyTopicVotesModel          model.WeeklyTopicVotesModel
	WeeklyTopicVoteCandidatesModel model.WeeklyTopicVoteCandidatesModel
	WeeklyTopicVoteRecordsModel    model.WeeklyTopicVoteRecordsModel
	DiscussionInfosModel           model.DiscussionInfosModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	var db sqlx.SqlConn
	var topicsModel model.TopicsModel
	var dailyTasksModel model.DailyTasksModel
	var dailyLogsModel model.DailyLogsModel
	var reviewItemsModel model.ReviewItemsModel
	var reviewRecordsModel model.ReviewRecordsModel
	var usersModel model.UsersModel
	var adminUsersModel model.AdminUsersModel
	var weeklyTopicVotesModel model.WeeklyTopicVotesModel
	var weeklyTopicVoteCandidatesModel model.WeeklyTopicVoteCandidatesModel
	var weeklyTopicVoteRecordsModel model.WeeklyTopicVoteRecordsModel
	var discussionInfosModel model.DiscussionInfosModel

	if c.Mysql.DataSource != "" {
		db = sqlx.NewMysql(c.Mysql.DataSource)
		topicsModel = model.NewTopicsModel(db)
		dailyTasksModel = model.NewDailyTasksModel(db)
		dailyLogsModel = model.NewDailyLogsModel(db)
		reviewItemsModel = model.NewReviewItemsModel(db)
		reviewRecordsModel = model.NewReviewRecordsModel(db)
		usersModel = model.NewUsersModel(db)
		adminUsersModel = model.NewAdminUsersModel(db)
		weeklyTopicVotesModel = model.NewWeeklyTopicVotesModel(db)
		weeklyTopicVoteCandidatesModel = model.NewWeeklyTopicVoteCandidatesModel(db)
		weeklyTopicVoteRecordsModel = model.NewWeeklyTopicVoteRecordsModel(db)
		discussionInfosModel = model.NewDiscussionInfosModel(db)
	}

	return &ServiceContext{
		Config:                         c,
		DB:                             db,
		TopicsModel:                    topicsModel,
		DailyTasksModel:                dailyTasksModel,
		DailyLogsModel:                 dailyLogsModel,
		ReviewItemsModel:               reviewItemsModel,
		ReviewRecordsModel:             reviewRecordsModel,
		UsersModel:                     usersModel,
		AdminUsersModel:                adminUsersModel,
		WeeklyTopicVotesModel:          weeklyTopicVotesModel,
		WeeklyTopicVoteCandidatesModel: weeklyTopicVoteCandidatesModel,
		WeeklyTopicVoteRecordsModel:    weeklyTopicVoteRecordsModel,
		DiscussionInfosModel:           discussionInfosModel,
	}
}
