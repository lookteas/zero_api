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
	awarenessFieldNames = builder.RawFieldNames(&Awareness{})
	awarenessRows       = strings.Join(awarenessFieldNames, ",")
)

var _ AwarenessModel = (*customAwarenessModel)(nil)

type (
	AwarenessModel interface {
		FindEligible(ctx context.Context) ([]Awareness, error)
		UpdateContent(ctx context.Context, id uint64, title, summary, details string) error
		Disable(ctx context.Context, id uint64) error
		withSession(session sqlx.Session) AwarenessModel
	}

	customAwarenessModel struct {
		conn  sqlx.SqlConn
		table string
	}

	Awareness struct {
		AwarenessId      uint64          `db:"awareness_id"`
		ChapterId        uint64          `db:"chapter_id"`
		RegionId         uint64          `db:"region_id"`
		SectionId        uint64          `db:"section_id"`
		SourceVolume     sql.NullString  `db:"source_volume"`
		SourceFile       sql.NullString  `db:"source_file"`
		PointNo          sql.NullString  `db:"point_no"`
		PointTitle       string          `db:"point_title"`
		Theme            sql.NullString  `db:"theme"`
		Summary          sql.NullString  `db:"summary"`
		QuantitativeData sql.NullString  `db:"quantitative_data"`
		ReferenceMin     sql.NullFloat64 `db:"reference_min"`
		ReferenceMax     sql.NullFloat64 `db:"reference_max"`
		ValueUnit        string          `db:"value_unit"`
		BetterDirection  string          `db:"better_direction"`
		Details          sql.NullString  `db:"details"`
		IsMeta           int64           `db:"is_meta"`
		Status           int64           `db:"status"`
		HasImages        int64           `db:"has_images"`
		ImageCount       int64           `db:"image_count"`
		CoverImageId     sql.NullInt64   `db:"cover_image_id"`
		ImageNotes       sql.NullString  `db:"image_notes"`
		ImagesJson       sql.NullString  `db:"images_json"`
		SortOrderGlobal  int64           `db:"sort_order_global"`
		CreatedAt        time.Time       `db:"created_at"`
		UpdatedAt        time.Time       `db:"updated_at"`
	}
)

func NewAwarenessModel(conn sqlx.SqlConn) AwarenessModel {
	return &customAwarenessModel{
		conn:  conn,
		table: "`awareness`",
	}
}

func (m *customAwarenessModel) FindEligible(ctx context.Context) ([]Awareness, error) {
	var resp []Awareness
	query := fmt.Sprintf("select %s from %s where `status` = 1 and `is_meta` = 0 order by `sort_order_global` asc, `awareness_id` asc", awarenessRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

func (m *customAwarenessModel) UpdateContent(ctx context.Context, id uint64, title, summary, details string) error {
	query := fmt.Sprintf("update %s set `point_title` = ?, `summary` = ?, `details` = ? where `awareness_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, title, nullAwarenessString(summary), nullAwarenessString(details), id)
	return err
}

func (m *customAwarenessModel) Disable(ctx context.Context, id uint64) error {
	query := fmt.Sprintf("update %s set `status` = 0 where `awareness_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

func (m *customAwarenessModel) withSession(session sqlx.Session) AwarenessModel {
	return NewAwarenessModel(sqlx.NewSqlConnFromSession(session))
}

func nullAwarenessString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}
