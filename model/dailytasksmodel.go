package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DailyTasksModel = (*customDailyTasksModel)(nil)

type (
	// DailyTasksModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDailyTasksModel.
	DailyTasksModel interface {
		dailyTasksModel
		withSession(session sqlx.Session) DailyTasksModel
	}

	customDailyTasksModel struct {
		*defaultDailyTasksModel
	}
)

// NewDailyTasksModel returns a model for the database table.
func NewDailyTasksModel(conn sqlx.SqlConn) DailyTasksModel {
	return &customDailyTasksModel{
		defaultDailyTasksModel: newDailyTasksModel(conn),
	}
}

func (m *customDailyTasksModel) withSession(session sqlx.Session) DailyTasksModel {
	return NewDailyTasksModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customDailyTasksModel) FindOneByUserIdTaskDate(ctx context.Context, userId uint64, taskDate time.Time) (*DailyTasks, error) {
	var resp DailyTasks
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `task_date` = ? limit 1", dailyTasksRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, taskDate.Format("2006-01-02"))
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
