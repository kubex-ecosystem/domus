package models

import (
	"errors"
	"reflect"
	"testing"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/kubex-ecosystem/domus/internal/module/kbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementation of the Model interface
type MockModel struct {
	kbx.Reference
	ID    uuid.UUID
	Valid bool
}

func (m *MockModel) GetID() uuid.UUID {
	return m.ID
}

func (m *MockModel) Validate() error {
	if m.Valid {
		return nil
	}
	return errors.New("validation failed")
}

func (m *MockModel) SetName(name string) {
	// No-op for mock
}

func (m *MockModel) String() string {
	return "MockModel"
}

func TestFromModel(t *testing.T) {
	// Register the MockModel type
	mockModelType := reflect.TypeOf(&MockModel{})
	err := RegisterModel(mockModelType)
	assert.NoError(t, err)

	t.Run("Valid Model", func(t *testing.T) {
		validModel := &MockModel{ID: uuid.New(), Valid: true, Reference: kbx.NewReference("123")}

		registry := NewModelRegistry[Model]()
		result := registry.FromModel(validModel)

		assert.NotNil(t, result)

		mr, ok := result.(*ModelRegistryImpl[Model])
		require.True(t, ok, "expected ModelRegistryImpl, got %T", result)

		recovered := mr.ToModel()
		require.NotNil(t, recovered)

		typed, ok := recovered.(*MockModel)
		require.True(t, ok, "expected *MockModel, got %T", recovered)
		assert.Equal(t, validModel.GetID(), typed.GetID())
		assert.NoError(t, typed.Validate())
	})

	t.Run("Nil Model", func(t *testing.T) {
		// a questão é o MockModel não está implementando o Model,
		// então não é possível fazer o type assertion.
		// O tipo generic é sempre uma interface referindo ao Model,
		// Então é sempre um ponteiro.
		registry := NewModelRegistry[*MockModel]()
		result := registry.FromModel(nil)

		assert.Nil(t, result)
	})

	t.Run("Invalid Model", func(t *testing.T) {
		invalidModel := &MockModel{ID: uuid.New(), Valid: false}
		registry := NewModelRegistry[*MockModel]()
		result := registry.FromModel(invalidModel)

		assert.Nil(t, result)
	})

	t.Run("Unregistered Model", func(t *testing.T) {
		unregisteredModel := &struct {
			ID string
		}{ID: "456"}
		registry := NewModelRegistry[*MockModel]()
		result := registry.FromModel(unregisteredModel)

		assert.Nil(t, result)
	})
}

func TestSerialization(t *testing.T) {
	t.Run("Serialize and Deserialize Valid Model", func(t *testing.T) {
		validModel := &MockModel{ID: uuid.New(), Valid: true, Reference: kbx.NewReference("123")}
		registry := NewModelRegistry[*MockModel]()
		registry.FromModel(validModel)

		// Serialize
		data, err := json.Marshal(registry)
		assert.NoError(t, err)

		// Deserialize
		deserializedRegistry, err := NewModelRegistryFromSerialized[*MockModel](data)
		assert.NoError(t, err)

		// Convert back to model
		deserializedModel := deserializedRegistry.ToModel().(*MockModel)
		assert.Equal(t, validModel.GetID(), deserializedModel.GetID())
		assert.NoError(t, deserializedModel.Validate())
	})
}
