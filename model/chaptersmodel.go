package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	chapterFieldNames = builder.RawFieldNames(&Chapter{})
	chapterRows       = strings.Join(chapterFieldNames, ",")
)

var _ ChaptersModel = (*customChaptersModel)(nil)

type (
	ChaptersModel interface {
		FindAll(ctx context.Context) ([]Chapter, error)
		FindOne(ctx context.Context, id uint64) (*Chapter, error)
		withSession(session sqlx.Session) ChaptersModel
	}

	customChaptersModel struct {
		conn  sqlx.SqlConn
		table string
	}

	Chapter struct {
		ChapterId        uint64         `db:"chapter_id"`
		ChapterNo        int64          `db:"chapter_no"`
		ChapterArea      sql.NullString `db:"chapter_area"`
		ChapterTitle     string         `db:"chapter_title"`
		ChapterFullTitle string         `db:"chapter_full_title"`
		SortOrder        int64          `db:"sort_order"`
		SourceVolume     sql.NullString `db:"source_volume"`
		Notes            sql.NullString `db:"notes"`
	}
)

func NewChaptersModel(conn sqlx.SqlConn) ChaptersModel {
	return &customChaptersModel{
		conn:  conn,
		table: "`chapters`",
	}
}

func (m *customChaptersModel) FindAll(ctx context.Context) ([]Chapter, error) {
	var resp []Chapter
	query := fmt.Sprintf("select %s from %s order by `sort_order` asc, `chapter_no` asc, `chapter_id` asc", chapterRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

func (m *customChaptersModel) FindOne(ctx context.Context, id uint64) (*Chapter, error) {
	var resp Chapter
	query := fmt.Sprintf("select %s from %s where `chapter_id` = ? limit 1", chapterRows, m.table)
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

func (m *customChaptersModel) withSession(session sqlx.Session) ChaptersModel {
	return NewChaptersModel(sqlx.NewSqlConnFromSession(session))
}

func nullChapterString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}
