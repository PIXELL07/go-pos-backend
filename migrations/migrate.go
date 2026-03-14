package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "5432"),
		getenv("DB_USER", "postgres"),
		getenv("DB_PASSWORD", ""),
		getenv("DB_NAME", "pos_db"),
		getenv("DB_SSLMODE", "disable"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	// Create migrations tracking table
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		log.Fatalf("create schema_migrations: %v", err)
	}

	// Gather .sql files
	dir := "migrations"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	entries, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		log.Fatalf("glob: %v", err)
	}
	sort.Strings(entries)

	applied := 0
	for _, path := range entries {
		version := filepath.Base(path)

		// Check if already applied
		var count int
		_ = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version).Scan(&count)
		if count > 0 {
			log.Printf("  skip  %s (already applied)", version)
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("read %s: %v", path, err)
		}

		// Skip empty / comment-only files
		if strings.TrimSpace(string(content)) == "" {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("begin tx for %s: %v", path, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			log.Fatalf("exec %s: %v", path, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version) VALUES ($1)", version,
		); err != nil {
			_ = tx.Rollback()
			log.Fatalf("record migration %s: %v", path, err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("commit %s: %v", path, err)
		}

		log.Printf("  apply %s ✅", version)
		applied++
	}

	if applied == 0 {
		log.Println("Nothing to migrate — database is up to date.")
	} else {
		log.Printf("Applied %d migration(s) successfully.", applied)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
