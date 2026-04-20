package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TopicsModel = (*customTopicsModel)(nil)

type (
	// TopicsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTopicsModel.
	TopicsModel interface {
		topicsModel
		FindLatestActiveByScheduleDate(ctx context.Context, scheduleDate time.Time) (*Topics, error)
		withSession(session sqlx.Session) TopicsModel
	}

	customTopicsModel struct {
		*defaultTopicsModel
	}
)

// NewTopicsModel returns a model for the database table.
func NewTopicsModel(conn sqlx.SqlConn) TopicsModel {
	return &customTopicsModel{
		defaultTopicsModel: newTopicsModel(conn),
	}
}

func (m *customTopicsModel) withSession(session sqlx.Session) TopicsModel {
	return NewTopicsModel(sqlx.NewSqlConnFromSession(session))
}


func (m *customTopicsModel) FindLatestActiveByScheduleDate(ctx context.Context, scheduleDate time.Time) (*Topics, error) {
	var resp Topics
	query := "select " + topicsRows + " from " + m.table + " where `status` = 1 and `schedule_date` <= ? order by `schedule_date` desc, `order_no` asc, `id` asc limit 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, scheduleDate)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		var fallback Topics
		fallbackQuery := "select " + topicsRows + " from " + m.table + " where `status` = 1 order by `schedule_date` is null asc, `schedule_date` asc, `order_no` asc, `id` asc limit 1"
		fallbackErr := m.conn.QueryRowCtx(ctx, &fallback, fallbackQuery)
		switch fallbackErr {
		case nil:
			return &fallback, nil
		case sqlx.ErrNotFound:
			return nil, ErrNotFound
		default:
			return nil, fallbackErr
		}
	default:
		return nil, err
	}
}
