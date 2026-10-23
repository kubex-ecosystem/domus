package types_test

import (
	"errors"
	"testing"

	ci "github.com/kubex-ecosystem/domus/internal/interfaces"
	"github.com/kubex-ecosystem/domus/internal/types"
)

// ─────────────────────────────────────────────────────────────────────────── //
// ValidationResult
// ─────────────────────────────────────────────────────────────────────────── //

func TestValidationResult_Valid(t *testing.T) {
	r := types.NewValidationResult(true, "all good", nil, nil)
	if !r.GetIsValid() {
		t.Error("expected IsValid=true")
	}
	if r.GetMessage() != "all good" {
		t.Errorf("expected message 'all good', got %q", r.GetMessage())
	}
	if r.GetError() != nil {
		t.Errorf("expected nil error, got %v", r.GetError())
	}
}

func TestValidationResult_Invalid_WithError(t *testing.T) {
	sentinel := errors.New("something broke")
	r := types.NewValidationResult(false, "failed", nil, sentinel)
	if r.GetIsValid() {
		t.Error("expected IsValid=false")
	}
	if !errors.Is(r.GetError(), sentinel) {
		t.Errorf("expected sentinel error, got %v", r.GetError())
	}
}

func TestValidationResult_String_Valid(t *testing.T) {
	r := types.NewValidationResult(true, "ok", nil, nil)
	s := r.String()
	if s == "" {
		t.Error("String() should not be empty for valid result")
	}
}

func TestValidationResult_String_Invalid(t *testing.T) {
	r := types.NewValidationResult(false, "bad input", nil, nil)
	s := r.String()
	if s == "" {
		t.Error("String() should not be empty for invalid result")
	}
}

func TestValidationResult_Metadata_SetAndGet(t *testing.T) {
	r := types.NewValidationResult(true, "ok", nil, nil)
	r.SetMetadata("field", "email")
	r.SetMetadata("count", 42)

	val, ok := r.GetMetadata("field")
	if !ok {
		t.Error("expected metadata 'field' to exist")
	}
	if val != "email" {
		t.Errorf("expected 'email', got %v", val)
	}

	val2, ok2 := r.GetMetadata("count")
	if !ok2 || val2 != 42 {
		t.Errorf("expected count=42, got ok=%v val=%v", ok2, val2)
	}
}

func TestValidationResult_Metadata_MissingKey(t *testing.T) {
	r := types.NewValidationResult(true, "ok", nil, nil)
	_, ok := r.GetMetadata("nonexistent")
	if ok {
		t.Error("expected ok=false for missing key")
	}
}

func TestValidationResult_Metadata_InitialMap(t *testing.T) {
	meta := map[string]any{"x": 1}
	r := types.NewValidationResult(true, "ok", meta, nil)
	val, ok := r.GetMetadata("x")
	if !ok || val != 1 {
		t.Errorf("expected x=1, got ok=%v val=%v", ok, val)
	}
}

// ─────────────────────────────────────────────────────────────────────────── //
// ValidationFunc
// ─────────────────────────────────────────────────────────────────────────── //

func makeVF(priority int, valid bool) ci.IValidationFunc[string] {
	return types.NewValidationFunc[string](priority, func(s *string, _ ...any) ci.IValidationResult {
		return types.NewValidationResult(valid, "", nil, nil)
	})
}

func TestValidationFunc_Priority(t *testing.T) {
	vf := makeVF(5, true)
	if vf.GetPriority() != 5 {
		t.Errorf("expected priority 5, got %d", vf.GetPriority())
	}
}

func TestValidationFunc_SetPriority(t *testing.T) {
	vf := makeVF(1, true)
	vf.SetPriority(99)
	if vf.GetPriority() != 99 {
		t.Errorf("expected priority 99, got %d", vf.GetPriority())
	}
}

func TestValidationFunc_GetFunction(t *testing.T) {
	vf := makeVF(1, true)
	if vf.GetFunction() == nil {
		t.Error("GetFunction() should not return nil")
	}
}

// ─────────────────────────────────────────────────────────────────────────── //
// Validation[T]
// ─────────────────────────────────────────────────────────────────────────── //

func TestValidation_AddValidator(t *testing.T) {
	v := types.NewValidation[string]()
	if err := v.AddValidator(makeVF(1, true)); err != nil {
		t.Fatalf("AddValidator failed: %v", err)
	}
	validators := v.GetValidators()
	if len(validators) != 1 {
		t.Errorf("expected 1 validator, got %d", len(validators))
	}
}

func TestValidation_AddValidator_DuplicatePriority(t *testing.T) {
	v := types.NewValidation[string]()
	v.AddValidator(makeVF(1, true))
	err := v.AddValidator(makeVF(1, true)) // same priority
	if err == nil {
		t.Error("expected error for duplicate priority, got nil")
	}
}

func TestValidation_RemoveValidator(t *testing.T) {
	v := types.NewValidation[string]()
	v.AddValidator(makeVF(1, true))
	v.AddValidator(makeVF(2, true))

	if err := v.RemoveValidator(1); err != nil {
		t.Fatalf("RemoveValidator(1) failed: %v", err)
	}
	if len(v.GetValidators()) != 1 {
		t.Errorf("expected 1 validator after remove, got %d", len(v.GetValidators()))
	}
}

func TestValidation_RemoveValidator_Nonexistent(t *testing.T) {
	v := types.NewValidation[string]()
	if err := v.RemoveValidator(99); err == nil {
		t.Error("expected error when removing nonexistent validator")
	}
}

func TestValidation_Validate_AllPass(t *testing.T) {
	v := types.NewValidation[string]()
	v.AddValidator(makeVF(1, true))
	v.AddValidator(makeVF(2, true))

	value := "hello"
	result := v.Validate(&value)
	if !result.GetIsValid() {
		t.Error("expected valid result when all validators pass")
	}
	if !v.IsValid() {
		t.Error("IsValid() should return true when all pass")
	}
}

func TestValidation_Validate_OneFails(t *testing.T) {
	v := types.NewValidation[string]()
	v.AddValidator(makeVF(1, true))
	v.AddValidator(makeVF(2, false)) // fails

	value := "hello"
	result := v.Validate(&value)
	if result.GetIsValid() {
		t.Error("expected invalid result when one validator fails")
	}
	if v.IsValid() {
		t.Error("IsValid() should return false when one validator fails")
	}
}

func TestValidation_Validate_CallsAllValidators(t *testing.T) {
	// sync.Map.Range does not guarantee iteration order, so we only verify
	// that all three validators are executed — not their execution sequence.
	v := types.NewValidation[string]()
	called := map[int]bool{}

	for _, p := range []int{1, 5, 10} {
		priority := p
		v.AddValidator(types.NewValidationFunc[string](priority, func(s *string, _ ...any) ci.IValidationResult {
			called[priority] = true
			return types.NewValidationResult(true, "", nil, nil)
		}))
	}

	value := "test"
	v.Validate(&value)

	if len(called) != 3 {
		t.Fatalf("expected 3 validators called, got %d: %v", len(called), called)
	}
	for _, p := range []int{1, 5, 10} {
		if !called[p] {
			t.Errorf("validator with priority %d was not called", p)
		}
	}
}

func TestValidation_ClearResults(t *testing.T) {
	// ClearResults resets the stored IValidationResult on each validator to nil.
	// GetResults still returns one entry per validator (the map key is the validator,
	// not the result), but each value is nil after clearing.
	v := types.NewValidation[string]()
	v.AddValidator(makeVF(1, true))

	value := "x"
	v.Validate(&value)
	v.ClearResults()

	results := v.GetResults()
	for priority, result := range results {
		if result != nil {
			t.Errorf("priority %d: expected nil result after ClearResults, got %v", priority, result)
		}
	}
}

func TestValidation_CheckIfWillValidate_WithValidators(t *testing.T) {
	v := types.NewValidation[string]()
	v.AddValidator(makeVF(1, true))
	if !v.CheckIfWillValidate() {
		t.Error("CheckIfWillValidate should return true when validators are registered")
	}
}

func TestValidation_CheckIfWillValidate_Empty(t *testing.T) {
	v := types.NewValidation[string]()
	if v.CheckIfWillValidate() {
		t.Error("CheckIfWillValidate should return false when no validators registered")
	}
}

func TestValidation_GetValidator(t *testing.T) {
	v := types.NewValidation[string]()
	v.AddValidator(makeVF(3, true))

	got, err := v.GetValidator(3)
	if err != nil {
		t.Fatalf("GetValidator(3): unexpected error: %v", err)
	}
	if got == nil {
		t.Error("GetValidator(3): expected non-nil")
	}

	_, err = v.GetValidator(99)
	if err == nil {
		t.Error("GetValidator(99): expected error for missing priority")
	}
}
