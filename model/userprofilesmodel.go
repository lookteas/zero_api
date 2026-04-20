package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ UserProfilesModel = (*customUserProfilesModel)(nil)

type (
	// UserProfilesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserProfilesModel.
	UserProfilesModel interface {
		userProfilesModel
		withSession(session sqlx.Session) UserProfilesModel
	}

	customUserProfilesModel struct {
		*defaultUserProfilesModel
	}
)

// NewUserProfilesModel returns a model for the database table.
func NewUserProfilesModel(conn sqlx.SqlConn) UserProfilesModel {
	return &customUserProfilesModel{
		defaultUserProfilesModel: newUserProfilesModel(conn),
	}
}

func (m *customUserProfilesModel) withSession(session sqlx.Session) UserProfilesModel {
	return NewUserProfilesModel(sqlx.NewSqlConnFromSession(session))
}
