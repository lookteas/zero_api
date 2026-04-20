package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ RefreshTokensModel = (*customRefreshTokensModel)(nil)

type (
	// RefreshTokensModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRefreshTokensModel.
	RefreshTokensModel interface {
		refreshTokensModel
		withSession(session sqlx.Session) RefreshTokensModel
	}

	customRefreshTokensModel struct {
		*defaultRefreshTokensModel
	}
)

// NewRefreshTokensModel returns a model for the database table.
func NewRefreshTokensModel(conn sqlx.SqlConn) RefreshTokensModel {
	return &customRefreshTokensModel{
		defaultRefreshTokensModel: newRefreshTokensModel(conn),
	}
}

func (m *customRefreshTokensModel) withSession(session sqlx.Session) RefreshTokensModel {
	return NewRefreshTokensModel(sqlx.NewSqlConnFromSession(session))
}
