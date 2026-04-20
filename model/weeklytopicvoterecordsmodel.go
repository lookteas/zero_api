package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ WeeklyTopicVoteRecordsModel = (*customWeeklyTopicVoteRecordsModel)(nil)

type (
	// WeeklyTopicVoteRecordsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWeeklyTopicVoteRecordsModel.
	WeeklyTopicVoteRecordsModel interface {
		weeklyTopicVoteRecordsModel
		withSession(session sqlx.Session) WeeklyTopicVoteRecordsModel
	}

	customWeeklyTopicVoteRecordsModel struct {
		*defaultWeeklyTopicVoteRecordsModel
	}
)

// NewWeeklyTopicVoteRecordsModel returns a model for the database table.
func NewWeeklyTopicVoteRecordsModel(conn sqlx.SqlConn) WeeklyTopicVoteRecordsModel {
	return &customWeeklyTopicVoteRecordsModel{
		defaultWeeklyTopicVoteRecordsModel: newWeeklyTopicVoteRecordsModel(conn),
	}
}

func (m *customWeeklyTopicVoteRecordsModel) withSession(session sqlx.Session) WeeklyTopicVoteRecordsModel {
	return NewWeeklyTopicVoteRecordsModel(sqlx.NewSqlConnFromSession(session))
}
