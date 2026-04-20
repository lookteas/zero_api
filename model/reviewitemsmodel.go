package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ReviewItemsModel = (*customReviewItemsModel)(nil)

type (
	// ReviewItemsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customReviewItemsModel.
	ReviewItemsModel interface {
		reviewItemsModel
		withSession(session sqlx.Session) ReviewItemsModel
	}

	customReviewItemsModel struct {
		*defaultReviewItemsModel
	}
)

// NewReviewItemsModel returns a model for the database table.
func NewReviewItemsModel(conn sqlx.SqlConn) ReviewItemsModel {
	return &customReviewItemsModel{
		defaultReviewItemsModel: newReviewItemsModel(conn),
	}
}

func (m *customReviewItemsModel) withSession(session sqlx.Session) ReviewItemsModel {
	return NewReviewItemsModel(sqlx.NewSqlConnFromSession(session))
}
