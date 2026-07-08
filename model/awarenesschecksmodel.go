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
	awarenessCheckFieldNames = builder.RawFieldNames(&AwarenessCheck{})
	awarenessCheckRows       = strings.Join(awarenessCheckFieldNames, ",")
)

var _ AwarenessChecksModel = (*customAwarenessChecksModel)(nil)

type (
	AwarenessChecksModel interface {
		Insert(ctx context.Context, data *AwarenessCheck) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*AwarenessCheck, error)
		FindCurrentByUserID(ctx context.Context, userID uint64) (*AwarenessCheck, error)
		FindLatestDoneByUserID(ctx context.Context, userID uint64) (*AwarenessCheck, error)
		FindByUserID(ctx context.Context, userID uint64, limit int64) ([]AwarenessCheck, error)
		Update(ctx context.Context, data *AwarenessCheck) error
		withSession(session sqlx.Session) AwarenessChecksModel
	}

	customAwarenessChecksModel struct {
		conn  sqlx.SqlConn
		table string
	}

	AwarenessCheck struct {
		CheckId       uint64          `db:"check_id"`
		UserId        uint64          `db:"user_id"`
		Status        string          `db:"status"`
		DoneChapters  int64           `db:"done_chapters"`
		TotalChapters int64           `db:"total_chapters"`
		Score         sql.NullFloat64 `db:"score"`
		RefScore      sql.NullFloat64 `db:"ref_score"`
		Delta         sql.NullFloat64 `db:"delta"`
		PrevCheckId   sql.NullInt64   `db:"prev_check_id"`
		StartedAt     time.Time       `db:"started_at"`
		CompletedAt   sql.NullTime    `db:"completed_at"`
		CreatedAt     time.Time       `db:"created_at"`
		UpdatedAt     time.Time       `db:"updated_at"`
	}
)

func NewAwarenessChecksModel(conn sqlx.SqlConn) AwarenessChecksModel {
	return &customAwarenessChecksModel{
		conn:  conn,
		table: "`awareness_checks`",
	}
}

func (m *customAwarenessChecksModel) Insert(ctx context.Context, data *AwarenessCheck) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (`user_id`, `status`, `done_chapters`, `total_chapters`, `score`, `ref_score`, `delta`, `prev_check_id`, `started_at`, `completed_at`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table)
	return m.conn.ExecCtx(ctx, query,
		data.UserId,
		data.Status,
		data.DoneChapters,
		data.TotalChapters,
		data.Score,
		data.RefScore,
		data.Delta,
		data.PrevCheckId,
		data.StartedAt,
		data.CompletedAt,
	)
}

func (m *customAwarenessChecksModel) FindOne(ctx context.Context, id uint64) (*AwarenessCheck, error) {
	var resp AwarenessCheck
	query := fmt.Sprintf("select %s from %s where `check_id` = ? limit 1", awarenessCheckRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customAwarenessChecksModel) FindCurrentByUserID(ctx context.Context, userID uint64) (*AwarenessCheck, error) {
	var resp AwarenessCheck
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `status` in ('draft', 'in_progress') order by `check_id` desc limit 1", awarenessCheckRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, userID)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customAwarenessChecksModel) FindLatestDoneByUserID(ctx context.Context, userID uint64) (*AwarenessCheck, error) {
	var resp AwarenessCheck
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `status` = 'completed' order by `completed_at` desc, `check_id` desc limit 1", awarenessCheckRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, userID)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customAwarenessChecksModel) FindByUserID(ctx context.Context, userID uint64, limit int64) ([]AwarenessCheck, error) {
	var resp []AwarenessCheck
	query := fmt.Sprintf("select %s from %s where `user_id` = ? order by `started_at` desc, `check_id` desc", awarenessCheckRows, m.table)
	if limit > 0 {
		query += " limit ?"
		return resp, m.conn.QueryRowsCtx(ctx, &resp, query, userID, limit)
	}
	return resp, m.conn.QueryRowsCtx(ctx, &resp, query, userID)
}

func (m *customAwarenessChecksModel) Update(ctx context.Context, data *AwarenessCheck) error {
	query := fmt.Sprintf("update %s set `status` = ?, `done_chapters` = ?, `total_chapters` = ?, `score` = ?, `ref_score` = ?, `delta` = ?, `prev_check_id` = ?, `completed_at` = ?, `updated_at` = now() where `check_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query,
		data.Status,
		data.DoneChapters,
		data.TotalChapters,
		data.Score,
		data.RefScore,
		data.Delta,
		data.PrevCheckId,
		data.CompletedAt,
		data.CheckId,
	)
	return err
}

func (m *customAwarenessChecksModel) withSession(session sqlx.Session) AwarenessChecksModel {
	return NewAwarenessChecksModel(sqlx.NewSqlConnFromSession(session))
}
