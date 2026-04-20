package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ DiscussionInfosModel = (*customDiscussionInfosModel)(nil)

type (
	// DiscussionInfosModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDiscussionInfosModel.
	DiscussionInfosModel interface {
		discussionInfosModel
		withSession(session sqlx.Session) DiscussionInfosModel
	}

	customDiscussionInfosModel struct {
		*defaultDiscussionInfosModel
	}
)

// NewDiscussionInfosModel returns a model for the database table.
func NewDiscussionInfosModel(conn sqlx.SqlConn) DiscussionInfosModel {
	return &customDiscussionInfosModel{
		defaultDiscussionInfosModel: newDiscussionInfosModel(conn),
	}
}

func (m *customDiscussionInfosModel) withSession(session sqlx.Session) DiscussionInfosModel {
	return NewDiscussionInfosModel(sqlx.NewSqlConnFromSession(session))
}
