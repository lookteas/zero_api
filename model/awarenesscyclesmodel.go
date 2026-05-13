package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	awarenessCyclesFieldNames = builder.RawFieldNames(&AwarenessCycles{})
	awarenessCyclesRows       = strings.Join(awarenessCyclesFieldNames, ",")
)

type (
	AwarenessCyclesModel interface {
		FindActiveByCommunity(ctx context.Context, communityId uint64) (*AwarenessCycles, error)
		Update(ctx context.Context, data *AwarenessCycles) error
		withSession(session sqlx.Session) AwarenessCyclesModel
	}

	customAwarenessCyclesModel struct {
		conn  sqlx.SqlConn
		table string
	}

	AwarenessCycles struct {
		CycleId             uint64        `db:"cycle_id"`
		CommunityId         uint64        `db:"community_id"`
		CycleName           string        `db:"cycle_name"`
		StartDate           time.Time     `db:"start_date"`
		RestDays            int64         `db:"rest_days"`
		ScheduleHorizonDays int64         `db:"schedule_horizon_days"`
		Status              string        `db:"status"`
		LastGeneratedUntil  sql.NullTime  `db:"last_generated_until"`
		CreatedBy           sql.NullInt64 `db:"created_by"`
		CreatedAt           time.Time     `db:"created_at"`
		UpdatedAt           time.Time     `db:"updated_at"`
	}
)

func NewAwarenessCyclesModel(conn sqlx.SqlConn) AwarenessCyclesModel {
	return &customAwarenessCyclesModel{
		conn:  conn,
		table: "`awareness_cycles`",
	}
}

func (m *customAwarenessCyclesModel) FindActiveByCommunity(ctx context.Context, communityId uint64) (*AwarenessCycles, error) {
	var resp AwarenessCycles
	query := fmt.Sprintf("select %s from %s where `community_id` = ? and `status` = 'active' order by `cycle_id` desc limit 1", awarenessCyclesRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, communityId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customAwarenessCyclesModel) Update(ctx context.Context, data *AwarenessCycles) error {
	query := fmt.Sprintf("update %s set `community_id`=?, `cycle_name`=?, `start_date`=?, `rest_days`=?, `schedule_horizon_days`=?, `status`=?, `last_generated_until`=?, `created_by`=? where `cycle_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.CommunityId, data.CycleName, data.StartDate, data.RestDays, data.ScheduleHorizonDays, data.Status, data.LastGeneratedUntil, data.CreatedBy, data.CycleId)
	return err
}

func (m *customAwarenessCyclesModel) withSession(session sqlx.Session) AwarenessCyclesModel {
	return NewAwarenessCyclesModel(sqlx.NewSqlConnFromSession(session))
}
