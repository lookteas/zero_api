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
	awarenessCyclePausesFieldNames        = builder.RawFieldNames(&AwarenessCyclePauses{})
	awarenessCyclePausesRows              = strings.Join(awarenessCyclePausesFieldNames, ",")
	awarenessCyclePausesRowsExpectAutoSet = strings.Join(stringx.Remove(awarenessCyclePausesFieldNames, "`pause_id`", "`created_at`", "`updated_at`"), ",")
)

type (
	AwarenessCyclePausesModel interface {
		FindActiveByCycle(ctx context.Context, cycleId uint64) ([]AwarenessCyclePauses, error)
		Insert(ctx context.Context, data *AwarenessCyclePauses) (sql.Result, error)
		DeleteByCycle(ctx context.Context, cycleId uint64) error
		withSession(session sqlx.Session) AwarenessCyclePausesModel
	}

	customAwarenessCyclePausesModel struct {
		conn  sqlx.SqlConn
		table string
	}

	AwarenessCyclePauses struct {
		PauseId        uint64         `db:"pause_id"`
		CycleId        uint64         `db:"cycle_id"`
		CommunityId    uint64         `db:"community_id"`
		PauseStartDate time.Time      `db:"pause_start_date"`
		PauseEndDate   time.Time      `db:"pause_end_date"`
		Reason         sql.NullString `db:"reason"`
		Status         int64          `db:"status"`
		CreatedBy      sql.NullInt64  `db:"created_by"`
		CreatedAt      time.Time      `db:"created_at"`
		UpdatedAt      time.Time      `db:"updated_at"`
	}
)

func NewAwarenessCyclePausesModel(conn sqlx.SqlConn) AwarenessCyclePausesModel {
	return &customAwarenessCyclePausesModel{
		conn:  conn,
		table: "`awareness_cycle_pauses`",
	}
}

func (m *customAwarenessCyclePausesModel) FindActiveByCycle(ctx context.Context, cycleId uint64) ([]AwarenessCyclePauses, error) {
	var resp []AwarenessCyclePauses
	query := fmt.Sprintf("select %s from %s where `cycle_id` = ? and `status` = 1 order by `pause_start_date` asc, `pause_id` asc", awarenessCyclePausesRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query, cycleId)
	return resp, err
}

func (m *customAwarenessCyclePausesModel) Insert(ctx context.Context, data *AwarenessCyclePauses) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?)", m.table, awarenessCyclePausesRowsExpectAutoSet)
	return m.conn.ExecCtx(ctx, query, data.CycleId, data.CommunityId, data.PauseStartDate, data.PauseEndDate, data.Reason, data.Status, data.CreatedBy)
}

func (m *customAwarenessCyclePausesModel) DeleteByCycle(ctx context.Context, cycleId uint64) error {
	query := fmt.Sprintf("delete from %s where `cycle_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, cycleId)
	return err
}

func (m *customAwarenessCyclePausesModel) withSession(session sqlx.Session) AwarenessCyclePausesModel {
	return NewAwarenessCyclePausesModel(sqlx.NewSqlConnFromSession(session))
}
