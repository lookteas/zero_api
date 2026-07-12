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
	awarenessCheckChapterFieldNames = builder.RawFieldNames(&AwarenessCheckChapter{})
	awarenessCheckChapterRows       = strings.Join(awarenessCheckChapterFieldNames, ",")
)

var _ AwarenessCheckChaptersModel = (*customAwarenessCheckChaptersModel)(nil)

type (
	AwarenessCheckChaptersModel interface {
		Insert(ctx context.Context, data *AwarenessCheckChapter) (sql.Result, error)
		FindByCheckID(ctx context.Context, checkID uint64) ([]AwarenessCheckChapter, error)
		FindCompletedByUserID(ctx context.Context, userID uint64, limit int64) ([]AwarenessCheckChapter, error)
		FindOneByCheckAndChapter(ctx context.Context, checkID uint64, chapterID uint64) (*AwarenessCheckChapter, error)
		Update(ctx context.Context, data *AwarenessCheckChapter) error
		withSession(session sqlx.Session) AwarenessCheckChaptersModel
	}

	customAwarenessCheckChaptersModel struct {
		conn  sqlx.SqlConn
		table string
	}

	AwarenessCheckChapter struct {
		CheckChapterId uint64          `db:"check_chapter_id"`
		CheckId        uint64          `db:"check_id"`
		UserId         uint64          `db:"user_id"`
		ChapterId      uint64          `db:"chapter_id"`
		TotalPoints    int64           `db:"total_points"`
		ScoredPoints   int64           `db:"scored_points"`
		Score          sql.NullFloat64 `db:"score"`
		RefScore       sql.NullFloat64 `db:"ref_score"`
		Delta          sql.NullFloat64 `db:"delta"`
		PrevScore      sql.NullFloat64 `db:"prev_score"`
		ScoreChange    sql.NullFloat64 `db:"score_change"`
		Status         string          `db:"status"`
		SubmittedAt    sql.NullTime    `db:"submitted_at"`
		CreatedAt      time.Time       `db:"created_at"`
		UpdatedAt      time.Time       `db:"updated_at"`
	}
)

func NewAwarenessCheckChaptersModel(conn sqlx.SqlConn) AwarenessCheckChaptersModel {
	return &customAwarenessCheckChaptersModel{
		conn:  conn,
		table: "`awareness_check_chapters`",
	}
}

func (m *customAwarenessCheckChaptersModel) Insert(ctx context.Context, data *AwarenessCheckChapter) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (`check_id`, `user_id`, `chapter_id`, `total_points`, `scored_points`, `score`, `ref_score`, `delta`, `prev_score`, `score_change`, `status`, `submitted_at`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table)
	return m.conn.ExecCtx(ctx, query,
		data.CheckId,
		data.UserId,
		data.ChapterId,
		data.TotalPoints,
		data.ScoredPoints,
		data.Score,
		data.RefScore,
		data.Delta,
		data.PrevScore,
		data.ScoreChange,
		data.Status,
		data.SubmittedAt,
	)
}

func (m *customAwarenessCheckChaptersModel) FindByCheckID(ctx context.Context, checkID uint64) ([]AwarenessCheckChapter, error) {
	var resp []AwarenessCheckChapter
	query := fmt.Sprintf("select %s from %s where `check_id` = ? order by `chapter_id` asc", awarenessCheckChapterRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query, checkID)
	return resp, err
}

func (m *customAwarenessCheckChaptersModel) FindCompletedByUserID(ctx context.Context, userID uint64, limit int64) ([]AwarenessCheckChapter, error) {
	var resp []AwarenessCheckChapter
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `status` = 'completed' order by `submitted_at` desc, `check_id` desc, `check_chapter_id` desc", awarenessCheckChapterRows, m.table)
	if limit > 0 {
		query += " limit ?"
		return resp, m.conn.QueryRowsCtx(ctx, &resp, query, userID, limit)
	}
	return resp, m.conn.QueryRowsCtx(ctx, &resp, query, userID)
}

func (m *customAwarenessCheckChaptersModel) FindOneByCheckAndChapter(ctx context.Context, checkID uint64, chapterID uint64) (*AwarenessCheckChapter, error) {
	var resp AwarenessCheckChapter
	query := fmt.Sprintf("select %s from %s where `check_id` = ? and `chapter_id` = ? limit 1", awarenessCheckChapterRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, checkID, chapterID)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customAwarenessCheckChaptersModel) Update(ctx context.Context, data *AwarenessCheckChapter) error {
	query := fmt.Sprintf("update %s set `scored_points` = ?, `score` = ?, `ref_score` = ?, `delta` = ?, `prev_score` = ?, `score_change` = ?, `status` = ?, `submitted_at` = ?, `updated_at` = now() where `check_chapter_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query,
		data.ScoredPoints,
		data.Score,
		data.RefScore,
		data.Delta,
		data.PrevScore,
		data.ScoreChange,
		data.Status,
		data.SubmittedAt,
		data.CheckChapterId,
	)
	return err
}

func (m *customAwarenessCheckChaptersModel) withSession(session sqlx.Session) AwarenessCheckChaptersModel {
	return NewAwarenessCheckChaptersModel(sqlx.NewSqlConnFromSession(session))
}
