package extensions

import (
	"context"
	"fmt"
	"sync"

	"github.com/kubex-ecosystem/domus/internal/execution"
	gl "github.com/kubex-ecosystem/logz"
)

// SchemaExtension defines a strict contract that any external module (e.g. GNyx, WMS)
// must fulfill to have its schema integrated and migrated by Domus.
type SchemaExtension interface {
	// ID returns the unique identifier for this extension (e.g., "gnyx_wms_schema").
	ID() string

	// Statements returns the ordered list of DDL statements to apply.
	Statements() []string

	// Dependencies returns the list of extension IDs that must be applied before this one.
	Dependencies() []string
}

// Manager holds the registry of available schema extensions.
type Manager struct {
	mu         sync.RWMutex
	extensions map[string]SchemaExtension
}

// NewManager creates a new SchemaExtension manager.
func NewManager() *Manager {
	return &Manager{
		extensions: make(map[string]SchemaExtension),
	}
}

// Register adds an extension to the registry.
func (m *Manager) Register(ext SchemaExtension) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.extensions[ext.ID()] = ext
}

// Apply runs the specified extension (and its dependencies recursively) using the provided executor.
// It tracks applied extensions in the database to ensure idempotency.
func (m *Manager) Apply(ctx context.Context, exec execution.Executor, extID string) error {
	m.mu.RLock()
	ext, exists := m.extensions[extID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("schema extension %q not found in registry", extID)
	}

	// First ensure the tracking table exists
	if err := m.ensureTrackingTable(ctx, exec); err != nil {
		return fmt.Errorf("failed to ensure extension tracking table: %w", err)
	}

	return m.applyRecursive(ctx, exec, ext, make(map[string]bool))
}

// ApplyAll runs all registered extensions in arbitrary order (dependencies are respected automatically).
func (m *Manager) ApplyAll(ctx context.Context, exec execution.Executor) error {
	m.mu.RLock()
	// Get a copy of IDs to avoid holding lock during apply
	var ids []string
	for id := range m.extensions {
		ids = append(ids, id)
	}
	m.mu.RUnlock()

	for _, id := range ids {
		if err := m.Apply(ctx, exec, id); err != nil {
			return fmt.Errorf("failed to apply extension %q: %w", id, err)
		}
	}
	return nil
}

func (m *Manager) applyRecursive(ctx context.Context, exec execution.Executor, ext SchemaExtension, visiting map[string]bool) error {
	if visiting[ext.ID()] {
		return fmt.Errorf("circular dependency detected involving extension %q", ext.ID())
	}

	visiting[ext.ID()] = true
	defer func() { visiting[ext.ID()] = false }()

	// Apply dependencies first
	for _, depID := range ext.Dependencies() {
		m.mu.RLock()
		depExt, exists := m.extensions[depID]
		m.mu.RUnlock()
		if !exists {
			return fmt.Errorf("dependency %q for extension %q not found", depID, ext.ID())
		}
		if err := m.applyRecursive(ctx, exec, depExt, visiting); err != nil {
			return err
		}
	}

	// Check if already applied
	applied, err := m.isApplied(ctx, exec, ext.ID())
	if err != nil {
		return fmt.Errorf("failed to check status of extension %q: %w", ext.ID(), err)
	}
	if applied {
		gl.Debugf("Extension %q already applied, skipping.", ext.ID())
		return nil
	}

	gl.Infof("Applying schema extension: %s", ext.ID())

	// Apply statements within a transaction if possible, or sequentially
	// Since DDLs might need to be run outside a transaction in some DBs, we run them sequentially.
	for i, stmt := range ext.Statements() {
		if _, err := exec.PG().Exec(ctx, stmt); err != nil {
			return fmt.Errorf("failed to apply statement %d of extension %q: %w", i, ext.ID(), err)
		}
	}

	// Mark as applied
	if err := m.markApplied(ctx, exec, ext.ID()); err != nil {
		return fmt.Errorf("failed to mark extension %q as applied: %w", ext.ID(), err)
	}

	gl.Infof("Schema extension %q applied successfully.", ext.ID())
	return nil
}

func (m *Manager) ensureTrackingTable(ctx context.Context, exec execution.Executor) error {
	stmt := `
		CREATE TABLE IF NOT EXISTS domus_schema_extensions (
			id TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`
	_, err := exec.PG().Exec(ctx, stmt)
	return err
}

func (m *Manager) isApplied(ctx context.Context, exec execution.Executor, extID string) (bool, error) {
	var count int
	err := exec.PG().QueryRow(ctx, "SELECT COUNT(1) FROM domus_schema_extensions WHERE id = $1", extID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (m *Manager) markApplied(ctx context.Context, exec execution.Executor, extID string) error {
	_, err := exec.PG().Exec(ctx, "INSERT INTO domus_schema_extensions (id) VALUES ($1) ON CONFLICT (id) DO NOTHING", extID)
	return err
}
