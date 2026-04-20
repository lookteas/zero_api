package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ DailyLogsModel = (*customDailyLogsModel)(nil)

type (
	// DailyLogsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDailyLogsModel.
	DailyLogsModel interface {
		dailyLogsModel
		withSession(session sqlx.Session) DailyLogsModel
	}

	customDailyLogsModel struct {
		*defaultDailyLogsModel
	}
)

// NewDailyLogsModel returns a model for the database table.
func NewDailyLogsModel(conn sqlx.SqlConn) DailyLogsModel {
	return &customDailyLogsModel{
		defaultDailyLogsModel: newDailyLogsModel(conn),
	}
}

func (m *customDailyLogsModel) withSession(session sqlx.Session) DailyLogsModel {
	return NewDailyLogsModel(sqlx.NewSqlConnFromSession(session))
}
