package logic

import (
	"context"
	"strings"
	"testing"

	"api/internal/config"
	"api/internal/svc"
)

func TestAwarenessCycleSettingsDefaultWithoutDB(t *testing.T) {
	t.Parallel()

	startDate, restDays, err := getAwarenessCycleSettings(context.Background(), &svc.ServiceContext{
		Config: config.Config{
			AwarenessCycle: config.AwarenessCycleConf{
				StartDate: "2026-06-01",
				RestDays:  5,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := startDate.Format("2006-01-02"), "2026-06-01"; got != want {
		t.Fatalf("expected start date %s, got %s", want, got)
	}
	if restDays != 5 {
		t.Fatalf("expected rest days 5, got %d", restDays)
	}
}

func TestValidateAwarenessCycleSettingsRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	if _, _, err := validateAwarenessCycleSettings("bad-date", 7); err == nil || !strings.Contains(err.Error(), "startDate") {
		t.Fatalf("expected startDate validation error, got %v", err)
	}
	if _, _, err := validateAwarenessCycleSettings("2026-06-01", 0); err == nil || !strings.Contains(err.Error(), "restDays") {
		t.Fatalf("expected restDays validation error, got %v", err)
	}
}
