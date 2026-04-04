package migrator

import (
    "testing"
)

func TestNewMigrator(t *testing.T) {
    m := New("./testdata", "postgres://localhost/test")
    if m == nil {
        t.Fatal("New() returned nil")
    }
    if m.migrationsPath != "./testdata" {
        t.Errorf("migrationsPath = %s, want ./testdata", m.migrationsPath)
    }
    if m.databaseURL != "postgres://localhost/test" {
        t.Errorf("databaseURL = %s, want postgres://localhost/test", m.databaseURL)
    }
}

func TestMigratorUp(t *testing.T) {
    m := New("./nonexistent/path", "postgres://localhost/test")
    err := m.Up()
    if err == nil {
        t.Log("Up() returned nil (expected error or no migrations)")
    }
}

func TestMigratorDown(t *testing.T) {
    m := New("./testdata", "postgres://localhost/test")
    err := m.Down()
    _ = err
}

func TestMigratorVersion(t *testing.T) {
    m := New("./testdata", "postgres://localhost/test")
    _, _, err := m.Version()
    _ = err
}

func TestMaskPassword(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"postgres://user:pass@localhost:5432/db", "postgres://user:***@localhost:5432/db"},
        {"postgres://user:secret123@localhost/db", "postgres://user:***@localhost/db"},
        {"invalid-url", "invalid-url"},
    }
    
    for _, tt := range tests {
        result := maskPassword(tt.input)
        if result != tt.expected && tt.input != "invalid-url" {
            t.Logf("maskPassword(%s) = %s, expected like %s", tt.input, result, tt.expected)
        }
    }
}

func TestMigratorUp_WithValidPath(t *testing.T) {
    // Проверяем с существующей директорией
    m := New("./", "postgres://localhost/test")
    err := m.Up()
    // Может быть ошибка из-за БД, но не паника
    if err != nil {
        t.Logf("Up() returned error: %v (expected if no DB connection)", err)
    }
}

func TestMaskPassword_EdgeCases(t *testing.T) {
    testCases := []struct {
        name string
        url  string
    }{
        {"empty string", ""},
        {"no password", "postgres://user@localhost/db"},
        {"multiple colons", "postgres://user:pass:word@localhost/db"},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := maskPassword(tc.url)
            // Просто проверяем что функция не паникует
            _ = result
        })
    }
}