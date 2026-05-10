package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func TestStripDatabaseFromDSN(t *testing.T) {
	input := "root:root123@tcp(127.0.0.1:3309)/zero_app?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
	want := "root:root123@tcp(127.0.0.1:3309)/"

	if got := stripDatabaseFromDSN(input); got != want {
		t.Fatalf("stripDatabaseFromDSN() = %q, want %q", got, want)
	}
}

func TestRunSQLFilesUsesSinglePhysicalConnection(t *testing.T) {
	resetInitdbTestDriver()

	dir := t.TempDir()
	first := filepath.Join(dir, "001.sql")
	second := filepath.Join(dir, "002.sql")
	if err := os.WriteFile(first, []byte("CREATE DATABASE IF NOT EXISTS zero_app; USE zero_app;"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("SET @value = 1; PREPARE stmt FROM 'SELECT 1'; EXECUTE stmt;"), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("initdbtest", "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(4)

	if err := runSQLFiles(context.Background(), db, []string{first, second}); err != nil {
		t.Fatal(err)
	}

	connIDs := initdbTestExecConnIDs()
	if len(connIDs) != 5 {
		t.Fatalf("expected 5 executed statements, got %d", len(connIDs))
	}
	for _, id := range connIDs[1:] {
		if id != connIDs[0] {
			t.Fatalf("expected all statements on connection %d, got sequence %v", connIDs[0], connIDs)
		}
	}
}

func TestResolveSQLDirPrefersWorktreeDocsSQL(t *testing.T) {
	cwd := t.TempDir()
	localSQLDir := filepath.Join(cwd, "docs", "sql")
	if err := os.MkdirAll(localSQLDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(localSQLDir, "001_init_schema.sql"), []byte("CREATE DATABASE zero_app;"), 0644); err != nil {
		t.Fatal(err)
	}

	configFile := filepath.Join(cwd, "etc", "zero-api.yaml")
	got := resolveSQLDir(configFile, cwd)
	want := localSQLDir
	if got != want {
		t.Fatalf("expected worktree docs/sql %q, got %q", want, got)
	}
}

func TestResolveSQLDirFallsBackWhenWorktreeDocsSQLLacksBaseline(t *testing.T) {
	workspace := t.TempDir()
	cwd := filepath.Join(workspace, "apps", "api", ".worktrees", "awareness-first-core")
	localSQLDir := filepath.Join(cwd, "docs", "sql")
	rootSQLDir := filepath.Join(workspace, "docs", "sql")
	if err := os.MkdirAll(localSQLDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(localSQLDir, "006_awareness_first_core.sql"), []byte("CREATE TABLE awareness;"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(rootSQLDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rootSQLDir, "001_init_schema.sql"), []byte("CREATE DATABASE zero_app;"), 0644); err != nil {
		t.Fatal(err)
	}

	configFile := filepath.Join(cwd, "etc", "zero-api.yaml")
	got := resolveSQLDir(configFile, cwd)
	if got != rootSQLDir {
		t.Fatalf("expected workspace docs/sql %q, got %q", rootSQLDir, got)
	}
}

func TestResolveSQLDirsIncludesWorkspaceAndLocalMigrations(t *testing.T) {
	workspace := t.TempDir()
	cwd := filepath.Join(workspace, "apps", "api")
	localSQLDir := filepath.Join(cwd, "docs", "sql")
	rootSQLDir := filepath.Join(workspace, "docs", "sql")
	if err := os.MkdirAll(localSQLDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(localSQLDir, "006_awareness_first_core.sql"), []byte("CREATE TABLE app_settings;"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(rootSQLDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rootSQLDir, "001_init_schema.sql"), []byte("CREATE DATABASE zero_app;"), 0644); err != nil {
		t.Fatal(err)
	}

	configFile := filepath.Join(cwd, "etc", "zero-api.yaml")
	got := resolveSQLDirs(configFile, cwd)
	want := []string{rootSQLDir, localSQLDir}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected sql dirs %v, got %v", want, got)
	}
}

func TestCollectSQLFilesKeepsDirectoryOrderAndSortsWithinEachDirectory(t *testing.T) {
	dir := t.TempDir()
	firstDir := filepath.Join(dir, "root")
	secondDir := filepath.Join(dir, "local")
	if err := os.MkdirAll(firstDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(secondDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		dir  string
		name string
	}{
		{firstDir, "002_second.sql"},
		{firstDir, "001_first.sql"},
		{secondDir, "006_local.sql"},
	} {
		if err := os.WriteFile(filepath.Join(item.dir, item.name), []byte("SELECT 1;"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := collectSQLFiles([]string{firstDir, secondDir})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		filepath.Join(firstDir, "001_first.sql"),
		filepath.Join(firstDir, "002_second.sql"),
		filepath.Join(secondDir, "006_local.sql"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected sql files %v, got %v", want, got)
	}
}

func TestResolveSQLDirFallsBackToWorkspaceDocsSQL(t *testing.T) {
	workspace := t.TempDir()
	cwd := filepath.Join(workspace, "apps", "api")
	if err := os.MkdirAll(filepath.Join(workspace, "docs", "sql"), 0755); err != nil {
		t.Fatal(err)
	}

	configFile := filepath.Join(cwd, "etc", "zero-api.yaml")
	got := resolveSQLDir(configFile, cwd)
	want := filepath.Join(workspace, "docs", "sql")
	if got != want {
		t.Fatalf("expected workspace docs/sql %q, got %q", want, got)
	}
}

var initdbTestState = struct {
	sync.Mutex
	nextID  int
	execIDs []int
}{}

func init() {
	sql.Register("initdbtest", initdbTestDriver{})
}

func resetInitdbTestDriver() {
	initdbTestState.Lock()
	defer initdbTestState.Unlock()
	initdbTestState.nextID = 0
	initdbTestState.execIDs = nil
}

func initdbTestExecConnIDs() []int {
	initdbTestState.Lock()
	defer initdbTestState.Unlock()
	ids := make([]int, len(initdbTestState.execIDs))
	copy(ids, initdbTestState.execIDs)
	return ids
}

type initdbTestDriver struct{}

func (initdbTestDriver) Open(string) (driver.Conn, error) {
	initdbTestState.Lock()
	defer initdbTestState.Unlock()
	initdbTestState.nextID++
	return &initdbTestConn{id: initdbTestState.nextID}, nil
}

type initdbTestConn struct {
	id int
}

func (c *initdbTestConn) Prepare(string) (driver.Stmt, error) {
	return nil, fmt.Errorf("prepare is not implemented")
}

func (c *initdbTestConn) Close() error {
	return nil
}

func (c *initdbTestConn) Begin() (driver.Tx, error) {
	return nil, fmt.Errorf("transactions are not implemented")
}

func (c *initdbTestConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	initdbTestState.Lock()
	defer initdbTestState.Unlock()
	initdbTestState.execIDs = append(initdbTestState.execIDs, c.id)
	return driver.RowsAffected(0), nil
}

var _ driver.ExecerContext = (*initdbTestConn)(nil)
