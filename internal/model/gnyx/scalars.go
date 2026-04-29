// Package gnyx centraliza os modelos de dados migrados do schema SQL original,
// permitindo que camadas superiores trabalhem com estruturas tipadas em Go.
package gnyx

import (
	"time"

	"github.com/google/uuid"
)

// Scalar aliases usados para padronizar tipos em todas as estruturas.
type (
	UUID = uuid.UUID
	Inet = string
)

type Timestamp struct {
	time.Time
}

func (t Timestamp) Unix() int64    { return t.Time.Unix() }
func (t Timestamp) String() string { return t.Time.String() }

// JSONValue representa colunas json/jsonb de forma flexível.
type JSONValue map[string]any
