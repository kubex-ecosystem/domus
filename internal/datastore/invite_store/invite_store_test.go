package invitestore

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ─────────────────────────────────────────────────────────────────────────── //
// noopExec — minimal PGExecutor for tests that don't hit the database.
// ─────────────────────────────────────────────────────────────────────────── //

type noopExec struct{}

func (e *noopExec) Exec(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (e *noopExec) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	return nil, nil
}
func (e *noopExec) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row { return nil }
func (e *noopExec) BeginTx(_ context.Context) (pgx.Tx, error)              { return nil, nil }
func (e *noopExec) Pool() *pgxpool.Pool                                     { return nil }

// ─────────────────────────────────────────────────────────────────────────── //
// tableName
// ─────────────────────────────────────────────────────────────────────────── //

func TestTableName(t *testing.T) {
	cases := []struct {
		invType InvitationType
		want    string
	}{
		{TypePartner, "partner_invitation"},
		{TypeInternal, "internal_invitation"},
		{TypeDemo, "internal_invitation"},
		{TypeEmployee, "internal_invitation"},
		{TypeUser, "internal_invitation"},
	}
	for _, c := range cases {
		got := tableName(c.invType)
		if got != c.want {
			t.Errorf("tableName(%q) = %q, want %q", c.invType, got, c.want)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────── //
// normalizePage / normalizeLimit / calcOffset / calcTotalPages
// ─────────────────────────────────────────────────────────────────────────── //

func TestNormalizePage(t *testing.T) {
	cases := []struct{ input, want int }{{0, 1}, {-5, 1}, {1, 1}, {3, 3}, {100, 100}}
	for _, c := range cases {
		if got := normalizePage(c.input); got != c.want {
			t.Errorf("normalizePage(%d) = %d, want %d", c.input, got, c.want)
		}
	}
}

func TestNormalizeLimit(t *testing.T) {
	cases := []struct {
		limit, def, max, want int
	}{
		{0, 20, 100, 20},   // zero → default
		{-1, 20, 100, 20},  // negative → default
		{50, 20, 100, 50},  // within bounds
		{200, 20, 100, 100}, // exceeds max → max
		{100, 20, 100, 100}, // exactly max
	}
	for _, c := range cases {
		if got := normalizeLimit(c.limit, c.def, c.max); got != c.want {
			t.Errorf("normalizeLimit(%d,%d,%d) = %d, want %d",
				c.limit, c.def, c.max, got, c.want)
		}
	}
}

func TestCalcOffset(t *testing.T) {
	cases := []struct{ page, limit, want int }{
		{1, 20, 0},
		{2, 20, 20},
		{3, 10, 20},
		{1, 100, 0},
		{5, 25, 100},
	}
	for _, c := range cases {
		if got := calcOffset(c.page, c.limit); got != c.want {
			t.Errorf("calcOffset(%d,%d) = %d, want %d", c.page, c.limit, got, c.want)
		}
	}
}

func TestCalcTotalPages(t *testing.T) {
	cases := []struct {
		total    int64
		limit    int
		want     int
	}{
		{0, 20, 0},    // no data → 0 pages
		{1, 20, 1},    // less than limit → 1 page
		{20, 20, 1},   // exactly 1 page
		{21, 20, 2},   // spillover → 2 pages
		{100, 20, 5},  // exactly 5 pages
		{101, 20, 6},  // 1 extra item → 6 pages
		{0, 0, 0},     // zero limit guard
	}
	for _, c := range cases {
		if got := calcTotalPages(c.total, c.limit); got != c.want {
			t.Errorf("calcTotalPages(%d,%d) = %d, want %d", c.total, c.limit, got, c.want)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────── //
// validateCreateInput
// ─────────────────────────────────────────────────────────────────────────── //

func validInput() *CreateInvitationInput {
	return &CreateInvitationInput{
		Type:      TypePartner,
		Name:      "Test User",
		Email:     "test@example.com",
		Role:      "admin",
		Token:     "tok-abc",
		TenantID:  "tenant-1",
		InvitedBy: "user-1",
	}
}

func TestValidateCreateInput_Valid(t *testing.T) {
	if err := validateCreateInput(validInput()); err != nil {
		t.Errorf("expected nil error for valid input, got %v", err)
	}
}

func TestValidateCreateInput_RequiredFields(t *testing.T) {
	cases := []struct {
		name  string
		mutate func(*CreateInvitationInput)
	}{
		{"empty email", func(i *CreateInvitationInput) { i.Email = "" }},
		{"whitespace email", func(i *CreateInvitationInput) { i.Email = "   " }},
		{"empty token", func(i *CreateInvitationInput) { i.Token = "" }},
		{"empty role", func(i *CreateInvitationInput) { i.Role = "" }},
		{"empty tenant_id", func(i *CreateInvitationInput) { i.TenantID = "" }},
		{"empty invited_by", func(i *CreateInvitationInput) { i.InvitedBy = "" }},
		{"invalid type", func(i *CreateInvitationInput) { i.Type = "unknown" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := validInput()
			c.mutate(in)
			if err := validateCreateInput(in); err == nil {
				t.Errorf("expected error for %q, got nil", c.name)
			}
		})
	}
}

func TestValidateCreateInput_InternalTypeAllowed(t *testing.T) {
	in := validInput()
	in.Type = TypeInternal
	if err := validateCreateInput(in); err != nil {
		t.Errorf("TypeInternal should be valid, got %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────── //
// buildInviteFilters
// ─────────────────────────────────────────────────────────────────────────── //

func ptr[T any](v T) *T { return &v }

func TestBuildInviteFilters_Nil(t *testing.T) {
	where, args := buildInviteFilters(nil)
	if where != "" || len(args) != 0 {
		t.Errorf("nil filters: want empty, got where=%q args=%v", where, args)
	}
}

func TestBuildInviteFilters_NoFilters(t *testing.T) {
	typ := TypePartner
	where, args := buildInviteFilters(&InvitationFilters{Type: &typ})
	if where != "" || len(args) != 0 {
		t.Errorf("empty filters: want empty WHERE, got %q / %v", where, args)
	}
}

func TestBuildInviteFilters_Email(t *testing.T) {
	typ := TypePartner
	where, args := buildInviteFilters(&InvitationFilters{
		Type:  &typ,
		Email: ptr("Test@EXAMPLE.COM"),
	})
	if where != "WHERE LOWER(email) = $1" {
		t.Errorf("unexpected WHERE: %q", where)
	}
	if len(args) != 1 || args[0] != "test@example.com" {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestBuildInviteFilters_MultipleFields(t *testing.T) {
	typ := TypePartner
	st := StatusPending
	where, args := buildInviteFilters(&InvitationFilters{
		Type:      &typ,
		TenantID:  ptr("t1"),
		Status:    &st,
		InvitedBy: ptr("u1"),
	})
	// Should produce 3 conditions
	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d: %v", len(args), args)
	}
	if where == "" {
		t.Error("expected non-empty WHERE clause")
	}
}

func TestBuildInviteFilters_WhitespaceIgnored(t *testing.T) {
	typ := TypePartner
	where, args := buildInviteFilters(&InvitationFilters{
		Type:  &typ,
		Email: ptr("   "), // whitespace-only → should be ignored
	})
	if where != "" || len(args) != 0 {
		t.Errorf("whitespace email should be ignored, got where=%q args=%v", where, args)
	}
}

// ─────────────────────────────────────────────────────────────────────────── //
// computeExpiry
// ─────────────────────────────────────────────────────────────────────────── //

func TestComputeExpiry_NilUsesDefaultTTL(t *testing.T) {
	ttl := 7 * 24 * time.Hour
	s := &pgInviteStore{defaultTTL: ttl}

	before := time.Now().UTC()
	expiry := s.computeExpiry(nil)
	after := time.Now().UTC()

	low := before.Add(ttl)
	high := after.Add(ttl)

	if expiry.Before(low) || expiry.After(high) {
		t.Errorf("expiry %v not in expected range [%v, %v]", expiry, low, high)
	}
}

func TestComputeExpiry_ExplicitTime(t *testing.T) {
	s := &pgInviteStore{defaultTTL: 7 * 24 * time.Hour}
	explicit := time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC)
	got := s.computeExpiry(&explicit)
	if !got.Equal(explicit) {
		t.Errorf("computeExpiry: got %v, want %v", got, explicit)
	}
}

// ─────────────────────────────────────────────────────────────────────────── //
// pgInviteStore — input validation via nil exec (returns before DB hit)
// ─────────────────────────────────────────────────────────────────────────── //

func newNilStore() InviteStore { return NewPGInviteStore(nil) }

func TestStore_Create_NilInput(t *testing.T) {
	_, err := newNilStore().Create(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil input, got nil")
	}
}

func TestStore_Create_InvalidInput(t *testing.T) {
	_, err := newNilStore().Create(context.Background(), &CreateInvitationInput{}) // all empty
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

func TestStore_GetByID_EmptyID(t *testing.T) {
	inv, err := newNilStore().GetByID(context.Background(), "", TypePartner)
	if err != nil || inv != nil {
		t.Errorf("empty ID: want (nil, nil), got (%v, %v)", inv, err)
	}
}

func TestStore_GetByToken_EmptyToken(t *testing.T) {
	inv, err := newNilStore().GetByToken(context.Background(), "")
	if err != nil || inv != nil {
		t.Errorf("empty token: want (nil, nil), got (%v, %v)", inv, err)
	}
}

func TestStore_Update_NilInput(t *testing.T) {
	_, err := newNilStore().Update(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil update input, got nil")
	}
}

func TestStore_Update_EmptyID(t *testing.T) {
	_, err := newNilStore().Update(context.Background(), &UpdateInvitationInput{ID: ""})
	if err == nil {
		t.Error("expected error for empty ID, got nil")
	}
}

func TestStore_Accept_EmptyToken(t *testing.T) {
	_, err := newNilStore().Accept(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty token, got nil")
	}
}

func TestStore_Delete_EmptyID(t *testing.T) {
	err := newNilStore().Delete(context.Background(), "", TypePartner)
	if err == nil {
		t.Error("expected error for empty ID, got nil")
	}
}

func TestStore_List_NilFilters(t *testing.T) {
	_, err := newNilStore().List(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil filters, got nil")
	}
}

func TestStore_List_NilType(t *testing.T) {
	_, err := newNilStore().List(context.Background(), &InvitationFilters{Type: nil})
	if err == nil {
		t.Error("expected error for nil filters.Type, got nil")
	}
}

func TestStore_Count_NilFilters(t *testing.T) {
	_, err := newNilStore().Count(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil filters, got nil")
	}
}

// ─────────────────────────────────────────────────────────────────────────── //
// Validate
// ─────────────────────────────────────────────────────────────────────────── //

func TestStore_Validate_NilExec(t *testing.T) {
	s := NewPGInviteStore(nil)
	if err := s.(interface{ Validate() error }).Validate(); err == nil {
		t.Error("expected error for nil executor, got nil")
	}
}

func TestStore_Validate_ValidExec(t *testing.T) {
	s := NewPGInviteStore(&noopExec{})
	if err := s.(interface{ Validate() error }).Validate(); err != nil {
		t.Errorf("expected nil error for valid executor, got %v", err)
	}
}

func TestStore_GetName(t *testing.T) {
	s := NewPGInviteStore(&noopExec{})
	name := s.(interface{ GetName() string }).GetName()
	if name == "" {
		t.Error("GetName() should not be empty")
	}
}
