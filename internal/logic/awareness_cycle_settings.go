package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"api/internal/svc"

	"github.com/go-sql-driver/mysql"
)

const (
	awarenessCycleStartDateKey   = "awareness_cycle_start_date"
	awarenessCycleRestDaysKey    = "awareness_cycle_rest_days"
	awarenessCyclePausedDatesKey = "awareness_cycle_paused_dates"
)

func getAwarenessCycleSettings(ctx context.Context, svcCtx *svc.ServiceContext) (time.Time, int, error) {
	startDate := parseAwarenessCycleStart(svcCtx.Config.AwarenessCycle.StartDate)
	restDays := svcCtx.Config.AwarenessCycle.RestDays
	if restDays <= 0 {
		restDays = defaultAwarenessCycleRestDays
	}
	if svcCtx.DB == nil {
		return startDate, restDays, nil
	}

	storedStart, err := getAppSetting(ctx, svcCtx, awarenessCycleStartDateKey)
	if err != nil {
		return time.Time{}, 0, err
	}
	if strings.TrimSpace(storedStart) != "" {
		parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(storedStart), time.Local)
		if err == nil {
			startDate = normalizeDate(parsed)
		}
	}

	storedRestDays, err := getAppSetting(ctx, svcCtx, awarenessCycleRestDaysKey)
	if err != nil {
		return time.Time{}, 0, err
	}
	if strings.TrimSpace(storedRestDays) != "" {
		parsed, err := strconv.Atoi(strings.TrimSpace(storedRestDays))
		if err == nil && parsed > 0 {
			restDays = parsed
		}
	}

	return startDate, restDays, nil
}

func updateAwarenessCycleSettings(ctx context.Context, svcCtx *svc.ServiceContext, startDate string, restDays int) error {
	parsedStartDate, parsedRestDays, err := validateAwarenessCycleSettings(startDate, restDays)
	if err != nil {
		return err
	}
	if svcCtx.DB == nil {
		return fmt.Errorf("database unavailable")
	}

	query := "insert into app_settings (setting_key, setting_value) values (?, ?) on duplicate key update setting_value = values(setting_value)"
	if _, err := svcCtx.DB.ExecCtx(ctx, query, awarenessCycleStartDateKey, parsedStartDate.Format("2006-01-02")); err != nil {
		return fmt.Errorf("save awareness cycle startDate: %w", err)
	}
	if _, err := svcCtx.DB.ExecCtx(ctx, query, awarenessCycleRestDaysKey, strconv.Itoa(parsedRestDays)); err != nil {
		return fmt.Errorf("save awareness cycle restDays: %w", err)
	}

	return nil
}

func validateAwarenessCycleSettings(startDate string, restDays int) (time.Time, int, error) {
	parsedStartDate, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(startDate), time.Local)
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("invalid startDate, expected YYYY-MM-DD")
	}
	if restDays <= 0 {
		return time.Time{}, 0, fmt.Errorf("invalid restDays, must be greater than 0")
	}

	return normalizeDate(parsedStartDate), restDays, nil
}

func getAppSetting(ctx context.Context, svcCtx *svc.ServiceContext, key string) (string, error) {
	var value string
	err := svcCtx.DB.QueryRowCtx(ctx, &value, "select setting_value from app_settings where setting_key = ? limit 1", key)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if isMySQLTableMissing(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("query app setting %s: %w", key, err)
	}

	return value, nil
}

func updateAwarenessCyclePausedDatesSetting(ctx context.Context, svcCtx *svc.ServiceContext, pausedDates []string) error {
	if svcCtx.DB == nil {
		return nil
	}
	encoded, err := json.Marshal(pausedDates)
	if err != nil {
		return err
	}
	query := "insert into app_settings (setting_key, setting_value) values (?, ?) on duplicate key update setting_value = values(setting_value)"
	_, err = svcCtx.DB.ExecCtx(ctx, query, awarenessCyclePausedDatesKey, string(encoded))
	return err
}

func isMySQLTableMissing(err error) bool {
	mysqlErr, ok := err.(*mysql.MySQLError)
	return ok && mysqlErr.Number == 1146
}
