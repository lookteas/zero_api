package logic

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"api/internal/config"
	"api/internal/svc"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type missingTableAppSettingsDB struct {
	sqlx.SqlConn
}

func (db missingTableAppSettingsDB) QueryRowCtx(_ context.Context, _ any, _ string, _ ...any) error {
	return &mysql.MySQLError{Number: 1146, Message: "Table 'zero_app.app_settings' doesn't exist"}
}

type storedAppSettingsDB struct {
	sqlx.SqlConn
	settings map[string]string
}

func (db storedAppSettingsDB) QueryRowCtx(_ context.Context, v any, _ string, args ...any) error {
	if len(args) == 0 {
		return sql.ErrNoRows
	}
	value, ok := db.settings[args[0].(string)]
	if !ok {
		return sql.ErrNoRows
	}

	target := reflect.ValueOf(v)
	if target.Kind() == reflect.Pointer && !target.IsNil() {
		elem := target.Elem()
		if elem.Kind() == reflect.String {
			elem.SetString(value)
		}
	}
	return nil
}

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

func TestAwarenessCycleSettingsDefaultWhenSettingsTableMissing(t *testing.T) {
	t.Parallel()

	startDate, restDays, err := getAwarenessCycleSettings(context.Background(), &svc.ServiceContext{
		Config: config.Config{
			AwarenessCycle: config.AwarenessCycleConf{
				StartDate: "2026-06-01",
				RestDays:  5,
			},
		},
		DB: missingTableAppSettingsDB{},
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

func TestAwarenessCycleSettingsUsesStoredValues(t *testing.T) {
	t.Parallel()

	startDate, restDays, err := getAwarenessCycleSettings(context.Background(), &svc.ServiceContext{
		Config: config.Config{
			AwarenessCycle: config.AwarenessCycleConf{
				StartDate: "2026-06-01",
				RestDays:  5,
			},
		},
		DB: storedAppSettingsDB{settings: map[string]string{
			awarenessCycleStartDateKey: "2026-07-01",
			awarenessCycleRestDaysKey:  "9",
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := startDate.Format("2006-01-02"), "2026-07-01"; got != want {
		t.Fatalf("expected stored start date %s, got %s", want, got)
	}
	if restDays != 9 {
		t.Fatalf("expected stored rest days 9, got %d", restDays)
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
