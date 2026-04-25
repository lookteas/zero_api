package main

import (
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

	sqlDir := filepath.Clean(filepath.Join(filepath.Dir(*configFile), "../../../docs/sql"))
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
	conn := sqlx.NewMysql(bootstrapDSN)
	for _, sqlFile := range sqlFiles {
		content, readErr := os.ReadFile(sqlFile)
		if readErr != nil {
			panic(readErr)
		}

		statements := splitSQLStatements(string(content))
		for _, stmt := range statements {
			if strings.TrimSpace(stmt) == "" {
				continue
			}

			if _, err = conn.Exec(stmt); err != nil {
				panic(fmt.Errorf("execute sql failed: %w\nfile: %s\nstatement: %s", err, sqlFile, stmt))
			}
		}
	}

	fmt.Println("database initialized successfully")
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
