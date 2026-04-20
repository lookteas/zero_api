package logic

import (
	"context"
	"testing"
)

func TestRequireAdminUserRejectsMissingAdminSession(t *testing.T) {
	err := requireAdminUser(context.Background())
	if err == nil {
		t.Fatal("expected missing admin session error")
	}
}

func TestRequireAdminUserAcceptsAdminContext(t *testing.T) {
	ctx := WithCurrentAdminID(context.Background(), 3)
	err := requireAdminUser(ctx)
	if err != nil {
		t.Fatalf("expected active admin context to pass, got %v", err)
	}
}
