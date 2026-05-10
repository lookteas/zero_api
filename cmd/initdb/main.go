package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"api/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var configFile = flag.String("f", "../../etc/zero-api.yaml", "config file")

func main() {
	flag.Parse()

	var c config.Config
	config.MustLoad(*configFile, &c)

	if c.Mysql.DataSource == "" {
		panic("mysql datasource is empty")
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	sqlDir := resolveSQLDir(*configFile, cwd)
	entries, err := os.ReadDir(sqlDir)
	if err != nil {
		panic(err)
	}

	sqlFiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		sqlFiles = append(sqlFiles, filepath.Join(sqlDir, entry.Name()))
	}
	sort.Strings(sqlFiles)

	bootstrapDSN := stripDatabaseFromDSN(c.Mysql.DataSource)
	rawDB, err := openRawDB(bootstrapDSN)
	if err != nil {
		panic(err)
	}
	defer rawDB.Close()

	if err = runSQLFiles(context.Background(), rawDB, sqlFiles); err != nil {
		panic(err)
	}

	fmt.Println("database initialized successfully")
}

func resolveSQLDir(configFile, cwd string) string {
	localSQLDir := filepath.Join(cwd, "docs", "sql")
	if stat, err := os.Stat(localSQLDir); err == nil && stat.IsDir() {
		return localSQLDir
	}

	return filepath.Clean(filepath.Join(filepath.Dir(configFile), "../../../docs/sql"))
}

func openRawDB(dsn string) (*sql.DB, error) {
	conn := sqlx.NewMysql(dsn)
	return conn.RawDB()
}

func runSQLFiles(ctx context.Context, db *sql.DB, sqlFiles []string) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	for _, sqlFile := range sqlFiles {
		content, readErr := os.ReadFile(sqlFile)
		if readErr != nil {
			return readErr
		}

		statements := splitSQLStatements(string(content))
		for _, stmt := range statements {
			if strings.TrimSpace(stmt) == "" {
				continue
			}

			if _, err = conn.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("execute sql failed: %w\nfile: %s\nstatement: %s", err, sqlFile, stmt)
			}
		}
	}

	return nil
}

func stripDatabaseFromDSN(dsn string) string {
	withoutParams := dsn
	if index := strings.Index(withoutParams, "?"); index >= 0 {
		withoutParams = withoutParams[:index]
	}

	re := regexp.MustCompile(`/[^/?]*$`)
	return re.ReplaceAllString(withoutParams, "/")
}

func splitSQLStatements(input string) []string {
	input = strings.TrimPrefix(input, "\uFEFF")
	parts := strings.Split(input, ";")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt == "" {
			continue
		}
		result = append(result, stmt)
	}

	return result
}
