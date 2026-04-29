package schemasstore

import (
	"context"
	"testing"
)

func TestMemorySchemaStore_Register(t *testing.T) {
	store := NewMemorySchemaStore()
	ctx := context.Background()

	schema := &Schema{
		Key:  "user",
		Name: "User Schema",
	}

	err := store.Register(ctx, schema)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = store.Register(ctx, schema)
	if err != ErrSchemaAlreadyExists {
		t.Fatalf("expected ErrSchemaAlreadyExists, got %v", err)
	}
}

func TestMemorySchemaStore_RegisterInvalid(t *testing.T) {
	store := NewMemorySchemaStore()
	ctx := context.Background()

	schema := &Schema{
		Key: "",
	}

	err := store.Register(ctx, schema)
	if err != ErrInvalidSchema {
		t.Fatalf("expected ErrInvalidSchema, got %v", err)
	}
}

func TestMemorySchemaStore_Get(t *testing.T) {
	store := NewMemorySchemaStore()
	ctx := context.Background()

	schema := &Schema{
		Key:  "user",
		Name: "User Schema",
	}

	_ = store.Register(ctx, schema)

	got, err := store.Get(ctx, "user")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.Key != "user" {
		t.Fatalf("expected key 'user', got %s", got.Key)
	}

	_, err = store.Get(ctx, "nonexistent")
	if err != ErrSchemaNotFound {
		t.Fatalf("expected ErrSchemaNotFound, got %v", err)
	}
}

func TestMemorySchemaStore_GetOrDefault(t *testing.T) {
	store := NewMemorySchemaStore()
	ctx := context.Background()

	defaultSchema := &Schema{
		Key:  "default",
		Name: "Default Schema",
	}

	got := store.GetOrDefault(ctx, "nonexistent", defaultSchema)
	if got.Key != "default" {
		t.Fatalf("expected default schema, got %v", got)
	}

	schema := &Schema{
		Key:  "user",
		Name: "User Schema",
	}
	_ = store.Register(ctx, schema)

	got = store.GetOrDefault(ctx, "user", defaultSchema)
	if got.Key != "user" {
		t.Fatalf("expected 'user', got %s", got.Key)
	}
}

func TestMemorySchemaStore_Update(t *testing.T) {
	store := NewMemorySchemaStore()
	ctx := context.Background()

	schema := &Schema{
		Key:  "user",
		Name: "User Schema",
	}

	err := store.Update(ctx, schema)
	if err != ErrSchemaNotFound {
		t.Fatalf("expected ErrSchemaNotFound, got %v", err)
	}

	_ = store.Register(ctx, schema)

	updated := &Schema{
		Key:  "user",
		Name: "Updated User Schema",
	}

	err = store.Update(ctx, updated)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, _ := store.Get(ctx, "user")
	if got.Name != "Updated User Schema" {
		t.Fatalf("expected updated name, got %s", got.Name)
	}
}

func TestMemorySchemaStore_Upsert(t *testing.T) {
	store := NewMemorySchemaStore()
	ctx := context.Background()

	schema := &Schema{
		Key:  "user",
		Name: "User Schema",
	}

	err := store.Upsert(ctx, schema)
	if err != nil {
		t.Fatalf("expected no error on insert, got %v", err)
	}

	updated := &Schema{
		Key:  "user",
		Name: "Updated User Schema",
	}

	err = store.Upsert(ctx, updated)
	if err != nil {
		t.Fatalf("expected no error on update, got %v", err)
	}

	got, _ := store.Get(ctx, "user")
	if got.Name != "Updated User Schema" {
		t.Fatalf("expected updated name, got %s", got.Name)
	}
}

func TestMemorySchemaStore_Delete(t *testing.T) {
	store := NewMemorySchemaStore()
	ctx := context.Background()

	err := store.Delete(ctx, "nonexistent")
	if err != ErrSchemaNotFound {
		t.Fatalf("expected ErrSchemaNotFound, got %v", err)
	}

	schema := &Schema{
		Key:  "user",
		Name: "User Schema",
	}
	_ = store.Register(ctx, schema)

	err = store.Delete(ctx, "user")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if store.Exists(ctx, "user") {
		t.Fatal("expected schema to be deleted")
	}
}

func TestMemorySchemaStore_Exists(t *testing.T) {
	store := NewMemorySchemaStore()
	ctx := context.Background()

	if store.Exists(ctx, "user") {
		t.Fatal("expected false for nonexistent schema")
	}

	schema := &Schema{
		Key:  "user",
		Name: "User Schema",
	}
	_ = store.Register(ctx, schema)

	if !store.Exists(ctx, "user") {
		t.Fatal("expected true for existing schema")
	}
}

func TestMemorySchemaStore_List(t *testing.T) {
	store := NewMemorySchemaStore()
	ctx := context.Background()

	schemas, err := store.List(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(schemas) != 0 {
		t.Fatalf("expected empty list, got %d items", len(schemas))
	}

	_ = store.Register(ctx, &Schema{Key: "user", Name: "User"})
	_ = store.Register(ctx, &Schema{Key: "product", Name: "Product"})

	schemas, err = store.List(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(schemas) != 2 {
		t.Fatalf("expected 2 schemas, got %d", len(schemas))
	}
}

func TestMemorySchemaStore_Keys(t *testing.T) {
	store := NewMemorySchemaStore()
	ctx := context.Background()

	_ = store.Register(ctx, &Schema{Key: "user", Name: "User"})
	_ = store.Register(ctx, &Schema{Key: "product", Name: "Product"})

	keys, err := store.Keys(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestMemorySchemaStore_Count(t *testing.T) {
	store := NewMemorySchemaStore()
	ctx := context.Background()

	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}

	_ = store.Register(ctx, &Schema{Key: "user", Name: "User"})
	_ = store.Register(ctx, &Schema{Key: "product", Name: "Product"})

	count, err = store.Count(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestMemorySchemaStore_Clear(t *testing.T) {
	store := NewMemorySchemaStore()
	ctx := context.Background()

	_ = store.Register(ctx, &Schema{Key: "user", Name: "User"})
	_ = store.Register(ctx, &Schema{Key: "product", Name: "Product"})

	err := store.Clear(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	count, _ := store.Count(ctx)
	if count != 0 {
		t.Fatalf("expected 0 after clear, got %d", count)
	}
}

func TestNewMemorySchemaStoreWith(t *testing.T) {
	store := NewMemorySchemaStoreWith(
		&Schema{Key: "user", Name: "User"},
		&Schema{Key: "product", Name: "Product"},
		nil,
		&Schema{Key: "", Name: "Invalid"},
	)
	ctx := context.Background()

	count, _ := store.Count(ctx)
	if count != 2 {
		t.Fatalf("expected 2 valid schemas, got %d", count)
	}
}

func TestFallbackSchemaStore_Get(t *testing.T) {
	primary := NewMemorySchemaStore()
	fallback := NewMemorySchemaStoreWith(
		&Schema{Key: "fallback-schema", Name: "From Fallback"},
	)
	store := NewFallbackSchemaStore(primary, fallback)
	ctx := context.Background()

	_ = primary.Register(ctx, &Schema{Key: "primary-schema", Name: "From Primary"})

	got, err := store.Get(ctx, "primary-schema")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Name != "From Primary" {
		t.Fatalf("expected 'From Primary', got %s", got.Name)
	}

	got, err = store.Get(ctx, "fallback-schema")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Name != "From Fallback" {
		t.Fatalf("expected 'From Fallback', got %s", got.Name)
	}

	_, err = store.Get(ctx, "nonexistent")
	if err != ErrSchemaNotFound {
		t.Fatalf("expected ErrSchemaNotFound, got %v", err)
	}
}

func TestFallbackSchemaStore_Register(t *testing.T) {
	primary := NewMemorySchemaStore()
	fallback := NewMemorySchemaStore()
	store := NewFallbackSchemaStore(primary, fallback)
	ctx := context.Background()

	schema := &Schema{Key: "new-schema", Name: "New"}
	err := store.Register(ctx, schema)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !primary.Exists(ctx, "new-schema") {
		t.Fatal("expected schema in primary store")
	}

	if fallback.Exists(ctx, "new-schema") {
		t.Fatal("expected schema NOT in fallback store")
	}
}

func TestFallbackSchemaStore_Exists(t *testing.T) {
	primary := NewMemorySchemaStoreWith(&Schema{Key: "primary", Name: "P"})
	fallback := NewMemorySchemaStoreWith(&Schema{Key: "fallback", Name: "F"})
	store := NewFallbackSchemaStore(primary, fallback)
	ctx := context.Background()

	if !store.Exists(ctx, "primary") {
		t.Fatal("expected true for primary schema")
	}

	if !store.Exists(ctx, "fallback") {
		t.Fatal("expected true for fallback schema")
	}

	if store.Exists(ctx, "nonexistent") {
		t.Fatal("expected false for nonexistent schema")
	}
}

func TestFallbackSchemaStore_List(t *testing.T) {
	primary := NewMemorySchemaStoreWith(
		&Schema{Key: "a", Name: "A"},
		&Schema{Key: "b", Name: "B"},
	)
	fallback := NewMemorySchemaStoreWith(
		&Schema{Key: "b", Name: "B-fallback"},
		&Schema{Key: "c", Name: "C"},
	)
	store := NewFallbackSchemaStore(primary, fallback)
	ctx := context.Background()

	schemas, err := store.List(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(schemas) != 3 {
		t.Fatalf("expected 3 unique schemas (a, b, c), got %d", len(schemas))
	}

	found := make(map[string]string)
	for _, s := range schemas {
		found[s.Key] = s.Name
	}

	if found["b"] != "B" {
		t.Fatal("expected primary 'b' to take precedence over fallback")
	}
}

func TestFallbackSchemaStore_Delete(t *testing.T) {
	primary := NewMemorySchemaStoreWith(&Schema{Key: "a", Name: "A"})
	fallback := NewMemorySchemaStoreWith(&Schema{Key: "b", Name: "B"})
	store := NewFallbackSchemaStore(primary, fallback)
	ctx := context.Background()

	err := store.Delete(ctx, "a")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if primary.Exists(ctx, "a") {
		t.Fatal("expected 'a' to be deleted from primary")
	}

	err = store.Delete(ctx, "b")
	if err != ErrSchemaNotFound {
		t.Fatalf("expected ErrSchemaNotFound (fallback is read-only), got %v", err)
	}
}

func TestFallbackSchemaStore_Clear(t *testing.T) {
	primary := NewMemorySchemaStoreWith(&Schema{Key: "a", Name: "A"})
	fallback := NewMemorySchemaStoreWith(&Schema{Key: "b", Name: "B"})
	store := NewFallbackSchemaStore(primary, fallback)
	ctx := context.Background()

	_ = store.Clear(ctx)

	count, _ := primary.Count(ctx)
	if count != 0 {
		t.Fatal("expected primary to be cleared")
	}

	count, _ = fallback.Count(ctx)
	if count != 1 {
		t.Fatal("expected fallback to remain untouched")
	}
}

func TestSchema_Validate(t *testing.T) {
	tests := []struct {
		name    string
		schema  *Schema
		wantErr error
	}{
		{
			name:    "valid schema",
			schema:  &Schema{Key: "user"},
			wantErr: nil,
		},
		{
			name:    "empty key",
			schema:  &Schema{Key: ""},
			wantErr: ErrInvalidSchema,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.schema.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
