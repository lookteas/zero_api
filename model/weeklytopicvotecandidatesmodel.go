package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ WeeklyTopicVoteCandidatesModel = (*customWeeklyTopicVoteCandidatesModel)(nil)

type (
	// WeeklyTopicVoteCandidatesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWeeklyTopicVoteCandidatesModel.
	WeeklyTopicVoteCandidatesModel interface {
		weeklyTopicVoteCandidatesModel
		withSession(session sqlx.Session) WeeklyTopicVoteCandidatesModel
	}

	customWeeklyTopicVoteCandidatesModel struct {
		*defaultWeeklyTopicVoteCandidatesModel
	}
)

// NewWeeklyTopicVoteCandidatesModel returns a model for the database table.
func NewWeeklyTopicVoteCandidatesModel(conn sqlx.SqlConn) WeeklyTopicVoteCandidatesModel {
	return &customWeeklyTopicVoteCandidatesModel{
		defaultWeeklyTopicVoteCandidatesModel: newWeeklyTopicVoteCandidatesModel(conn),
	}
}

func (m *customWeeklyTopicVoteCandidatesModel) withSession(session sqlx.Session) WeeklyTopicVoteCandidatesModel {
	return NewWeeklyTopicVoteCandidatesModel(sqlx.NewSqlConnFromSession(session))
}
