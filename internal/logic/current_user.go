package logic

import (
	"context"
	"strconv"
	"strings"
)

const defaultCurrentUserID uint64 = 1

type currentUserIDKey struct{}

func WithCurrentUserID(ctx context.Context, userID uint64) context.Context {
	if userID == 0 {
		return ctx
	}

	return context.WithValue(ctx, currentUserIDKey{}, userID)
}

func ParseCurrentUserID(value string) uint64 {
	userID, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil || userID == 0 {
		return 0
	}

	return userID
}

func currentUserID(ctx context.Context) uint64 {
	userID, ok := ctx.Value(currentUserIDKey{}).(uint64)
	if !ok || userID == 0 {
		return defaultCurrentUserID
	}

	return userID
}
