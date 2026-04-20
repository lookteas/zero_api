package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ReviewRecordsModel = (*customReviewRecordsModel)(nil)

type (
	// ReviewRecordsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customReviewRecordsModel.
	ReviewRecordsModel interface {
		reviewRecordsModel
		withSession(session sqlx.Session) ReviewRecordsModel
	}

	customReviewRecordsModel struct {
		*defaultReviewRecordsModel
	}
)

// NewReviewRecordsModel returns a model for the database table.
func NewReviewRecordsModel(conn sqlx.SqlConn) ReviewRecordsModel {
	return &customReviewRecordsModel{
		defaultReviewRecordsModel: newReviewRecordsModel(conn),
	}
}

func (m *customReviewRecordsModel) withSession(session sqlx.Session) ReviewRecordsModel {
	return NewReviewRecordsModel(sqlx.NewSqlConnFromSession(session))
}
