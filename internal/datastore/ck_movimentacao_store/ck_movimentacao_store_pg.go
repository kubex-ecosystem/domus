package ckmovimentacaostore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kubex-ecosystem/domus/internal/execution"
	t "github.com/kubex-ecosystem/domus/internal/types"
)

type pgCKMovimentacaoStore struct {
	exec execution.PGExecutor
}

func NewPGCKMovimentacaoStore(exec execution.PGExecutor) CKMovimentacaoStore {
	return &pgCKMovimentacaoStore{exec: exec}
}

func (s *pgCKMovimentacaoStore) Registrar(ctx context.Context, input *RegistrarMovimentacaoInput) (*CKMovimentacao, error) {
	if input == nil || strings.TrimSpace(input.InsumoID) == "" {
		return nil, fmt.Errorf("insumo_id is required: %w", t.ErrInvalidInput)
	}

	// Para tipo 'baixa': idempotente via unique index. INSERT ... ON CONFLICT DO NOTHING;
	// se não inseriu (conflito), busca o registro existente.
	if input.Tipo == "baixa" && input.OrderID != nil {
		const q = `
			INSERT INTO cozinha_pass.movimentacao_estoque (insumo_id, order_id, tipo, quantidade, created_at)
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT ON CONSTRAINT idx_ck_movimentacao_baixa_idempotente DO NOTHING
			RETURNING id, insumo_id, order_id, tipo, quantidade, created_at
		`
		row := s.exec.QueryRow(ctx, q, input.InsumoID, input.OrderID, input.Tipo, input.Quantidade, time.Now().UTC())
		mov, err := scanCKMovimentacao(row)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// Conflito: retorna o registro existente
				return s.buscarBaixaExistente(ctx, *input.OrderID, input.InsumoID)
			}
			return nil, fmt.Errorf("failed to register baixa: %w", err)
		}
		return mov, nil
	}

	const q = `
		INSERT INTO cozinha_pass.movimentacao_estoque (insumo_id, order_id, tipo, quantidade, created_at)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, insumo_id, order_id, tipo, quantidade, created_at
	`
	mov, err := scanCKMovimentacao(s.exec.QueryRow(ctx, q,
		input.InsumoID, input.OrderID, input.Tipo, input.Quantidade, time.Now().UTC(),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to register movimentacao: %w", err)
	}
	return mov, nil
}

func (s *pgCKMovimentacaoStore) buscarBaixaExistente(ctx context.Context, orderID, insumoID string) (*CKMovimentacao, error) {
	const q = `
		SELECT id, insumo_id, order_id, tipo, quantidade, created_at
		FROM cozinha_pass.movimentacao_estoque
		WHERE order_id = $1 AND insumo_id = $2 AND tipo = 'baixa'
	`
	mov, err := scanCKMovimentacao(s.exec.QueryRow(ctx, q, orderID, insumoID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing baixa: %w", err)
	}
	return mov, nil
}

func (s *pgCKMovimentacaoStore) ListByInsumo(ctx context.Context, insumoID string, limit int) ([]CKMovimentacao, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.exec.Query(ctx, `
		SELECT id, insumo_id, order_id, tipo, quantidade, created_at
		FROM cozinha_pass.movimentacao_estoque
		WHERE insumo_id = $1
		ORDER BY created_at DESC LIMIT $2
	`, insumoID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list movimentacoes by insumo: %w", err)
	}
	defer rows.Close()
	return scanCKMovimentacaoRows(rows)
}

func (s *pgCKMovimentacaoStore) ListByOrder(ctx context.Context, orderID string) ([]CKMovimentacao, error) {
	rows, err := s.exec.Query(ctx, `
		SELECT id, insumo_id, order_id, tipo, quantidade, created_at
		FROM cozinha_pass.movimentacao_estoque
		WHERE order_id = $1
		ORDER BY created_at DESC
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to list movimentacoes by order: %w", err)
	}
	defer rows.Close()
	return scanCKMovimentacaoRows(rows)
}

func scanCKMovimentacao(scanner interface{ Scan(...any) error }) (*CKMovimentacao, error) {
	var m CKMovimentacao
	var createdAt time.Time
	err := scanner.Scan(&m.ID, &m.InsumoID, &m.OrderID, &m.Tipo, &m.Quantidade, &createdAt)
	if err != nil {
		return nil, err
	}
	m.CreatedAt = &createdAt
	return &m, nil
}

func scanCKMovimentacaoRows(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]CKMovimentacao, error) {
	movs := []CKMovimentacao{}
	for rows.Next() {
		m, err := scanCKMovimentacao(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan movimentacao: %w", err)
		}
		movs = append(movs, *m)
	}
	return movs, rows.Err()
}

func (s *pgCKMovimentacaoStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[CKMovimentacao](), "pg_ck_movimentacao_store", nil
}
func (s *pgCKMovimentacaoStore) GetName() string { return "pg_ck_movimentacao_store" }
func (s *pgCKMovimentacaoStore) Name() string    { return "ck_movimentacao" }
func (s *pgCKMovimentacaoStore) Validate() error {
	if s.exec == nil {
		return errors.New("PGExecutor is nil")
	}
	return nil
}
func (s *pgCKMovimentacaoStore) Close() error { return nil }
