package logic

import (
	"errors"
	"strings"

	"api/internal/types"
)

type adminUserQuery struct {
	Page     int64
	PageSize int64
	Status   string
	Keyword  string
}

type adminUserTopicCountRow struct {
	UserId     uint64 `db:"user_id"`
	TopicCount int64  `db:"topic_count"`
}

type adminUserProfileInput struct {
	Nickname string
	Email    string
	Mobile   string
	Avatar   string
	Status   int64
}

func normalizeAdminUserQuery(req types.AdminUserQueryReq) adminUserQuery {
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	status := strings.TrimSpace(req.Status)
	if status != "" && status != "0" && status != "1" {
		status = ""
	}

	return adminUserQuery{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		Keyword:  strings.TrimSpace(req.Keyword),
	}
}

func normalizeAdminUserProfileInput(req *types.AdminUpdateUserReq) (adminUserProfileInput, error) {
	if req == nil {
		return adminUserProfileInput{}, errors.New("user profile is required")
	}

	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		return adminUserProfileInput{}, errors.New("nickname is required")
	}

	email := strings.TrimSpace(req.Email)
	if email != "" && !isEmail(email) {
		return adminUserProfileInput{}, errors.New("email format is invalid")
	}

	mobile := strings.TrimSpace(req.Mobile)
	if mobile != "" && !isMobile(mobile) {
		return adminUserProfileInput{}, errors.New("mobile format is invalid")
	}

	status := req.Status
	if status != 0 && status != 1 {
		return adminUserProfileInput{}, errors.New("status must be 0 or 1")
	}

	return adminUserProfileInput{
		Nickname: nickname,
		Email:    email,
		Mobile:   mobile,
		Avatar:   strings.TrimSpace(req.Avatar),
		Status:   status,
	}, nil
}

func adminUserIDs(users []types.AdminUserInfo) []uint64 {
	ids := make([]uint64, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.Id)
	}
	return ids
}

func applyAdminUserTopicCounts(users []types.AdminUserInfo, counts map[uint64]int64) {
	for index := range users {
		users[index].TopicCount = counts[users[index].Id]
	}
}

func buildQuestionPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}

	parts := make([]string, count)
	for index := range parts {
		parts[index] = "?"
	}
	return strings.Join(parts, ",")
}
