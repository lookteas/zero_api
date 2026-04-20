package logic

import (
	"context"
	"strconv"
	"strings"
)

type currentAdminIDKey struct{}

func WithCurrentAdminID(ctx context.Context, adminID uint64) context.Context {
	if adminID == 0 {
		return ctx
	}

	return context.WithValue(ctx, currentAdminIDKey{}, adminID)
}

func ParseCurrentAdminID(value string) uint64 {
	adminID, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil || adminID == 0 {
		return 0
	}

	return adminID
}

func currentAdminID(ctx context.Context) uint64 {
	adminID, ok := ctx.Value(currentAdminIDKey{}).(uint64)
	if !ok || adminID == 0 {
		return 0
	}

	return adminID
}
