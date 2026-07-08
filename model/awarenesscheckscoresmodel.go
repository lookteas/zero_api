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
	awarenessCheckScoreFieldNames = builder.RawFieldNames(&AwarenessCheckScore{})
	awarenessCheckScoreRows       = strings.Join(awarenessCheckScoreFieldNames, ",")
)

var _ AwarenessCheckScoresModel = (*customAwarenessCheckScoresModel)(nil)

type (
	AwarenessCheckScoresModel interface {
		Upsert(ctx context.Context, data *AwarenessCheckScore) (sql.Result, error)
		FindByCheckID(ctx context.Context, checkID uint64) ([]AwarenessCheckScore, error)
		FindByCheckAndChapter(ctx context.Context, checkID uint64, chapterID uint64) ([]AwarenessCheckScore, error)
		withSession(session sqlx.Session) AwarenessCheckScoresModel
	}

	customAwarenessCheckScoresModel struct {
		conn  sqlx.SqlConn
		table string
	}

	AwarenessCheckScore struct {
		ScoreId     uint64          `db:"score_id"`
		CheckId     uint64          `db:"check_id"`
		UserId      uint64          `db:"user_id"`
		ChapterId   uint64          `db:"chapter_id"`
		AwarenessId uint64          `db:"awareness_id"`
		SelfScore   float64         `db:"self_score"`
		Score       float64         `db:"score"`
		RefScore    float64         `db:"ref_score"`
		Delta       float64         `db:"delta"`
		PrevScore   sql.NullFloat64 `db:"prev_score"`
		ScoreChange sql.NullFloat64 `db:"score_change"`
		CreatedAt   time.Time       `db:"created_at"`
		UpdatedAt   time.Time       `db:"updated_at"`
	}
)

func NewAwarenessCheckScoresModel(conn sqlx.SqlConn) AwarenessCheckScoresModel {
	return &customAwarenessCheckScoresModel{
		conn:  conn,
		table: "`awareness_check_scores`",
	}
}

func (m *customAwarenessCheckScoresModel) Upsert(ctx context.Context, data *AwarenessCheckScore) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (`check_id`, `user_id`, `chapter_id`, `awareness_id`, `self_score`, `score`, `ref_score`, `delta`, `prev_score`, `score_change`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) on duplicate key update `self_score` = values(`self_score`), `score` = values(`score`), `ref_score` = values(`ref_score`), `delta` = values(`delta`), `prev_score` = values(`prev_score`), `score_change` = values(`score_change`), `updated_at` = now()", m.table)
	return m.conn.ExecCtx(ctx, query,
		data.CheckId,
		data.UserId,
		data.ChapterId,
		data.AwarenessId,
		data.SelfScore,
		data.Score,
		data.RefScore,
		data.Delta,
		data.PrevScore,
		data.ScoreChange,
	)
}

func (m *customAwarenessCheckScoresModel) FindByCheckID(ctx context.Context, checkID uint64) ([]AwarenessCheckScore, error) {
	var resp []AwarenessCheckScore
	query := fmt.Sprintf("select %s from %s where `check_id` = ? order by `chapter_id` asc, `awareness_id` asc", awarenessCheckScoreRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query, checkID)
	return resp, err
}

func (m *customAwarenessCheckScoresModel) FindByCheckAndChapter(ctx context.Context, checkID uint64, chapterID uint64) ([]AwarenessCheckScore, error) {
	var resp []AwarenessCheckScore
	query := fmt.Sprintf("select %s from %s where `check_id` = ? and `chapter_id` = ? order by `awareness_id` asc", awarenessCheckScoreRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query, checkID, chapterID)
	return resp, err
}

func (m *customAwarenessCheckScoresModel) withSession(session sqlx.Session) AwarenessCheckScoresModel {
	return NewAwarenessCheckScoresModel(sqlx.NewSqlConnFromSession(session))
}
