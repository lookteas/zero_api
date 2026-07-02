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
	freeModePracticeFieldNames = builder.RawFieldNames(&FreeModePractice{})
	freeModePracticeRows       = strings.Join(freeModePracticeFieldNames, ",")
)

var _ FreeModePracticesModel = (*customFreeModePracticesModel)(nil)

type (
	FreeModePracticesModel interface {
		Insert(ctx context.Context, data *FreeModePractice) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*FreeModePractice, error)
		FindByUserID(ctx context.Context, userID uint64, limit int64) ([]FreeModePractice, error)
		Update(ctx context.Context, data *FreeModePractice) error
		withSession(session sqlx.Session) FreeModePracticesModel
	}

	customFreeModePracticesModel struct {
		conn  sqlx.SqlConn
		table string
	}

	FreeModePractice struct {
		PracticeId       uint64         `db:"practice_id"`
		UserId           uint64         `db:"user_id"`
		PracticeDate     time.Time      `db:"practice_date"`
		ChapterId        uint64         `db:"chapter_id"`
		ChapterNo        int64          `db:"chapter_no"`
		ChapterTitle     string         `db:"chapter_title"`
		ChapterFullTitle string         `db:"chapter_full_title"`
		AwarenessId      uint64         `db:"awareness_id"`
		SectionId        uint64         `db:"section_id"`
		AwarenessTitle   string         `db:"awareness_title"`
		AwarenessSummary sql.NullString `db:"awareness_summary"`
		AwarenessDetails sql.NullString `db:"awareness_details"`
		PracticeNote     sql.NullString `db:"practice_note"`
		CreatedAt        time.Time      `db:"created_at"`
		UpdatedAt        time.Time      `db:"updated_at"`
	}
)

func NewFreeModePracticesModel(conn sqlx.SqlConn) FreeModePracticesModel {
	return &customFreeModePracticesModel{
		conn:  conn,
		table: "`free_mode_practices`",
	}
}

func (m *customFreeModePracticesModel) Insert(ctx context.Context, data *FreeModePractice) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (`user_id`, `practice_date`, `chapter_id`, `chapter_no`, `chapter_title`, `chapter_full_title`, `awareness_id`, `section_id`, `awareness_title`, `awareness_summary`, `awareness_details`, `practice_note`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table)
	return m.conn.ExecCtx(ctx, query,
		data.UserId,
		data.PracticeDate,
		data.ChapterId,
		data.ChapterNo,
		data.ChapterTitle,
		data.ChapterFullTitle,
		data.AwarenessId,
		data.SectionId,
		data.AwarenessTitle,
		data.AwarenessSummary,
		data.AwarenessDetails,
		data.PracticeNote,
	)
}

func (m *customFreeModePracticesModel) FindOne(ctx context.Context, id uint64) (*FreeModePractice, error) {
	var resp FreeModePractice
	query := fmt.Sprintf("select %s from %s where `practice_id` = ? limit 1", freeModePracticeRows, m.table)
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

func (m *customFreeModePracticesModel) FindByUserID(ctx context.Context, userID uint64, limit int64) ([]FreeModePractice, error) {
	var resp []FreeModePractice
	query := fmt.Sprintf("select %s from %s where `user_id` = ? order by `practice_date` desc, `practice_id` desc", freeModePracticeRows, m.table)
	if limit > 0 {
		query += " limit ?"
		return resp, m.conn.QueryRowsCtx(ctx, &resp, query, userID, limit)
	}
	return resp, m.conn.QueryRowsCtx(ctx, &resp, query, userID)
}

func (m *customFreeModePracticesModel) Update(ctx context.Context, data *FreeModePractice) error {
	query := fmt.Sprintf("update %s set `practice_note` = ?, `updated_at` = now() where `practice_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.PracticeNote, data.PracticeId)
	return err
}

func (m *customFreeModePracticesModel) withSession(session sqlx.Session) FreeModePracticesModel {
	return NewFreeModePracticesModel(sqlx.NewSqlConnFromSession(session))
}
