// Package remotestack provides migration management for PostgreSQL
package remotestack

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	"github.com/kubex-ecosystem/domus/internal/bootstrap"
	"github.com/kubex-ecosystem/domus/internal/module/kbx"
	"github.com/kubex-ecosystem/domus/internal/types"

	kbxGet "github.com/kubex-ecosystem/kbx/get"
	kbxTypes "github.com/kubex-ecosystem/kbx/types"
	logz "github.com/kubex-ecosystem/logz"

	_ "github.com/lib/pq"
)

// MigrationResult represents the result of a migration execution
type MigrationResult struct {
	FileName        string
	TotalStatements int
	SuccessfulStmts int
	FailedStmts     int
	Errors          []StatementError
	Duration        time.Duration
}

// StatementError represents an error in a specific SQL statement
type StatementError struct {
	Statement string
	Error     string
	Line      int
}

// MigrationManager handles database initialization and migrations with error recovery
type MigrationManager struct {
	count    int
	attempts int
	dsn      string
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(dsn string, logger *logz.LoggerZ) *MigrationManager {
	return &MigrationManager{
		count:    0,
		attempts: 5,
		dsn:      dsn,
	}
}

// WaitForPostgres waits for PostgreSQL to be ready with exponential backoff
func (m *MigrationManager) WaitForPostgres(ctx context.Context, maxWait time.Duration) error {
	deadline := time.Now().Add(maxWait)
	m.count = 0

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if m.count >= m.attempts {
			break
		}
		// Increment attempt and count at the start of the loop
		// this ensures they are updated for each try
		// and represent the current attempt number
		m.count++

		db := kbxGet.ValErrOr(
			m.OpenCurrentPGConn,
			kbxGet.ValErrOr(
				m.OpenCannonicalPGConn,
				nil,
			),
		)
		if db == nil {
			logz.Debugf("Attempt %d: Connection failed", m.count)
			time.Sleep(time.Duration(m.count) * time.Second)
			continue
		} else {
			logz.Debugf("Attempt %d: Connection established", m.count)
		}

		ctxConnection, cancelConnection := context.WithTimeout(ctx, 5*time.Second)
		defer func(db *sql.DB, cancel context.CancelFunc) {
			if cancelConnection != nil {
				cancelConnection()
			}
			if db != nil {
				_ = db.Close()
			}
		}(db, cancelConnection)

		if err := db.PingContext(ctxConnection); err != nil {

			logz.Debugf("Attempt %d: Ping failed: %v", m.count, err)
			time.Sleep(time.Duration(m.count) * time.Second)
			continue
		} else {
			logz.Debugf("Attempt %d: Ping successful", m.count)
			return nil
		}
	}

	return logz.Errorf("PostgreSQL not ready after %v (tried %d times)", maxWait, m.count-1)
}

func (m *MigrationManager) OpenCurrentPGConn() (*sql.DB, error) {
	return sql.Open("postgres", m.dsn)
}

func (m *MigrationManager) OpenCannonicalPGConn() (*sql.DB, error) {
	cDsn, err := getCanonicalPostgresDSN(m.dsn)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("postgres", cDsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// RunMigrations executes all SQL files in order with error recovery
func (m *MigrationManager) RunMigrations(ctx context.Context, migrationInfo *kbx.MigrationInfo) ([]MigrationResult, error) {
	db, err := sql.Open("postgres", m.dsn)
	if err != nil {
		return nil, logz.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	migrations, err := loadMigrationOrder()
	if err != nil {
		logz.Warnf("could not load bootstrap manifest: %v. Falling back to legacy migrations.", err)
		migrations = []string{"001_init.sql", "002_hardening.sql", "003_create_user_invitations.sql"}
	}
	results := make([]MigrationResult, 0, len(migrations))

	logz.Info("Starting PostgreSQL migrations with error recovery...")

	for _, filename := range migrations {
		result := m.executeSQLFileWithRecovery(ctx, db, filename)
		results = append(results, result)

		// Log summary for this file
		if result.FailedStmts == 0 {
			logz.Successf("%s: %d/%d statements executed successfully (%.2fs)",
				filename, result.SuccessfulStmts, result.TotalStatements, result.Duration.Seconds())
		} else {
			logz.Warnf("%s: %d/%d statements succeeded, %d failed (%.2fs)",
				filename, result.SuccessfulStmts, result.TotalStatements, result.FailedStmts, result.Duration.Seconds()) // Log first few errors for debugging
			for i, err := range result.Errors {
				if i >= 3 { // Limit error logging
					logz.Warnf("... and %d more errors", len(result.Errors)-i)
					break
				}
				logz.Errorf("   Line %d: %s", err.Line, err.Error)
			}
		}
	}

	// Overall summary
	totalSuccess := 0
	totalFailed := 0
	for _, r := range results {
		totalSuccess += r.SuccessfulStmts
		totalFailed += r.FailedStmts
	}

	if totalFailed == 0 {
		logz.Successf("All migrations completed successfully! (%d statements)", totalSuccess)
	} else {
		logz.Warnf("Migrations completed with partial success: %d succeeded, %d failed", totalSuccess, totalFailed)
	}

	return results, nil
}

func (m *MigrationManager) EndpointRedacted(ctx context.Context, conn *types.DBConnection) (string, error) {
	return logz.Sprintf("redacted(%s)", conn.Config.Config.DSN), nil
}

// executeSQLFileWithRecovery executes a single SQL file with statement-level error recovery
func (m *MigrationManager) executeSQLFileWithRecovery(ctx context.Context, db *sql.DB, filename string) MigrationResult {
	start := time.Now()
	result := MigrationResult{
		FileName: filename,
		Errors:   make([]StatementError, 0),
	}

	// Read file content (embed.FS is already rooted at assets/)
	content, err := bootstrap.MigrationFiles.ReadFile(filepath.Join("embedded", filename))
	if err != nil {
		result.Errors = append(result.Errors, StatementError{
			Statement: "",
			Error:     logz.Errorf("Failed to read file: %v", err).Error(),
			Line:      0,
		})
		result.Duration = time.Since(start)
		return result
	}

	// Pre-process: Remove psql meta-commands (lines starting with backslash)
	cleanedContent := m.removeMetaCommands(string(content))

	// Parse SQL statements
	statements := m.parseSQL(cleanedContent)
	result.TotalStatements = len(statements)

	logz.Debugf("Executing %s (%d statements)...", filename, len(statements))

	// Execute each statement individually
	for i, stmt := range statements {
		if strings.TrimSpace(stmt.SQL) == "" {
			continue
		}

		// Execute statement with timeout
		stmtCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		_, err := db.ExecContext(stmtCtx, stmt.SQL)
		cancel()

		if err != nil {
			result.FailedStmts++
			result.Errors = append(result.Errors, StatementError{
				Statement: stmt.SQL,
				Error:     err.Error(),
				Line:      stmt.Line,
			})

			// Log individual statement error (debug level to avoid spam)
			logz.Debugf("Statement %d failed: %v", i+1, err)
		} else {
			result.SuccessfulStmts++
			logz.Debugf("Statement %d executed", i+1)
		}
	}

	result.Duration = time.Since(start)
	return result
}

func getCanonicalPostgresDSN(dsn string) (string, error) {
	// por postgres, porque containers novos ou PGs totalmente sem
	// inicialização, não possuem o database alvo ainda criado, e NUNCA irão conectar.
	dsnKV, err := kbxTypes.ParseDSN(dsn)
	if err != nil {
		return "", logz.Errorf("failed to parse DSN: %v", err)
	}
	dsnKV.DBName = "postgres"
	return dsnKV.ToURL(), nil
}

// SQLStatement represents a parsed SQL statement with line information
type SQLStatement struct {
	SQL  string
	Line int
}

type manifest struct {
	ExecutionOrder []struct {
		File string `json:"file"`
	} `json:"execution_order"`
}

func loadMigrationOrder() ([]string, error) {
	data, err := bootstrap.MigrationFiles.ReadFile(filepath.Join("embedded", "bootstrap.manifest.json"))
	if err != nil {
		return nil, err
	}

	var mf manifest
	if err := json.Unmarshal(data, &mf); err != nil {
		return nil, err
	}

	if len(mf.ExecutionOrder) == 0 {
		return nil, logz.Errorf("manifest execution_order is empty")
	}

	files := make([]string, 0, len(mf.ExecutionOrder))
	for _, step := range mf.ExecutionOrder {
		if strings.TrimSpace(step.File) == "" {
			continue
		}
		files = append(files, step.File)
	}
	if len(files) == 0 {
		return nil, logz.Errorf("manifest has no valid files")
	}
	return files, nil
}

// removeMetaCommands removes psql meta-commands (lines starting with backslash) from SQL content.
// Examples: \echo, \set, \timing, \connect, etc.
// These commands are psql-specific and not valid SQL for database/sql driver.
func (m *MigrationManager) removeMetaCommands(content string) string {
	var result strings.Builder
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip lines that start with backslash (psql meta-commands)
		if strings.HasPrefix(trimmed, "\\") {
			// Replace with empty line to preserve line numbers for error reporting
			result.WriteString("\n")
			continue
		}
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// parseSQL splits SQL content into individual statements, preserving line numbers
func (m *MigrationManager) parseSQL(content string) []SQLStatement {
	var stmts []SQLStatement
	var b strings.Builder
	line := 1
	stmtStartLine := 1

	// state
	var inLineComment bool
	var inBlockComment bool
	var dollarTag string

	runes := []rune(content)
	i := 0
	for i < len(runes) {
		r := runes[i]

		// track line numbers
		if r == '\n' {
			line++
		}

		// handle end of line comment
		if inLineComment {
			b.WriteRune(r)
			if r == '\n' {
				inLineComment = false
			}
			i++
			continue
		}

		// handle end of block comment
		if inBlockComment {
			b.WriteRune(r)
			if r == '*' && i+1 < len(runes) && runes[i+1] == '/' {
				b.WriteRune(runes[i+1])
				i += 2
				inBlockComment = false
				continue
			}
			i++
			continue
		}

		// handle dollar-quote content
		if dollarTag != "" {
			b.WriteRune(r)
			tagRunes := []rune(dollarTag)
			if r == tagRunes[0] && i+len(tagRunes) <= len(runes) {
				match := true
				for k := 0; k < len(tagRunes); k++ {
					if runes[i+k] != tagRunes[k] {
						match = false
						break
					}
				}
				if match {
					for k := 1; k < len(tagRunes); k++ {
						b.WriteRune(runes[i+k])
					}
					i += len(tagRunes)
					dollarTag = ""
					continue
				}
			}
			i++
			continue
		}

		// detect start of line comment --
		if r == '-' && i+1 < len(runes) && runes[i+1] == '-' {
			inLineComment = true
			b.WriteRune(r)
			b.WriteRune(runes[i+1])
			i += 2
			continue
		}

		// detect start of block comment /*
		if r == '/' && i+1 < len(runes) && runes[i+1] == '*' {
			inBlockComment = true
			b.WriteRune(r)
			b.WriteRune(runes[i+1])
			i += 2
			continue
		}

		// detect dollar quote start $tag$
		if r == '$' {
			j := i + 1
			for j < len(runes) && runes[j] != '$' && ((runes[j] >= 'a' && runes[j] <= 'z') || (runes[j] >= 'A' && runes[j] <= 'Z') || (runes[j] >= '0' && runes[j] <= '9') || runes[j] == '_') {
				j++
			}
			if j < len(runes) && runes[j] == '$' {
				tagRunes := runes[i : j+1]
				dollarTag = string(tagRunes)
				for k := 0; k < len(tagRunes); k++ {
					b.WriteRune(tagRunes[k])
				}
				i = j + 1
				continue
			}
		}

		// single-quote start
		if r == '\'' {
			b.WriteRune(r)
			i++
			for i < len(runes) {
				b.WriteRune(runes[i])
				if runes[i] == '\'' {
					if i+1 < len(runes) && runes[i+1] == '\'' {
						b.WriteRune(runes[i+1])
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}
			continue
		}

		// double-quote start (identifiers)
		if r == '"' {
			b.WriteRune(r)
			i++
			for i < len(runes) {
				b.WriteRune(runes[i])
				if runes[i] == '"' {
					if i+1 < len(runes) && runes[i+1] == '"' {
						b.WriteRune(runes[i+1])
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}
			continue
		}

		// semicolon at top level -> end statement
		if r == ';' {
			b.WriteRune(r)
			stmt := strings.TrimSpace(b.String())
			if stmt != "" && stmt != ";" {
				stmts = append(stmts, SQLStatement{
					SQL:  stmt,
					Line: stmtStartLine,
				})
			}
			b.Reset()
			i++
			for i < len(runes) && runes[i] == '\n' {
				i++
				line++
			}
			stmtStartLine = line
			continue
		}

		if b.Len() == 0 {
			stmtStartLine = line
		}
		b.WriteRune(r)
		i++
	}

	if strings.TrimSpace(b.String()) != "" {
		stmts = append(stmts, SQLStatement{
			SQL:  strings.TrimSpace(b.String()),
			Line: stmtStartLine,
		})
	}

	return stmts
}

// SchemaExists checks if the required schema is already initialized
func (m *MigrationManager) SchemaExists() (bool, error) {
	db, err := sql.Open("postgres", m.dsn)
	if err != nil {
		return false, err
	}
	defer db.Close()

	// Check for tables in public schema (more reliable than just counting)
	var count int
	query := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_type = 'BASE TABLE'
	`

	err = db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, err
	}

	// Also check for extensions (our init script creates extensions)
	var extCount int
	extQuery := `
		SELECT COUNT(*)
		FROM pg_extension
		WHERE extname IN ('uuid-ossp', 'pgcrypto', 'pg_trgm')
	`

	err = db.QueryRow(extQuery).Scan(&extCount)
	if err != nil {
		logz.Debugf("Could not check extensions: %v", err)
		extCount = 0
	}

	// Consider schema exists if we have tables OR our extensions
	exists := count > 0 || extCount >= 2

	if exists {
		logz.Debugf("Schema check: %d tables, %d extensions found", count, extCount)
	}

	return exists, nil
}

// MissingTables returns which required base tables are missing from the target schema.
func (m *MigrationManager) MissingTables(schema string, tables ...string) ([]string, error) {
	db, err := sql.Open("postgres", m.dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	missing := make([]string, 0)
	for _, table := range tables {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.tables
				WHERE table_schema = $1
				  AND table_name = $2
				  AND table_type = 'BASE TABLE'
			)
		`
		if err := db.QueryRow(query, schema, table).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			missing = append(missing, table)
		}
	}

	return missing, nil
}

// ValidateConnection tests the database connection
func (m *MigrationManager) ValidateConnection() error {
	db, err := sql.Open("postgres", m.dsn)
	if err != nil {
		return logz.Errorf("failed to connect: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return logz.Errorf("failed to ping database: %v", err)
	}

	return nil
}
