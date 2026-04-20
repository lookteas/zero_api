package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ AuthVerificationCodesModel = (*customAuthVerificationCodesModel)(nil)

type (
	// AuthVerificationCodesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAuthVerificationCodesModel.
	AuthVerificationCodesModel interface {
		authVerificationCodesModel
		withSession(session sqlx.Session) AuthVerificationCodesModel
	}

	customAuthVerificationCodesModel struct {
		*defaultAuthVerificationCodesModel
	}
)

// NewAuthVerificationCodesModel returns a model for the database table.
func NewAuthVerificationCodesModel(conn sqlx.SqlConn) AuthVerificationCodesModel {
	return &customAuthVerificationCodesModel{
		defaultAuthVerificationCodesModel: newAuthVerificationCodesModel(conn),
	}
}

func (m *customAuthVerificationCodesModel) withSession(session sqlx.Session) AuthVerificationCodesModel {
	return NewAuthVerificationCodesModel(sqlx.NewSqlConnFromSession(session))
}
