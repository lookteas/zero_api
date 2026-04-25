package logic

import (
	"context"
	"fmt"
	"strings"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AdminListUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListUsersLogic {
	return &AdminListUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListUsersLogic) AdminListUsers(req *types.AdminUserQueryReq) (resp *types.AdminUserListResp, err error) {
	if err = requireAdminUser(l.ctx); err != nil {
		return nil, err
	}

	query := normalizeAdminUserQuery(*req)
	if l.svcCtx.DB == nil || l.svcCtx.UsersModel == nil {
		return &types.AdminUserListResp{
			Code:    0,
			Message: "ok",
			Data: types.AdminUserListData{
				List:       []types.AdminUserInfo{},
				Pagination: types.Pagination{Page: query.Page, PageSize: query.PageSize, Total: 0},
				Summary:    types.AdminUserSummary{},
			},
		}, nil
	}

	where, args := buildAdminUsersWhere(query)

	countQuery := "select count(1) from users" + where
	var total int64
	if err = l.svcCtx.DB.QueryRowCtx(l.ctx, &total, countQuery, args...); err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	listQuery := "select id, account, email, mobile, password_hash, nickname, avatar, status, last_login_at, created_at, updated_at from users" + where + " order by created_at desc, id desc limit ? offset ?"
	listArgs := append(append([]any{}, args...), query.PageSize, (query.Page-1)*query.PageSize)

	var items []model.Users
	if err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &items, listQuery, listArgs...); err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}

	summary, err := queryAdminUsersSummary(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, err
	}

	list := make([]types.AdminUserInfo, 0, len(items))
	for _, item := range items {
		copied := item
		list = append(list, adminUserToInfo(&copied))
	}
	topicCounts, err := queryAdminUserTopicCounts(l.ctx, l.svcCtx.DB, adminUserIDs(list))
	if err != nil {
		return nil, err
	}
	applyAdminUserTopicCounts(list, topicCounts)

	return &types.AdminUserListResp{
		Code:    0,
		Message: "ok",
		Data: types.AdminUserListData{
			List:       list,
			Pagination: types.Pagination{Page: query.Page, PageSize: query.PageSize, Total: total},
			Summary:    summary,
		},
	}, nil
}

func queryAdminUserTopicCounts(ctx context.Context, db sqlx.SqlConn, userIDs []uint64) (map[uint64]int64, error) {
	counts := make(map[uint64]int64, len(userIDs))
	if len(userIDs) == 0 {
		return counts, nil
	}

	args := make([]any, 0, len(userIDs))
	for _, userID := range userIDs {
		args = append(args, userID)
	}

	query := "select user_id, count(1) as topic_count from daily_tasks where user_id in (" + buildQuestionPlaceholders(len(userIDs)) + ") group by user_id"
	var rows []adminUserTopicCountRow
	if err := db.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("query user topic counts: %w", err)
	}

	for _, row := range rows {
		counts[row.UserId] = row.TopicCount
	}
	return counts, nil
}

func buildAdminUsersWhere(query adminUserQuery) (string, []any) {
	clauses := make([]string, 0, 2)
	args := make([]any, 0, 5)

	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		clauses = append(clauses, "(`account` like ? or `nickname` like ? or `email` like ? or `mobile` like ?)")
		args = append(args, keyword, keyword, keyword, keyword)
	}
	if query.Status != "" {
		clauses = append(clauses, "`status` = ?")
		args = append(args, query.Status)
	}

	if len(clauses) == 0 {
		return "", args
	}

	return " where " + strings.Join(clauses, " and "), args
}

func queryAdminUsersSummary(ctx context.Context, db sqlx.SqlConn) (types.AdminUserSummary, error) {
	type summaryRow struct {
		Total      int64 `db:"total"`
		Active     int64 `db:"active"`
		WithEmail  int64 `db:"with_email"`
		WithMobile int64 `db:"with_mobile"`
	}

	var row summaryRow
	query := "select count(1) as total, ifnull(sum(case when status = 1 then 1 else 0 end), 0) as active, ifnull(sum(case when email is not null and email <> '' then 1 else 0 end), 0) as with_email, ifnull(sum(case when mobile is not null and mobile <> '' then 1 else 0 end), 0) as with_mobile from users"
	if err := db.QueryRowCtx(ctx, &row, query); err != nil {
		return types.AdminUserSummary{}, fmt.Errorf("query user summary: %w", err)
	}

	return types.AdminUserSummary{
		Total:      row.Total,
		Active:     row.Active,
		WithEmail:  row.WithEmail,
		WithMobile: row.WithMobile,
	}, nil
}

func adminUserToInfo(item *model.Users) types.AdminUserInfo {
	lastLoginAt := ""
	if item.LastLoginAt.Valid {
		lastLoginAt = item.LastLoginAt.Time.Format("2006-01-02 15:04:05")
	}

	email := ""
	if item.Email.Valid {
		email = item.Email.String
	}

	mobile := ""
	if item.Mobile.Valid {
		mobile = item.Mobile.String
	}

	return types.AdminUserInfo{
		Id:          item.Id,
		Account:     item.Account,
		Email:       email,
		Mobile:      mobile,
		Nickname:    item.Nickname,
		Avatar:      item.Avatar,
		Status:      int64(item.Status),
		CreatedAt:   item.CreatedAt.Format("2006-01-02 15:04:05"),
		LastLoginAt: lastLoginAt,
	}
}
