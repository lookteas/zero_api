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
		FindOne(ctx context.Context, id uint64) (*Awareness, error)
		CreateMinimal(ctx context.Context, title, summary, details string) (*Awareness, error)
		MoveToPosition(ctx context.Context, id uint64, position int64) error
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

func (m *customAwarenessModel) FindOne(ctx context.Context, id uint64) (*Awareness, error) {
	var resp Awareness
	query := fmt.Sprintf("select %s from %s where `awareness_id` = ? limit 1", awarenessRows, m.table)
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

func (m *customAwarenessModel) CreateMinimal(ctx context.Context, title, summary, details string) (*Awareness, error) {
	query := fmt.Sprintf("insert into %s (`chapter_id`, `region_id`, `section_id`, `point_title`, `summary`, `details`, `status`, `is_meta`, `sort_order_global`) values (0, 0, 0, ?, ?, ?, 1, 0, 999999)", m.table)
	result, err := m.conn.ExecCtx(ctx, query, title, nullAwarenessString(summary), nullAwarenessString(details))
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return m.FindOne(ctx, uint64(id))
}

func (m *customAwarenessModel) MoveToPosition(ctx context.Context, id uint64, position int64) error {
	if position <= 0 {
		position = 1
	}

	points, err := m.FindEligible(ctx)
	if err != nil {
		return err
	}
	orderedIDs := make([]uint64, 0, len(points))
	found := false
	for _, point := range points {
		if point.AwarenessId == id {
			found = true
			continue
		}
		orderedIDs = append(orderedIDs, point.AwarenessId)
	}
	if !found {
		return ErrNotFound
	}

	insertAt := int(position - 1)
	if insertAt < 0 {
		insertAt = 0
	}
	if insertAt > len(orderedIDs) {
		insertAt = len(orderedIDs)
	}
	orderedIDs = append(orderedIDs, 0)
	copy(orderedIDs[insertAt+1:], orderedIDs[insertAt:])
	orderedIDs[insertAt] = id

	query := fmt.Sprintf("update %s set `sort_order_global` = ? where `awareness_id` = ?", m.table)
	for order, orderedID := range orderedIDs {
		if _, err = m.conn.ExecCtx(ctx, query, order+1, orderedID); err != nil {
			return err
		}
	}
	return nil
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
