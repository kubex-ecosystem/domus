package types

import (
	"github.com/kubex-ecosystem/logz"
)

// Common logz errors.
var (
	ErrTerminal     = logz.Errorf("pipeline already in terminal stage")
	ErrNotFound     = logz.Errorf("entry not found")
	ErrInvalid      = logz.Errorf("invalid entry")
	ErrClosed       = logz.Errorf("entry already closed")
	ErrTimeout      = logz.Errorf("entry processing timeout")
	ErrCanceled     = logz.Errorf("entry processing canceled")
	ErrAccess       = logz.Errorf("access denied")
	ErrExists       = logz.Errorf("entry already exists")
	ErrLocked       = logz.Errorf("entry is locked")
	ErrCorrupt      = logz.Errorf("entry is corrupt")
	ErrUnknown      = logz.Errorf("unknown error")
	ErrInvalidInput = logz.Errorf("invalid input")
	ErrConflict     = logz.Errorf("entry conflict")
)

// Domus domain errors.
var (
	ErrUnimplemented   = logz.Errorf("unimplemented")
	ErrIncompleteData  = logz.Errorf("incomplete data")
	ErrInvalidData     = logz.Errorf("invalid data")
	ErrInvalidConfig   = logz.Errorf("invalid configuration")
	ErrInvalidState    = logz.Errorf("invalid state")
	ErrInvalidQuery    = logz.Errorf("invalid query")
	ErrInvalidPipeline = logz.Errorf("invalid pipeline")
	ErrInvalidStage    = logz.Errorf("invalid stage")
	ErrInvalidEntry    = logz.Errorf("invalid entry")
)

// Operational errors.
var (
	ErrFieldRequired = logz.Errorf("field required")
	ErrFieldTooLong  = logz.Errorf("field too long")
	ErrFieldTooShort = logz.Errorf("field too short")
	ErrFieldInvalid  = logz.Errorf("field invalid")
	ErrFieldMissing  = logz.Errorf("field missing")
)

// Unreachable code.
var (
	ErrUnreachable    = logz.Errorf("unreachable code")
	ErrNotImplemented = logz.Errorf("not implemented")
	ErrNotSupported   = logz.Errorf("not supported")
	ErrNotConfigured  = logz.Errorf("not configured")
	ErrNotInitialized = logz.Errorf("not initialized")
	ErrNotReady       = logz.Errorf("not ready")
	ErrNotPersisted   = logz.Errorf("not persisted")
)
