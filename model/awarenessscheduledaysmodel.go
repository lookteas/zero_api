package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	awarenessScheduleDaysFieldNames          = builder.RawFieldNames(&AwarenessScheduleDays{})
	awarenessScheduleDaysRows                = strings.Join(awarenessScheduleDaysFieldNames, ",")
	awarenessScheduleDaysRowsExpectAutoSet   = strings.Join(stringx.Remove(awarenessScheduleDaysFieldNames, "`schedule_day_id`", "`created_at`", "`updated_at`"), ",")
	awarenessScheduleDaysRowsWithPlaceHolder = strings.Join(stringx.Remove(awarenessScheduleDaysFieldNames, "`schedule_day_id`", "`created_at`", "`updated_at`"), "=?,") + "=?"
)

type (
	AwarenessScheduleDaysModel interface {
		FindOneByCycleIdScheduleDate(ctx context.Context, cycleId uint64, scheduleDate time.Time) (*AwarenessScheduleDays, error)
		FindByCommunityDateRange(ctx context.Context, communityId uint64, startDate, endDate time.Time) ([]AwarenessScheduleDays, error)
		Insert(ctx context.Context, data *AwarenessScheduleDays) (sql.Result, error)
		Upsert(ctx context.Context, data *AwarenessScheduleDays) error
		DeleteFutureByCycle(ctx context.Context, cycleId uint64, fromDate time.Time) error
		withSession(session sqlx.Session) AwarenessScheduleDaysModel
	}

	customAwarenessScheduleDaysModel struct {
		conn  sqlx.SqlConn
		table string
	}

	AwarenessScheduleDays struct {
		ScheduleDayId     uint64         `db:"schedule_day_id"`
		CycleId           uint64         `db:"cycle_id"`
		CommunityId       uint64         `db:"community_id"`
		ScheduleDate      time.Time      `db:"schedule_date"`
		DayType           string         `db:"day_type"`
		AwarenessId       sql.NullInt64  `db:"awareness_id"`
		CycleIndex        int64          `db:"cycle_index"`
		CycleDayIndex     sql.NullInt64  `db:"cycle_day_index"`
		EffectiveDayIndex sql.NullInt64  `db:"effective_day_index"`
		PauseId           sql.NullInt64  `db:"pause_id"`
		PauseReason       sql.NullString `db:"pause_reason"`
		GeneratedVersion  uint64         `db:"generated_version"`
		CreatedAt         time.Time      `db:"created_at"`
		UpdatedAt         time.Time      `db:"updated_at"`
	}
)

func NewAwarenessScheduleDaysModel(conn sqlx.SqlConn) AwarenessScheduleDaysModel {
	return &customAwarenessScheduleDaysModel{
		conn:  conn,
		table: "`awareness_schedule_days`",
	}
}

func (m *customAwarenessScheduleDaysModel) FindOneByCycleIdScheduleDate(ctx context.Context, cycleId uint64, scheduleDate time.Time) (*AwarenessScheduleDays, error) {
	var resp AwarenessScheduleDays
	query := fmt.Sprintf("select %s from %s where `cycle_id` = ? and `schedule_date` = ? limit 1", awarenessScheduleDaysRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, cycleId, scheduleDate.Format("2006-01-02"))
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customAwarenessScheduleDaysModel) FindByCommunityDateRange(ctx context.Context, communityId uint64, startDate, endDate time.Time) ([]AwarenessScheduleDays, error) {
	var resp []AwarenessScheduleDays
	query := fmt.Sprintf("select %s from %s where `community_id` = ? and `schedule_date` >= ? and `schedule_date` <= ? order by `schedule_date` asc", awarenessScheduleDaysRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query, communityId, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	return resp, err
}

func (m *customAwarenessScheduleDaysModel) Insert(ctx context.Context, data *AwarenessScheduleDays) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, awarenessScheduleDaysRowsExpectAutoSet)
	return m.conn.ExecCtx(ctx, query, data.CycleId, data.CommunityId, data.ScheduleDate, data.DayType, data.AwarenessId, data.CycleIndex, data.CycleDayIndex, data.EffectiveDayIndex, data.PauseId, data.PauseReason, data.GeneratedVersion)
}

func (m *customAwarenessScheduleDaysModel) Upsert(ctx context.Context, data *AwarenessScheduleDays) error {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) on duplicate key update %s", m.table, awarenessScheduleDaysRowsExpectAutoSet, awarenessScheduleDaysRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query,
		data.CycleId, data.CommunityId, data.ScheduleDate, data.DayType, data.AwarenessId, data.CycleIndex, data.CycleDayIndex, data.EffectiveDayIndex, data.PauseId, data.PauseReason, data.GeneratedVersion,
		data.CycleId, data.CommunityId, data.ScheduleDate, data.DayType, data.AwarenessId, data.CycleIndex, data.CycleDayIndex, data.EffectiveDayIndex, data.PauseId, data.PauseReason, data.GeneratedVersion,
	)
	return err
}

func (m *customAwarenessScheduleDaysModel) DeleteFutureByCycle(ctx context.Context, cycleId uint64, fromDate time.Time) error {
	query := fmt.Sprintf("delete from %s where `cycle_id` = ? and `schedule_date` >= ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, cycleId, fromDate.Format("2006-01-02"))
	return err
}

func (m *customAwarenessScheduleDaysModel) withSession(session sqlx.Session) AwarenessScheduleDaysModel {
	return NewAwarenessScheduleDaysModel(sqlx.NewSqlConnFromSession(session))
}
