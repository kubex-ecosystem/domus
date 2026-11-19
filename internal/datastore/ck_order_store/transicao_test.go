package ckorderstore

import "testing"

func TestTransicaoValida(t *testing.T) {
	casos := []struct {
		de, para string
		ok       bool
	}{
		{"new", "preparing", true},
		{"preparing", "ready", true},
		{"ready", "delivered", true},
		{"new", "ready", true},       // pular coluna no kanban
		{"ready", "preparing", true}, // voltar coluna
		{"new", "delivered", true},   // balcão: entrega direta
		{"preparing", "cancelled", true},
		{"ready", "ready", true},
		{"delivered", "delivered", true}, // repetição idempotente
		{"cancelled", "cancelled", true},
		{"delivered", "ready", false},
		{"delivered", "cancelled", false},
		{"cancelled", "new", false},
		{"cancelled", "delivered", false},
		{"new", "pending", false},
		{"pending", "pending", false},
		{"", "new", false},
	}
	for _, c := range casos {
		if got := TransicaoValida(c.de, c.para); got != c.ok {
			t.Errorf("TransicaoValida(%q, %q) = %v, quer %v", c.de, c.para, got, c.ok)
		}
	}
}
