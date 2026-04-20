package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ WeeklyTopicVotesModel = (*customWeeklyTopicVotesModel)(nil)

type (
	// WeeklyTopicVotesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWeeklyTopicVotesModel.
	WeeklyTopicVotesModel interface {
		weeklyTopicVotesModel
		withSession(session sqlx.Session) WeeklyTopicVotesModel
	}

	customWeeklyTopicVotesModel struct {
		*defaultWeeklyTopicVotesModel
	}
)

// NewWeeklyTopicVotesModel returns a model for the database table.
func NewWeeklyTopicVotesModel(conn sqlx.SqlConn) WeeklyTopicVotesModel {
	return &customWeeklyTopicVotesModel{
		defaultWeeklyTopicVotesModel: newWeeklyTopicVotesModel(conn),
	}
}

func (m *customWeeklyTopicVotesModel) withSession(session sqlx.Session) WeeklyTopicVotesModel {
	return NewWeeklyTopicVotesModel(sqlx.NewSqlConnFromSession(session))
}
