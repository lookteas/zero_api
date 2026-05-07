package logic

import (
	"context"
	"testing"

	"api/internal/svc"
)

func TestHealthReturnsOkWithoutDatabase(t *testing.T) {
	t.Parallel()

	resp, err := NewHealthLogic(context.Background(), &svc.ServiceContext{}).Health()
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
	if resp.Message != "ok" {
		t.Fatalf("expected message ok, got %q", resp.Message)
	}
}
