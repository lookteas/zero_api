package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ OperationLogsModel = (*customOperationLogsModel)(nil)

type (
	// OperationLogsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customOperationLogsModel.
	OperationLogsModel interface {
		operationLogsModel
		withSession(session sqlx.Session) OperationLogsModel
	}

	customOperationLogsModel struct {
		*defaultOperationLogsModel
	}
)

// NewOperationLogsModel returns a model for the database table.
func NewOperationLogsModel(conn sqlx.SqlConn) OperationLogsModel {
	return &customOperationLogsModel{
		defaultOperationLogsModel: newOperationLogsModel(conn),
	}
}

func (m *customOperationLogsModel) withSession(session sqlx.Session) OperationLogsModel {
	return NewOperationLogsModel(sqlx.NewSqlConnFromSession(session))
}
