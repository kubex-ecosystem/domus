package ckorderstore

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	ckmov "github.com/kubex-ecosystem/domus/internal/datastore/ck_movimentacao_store"
	"github.com/kubex-ecosystem/domus/internal/execution"
)

// Teste de integração contra Postgres de verdade. Só roda com
// DOMUS_CK_TEST_PG_DSN definido, e esse banco precisa ser descartável:
// o teste apaga e recria o schema cozinha_pass.
//
//	DOMUS_CK_TEST_PG_DSN=postgres://postgres@localhost:5432/postgres go test ./internal/datastore/ck_order_store/
func pgDeTeste(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DOMUS_CK_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("DOMUS_CK_TEST_PG_DSN não definido")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	raw, err := os.ReadFile("../../bootstrap/embedded/core/etapa_25_cozinha_pass.sql")
	if err != nil {
		t.Fatal(err)
	}
	var ddl []string
	for _, l := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(l), `\echo`) {
			ddl = append(ddl, l)
		}
	}
	setup := `
		DROP SCHEMA IF EXISTS cozinha_pass CASCADE;
		CREATE TABLE IF NOT EXISTS public.tenant (id UUID PRIMARY KEY);
	` + strings.Join(ddl, "\n") + `
		ALTER TABLE cozinha_pass.orders ADD COLUMN IF NOT EXISTS external_id TEXT;
	`
	if _, err := pool.Exec(ctx, setup); err != nil {
		t.Fatalf("setup do schema: %v", err)
	}
	return pool
}

func TestUpdateStatus_BaixaNaEntrega(t *testing.T) {
	pool := pgDeTeste(t)
	ctx := context.Background()
	q := func(sql string, args ...any) string {
		t.Helper()
		var id string
		if err := pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
		return id
	}
	estoque := func(id string) float64 {
		t.Helper()
		var v float64
		if err := pool.QueryRow(ctx, `SELECT estoque::float8 FROM cozinha_pass.insumo WHERE id=$1`, id).Scan(&v); err != nil {
			t.Fatal(err)
		}
		return v
	}

	co := q(`INSERT INTO cozinha_pass.company (name) VALUES ('A') RETURNING id`)
	outra := q(`INSERT INTO cozinha_pass.company (name) VALUES ('B') RETURNING id`)
	tomate := q(`INSERT INTO cozinha_pass.insumo (company_id,nome,unidade,estoque) VALUES ($1,'tomate','g',1000) RETURNING id`, co)
	massa := q(`INSERT INTO cozinha_pass.insumo (company_id,nome,unidade,estoque) VALUES ($1,'massa','g',150) RETURNING id`, co)
	alheio := q(`INSERT INTO cozinha_pass.insumo (company_id,nome,unidade,estoque) VALUES ($1,'alheio','g',500) RETURNING id`, outra)

	// molho (processado) = 200 g tomate; prato (final) = 2 molho + 100 g massa + 10 g de insumo de outra empresa.
	molho := q(`INSERT INTO cozinha_pass.produto (company_id,nome,tipo) VALUES ($1,'molho','processado') RETURNING id`, co)
	prato := q(`INSERT INTO cozinha_pass.produto (company_id,nome,tipo,preco_venda) VALUES ($1,'prato','final',30) RETURNING id`, co)
	q(`INSERT INTO cozinha_pass.componente_ficha (produto_id,ref_tipo,insumo_ref_id,quantidade) VALUES ($1,'insumo',$2,200) RETURNING id`, molho, tomate)
	q(`INSERT INTO cozinha_pass.componente_ficha (produto_id,ref_tipo,produto_ref_id,quantidade) VALUES ($1,'produto',$2,2) RETURNING id`, prato, molho)
	q(`INSERT INTO cozinha_pass.componente_ficha (produto_id,ref_tipo,insumo_ref_id,quantidade) VALUES ($1,'insumo',$2,100) RETURNING id`, prato, massa)
	q(`INSERT INTO cozinha_pass.componente_ficha (produto_id,ref_tipo,insumo_ref_id,quantidade) VALUES ($1,'insumo',$2,10) RETURNING id`, prato, alheio)
	// ciclo: molho também usa prato. A guarda de ciclo tem que cortar isso.
	q(`INSERT INTO cozinha_pass.componente_ficha (produto_id,ref_tipo,produto_ref_id,quantidade) VALUES ($1,'produto',$2,1) RETURNING id`, molho, prato)

	st := NewPGCKOrderStore(execution.NewPGExecutor(pool))
	pedido, err := st.Create(ctx, &CreateCKOrderInput{
		CompanyID: co, Ticket: "#1", ClientName: "c", Channel: "manual",
		Items: []CreateCKOrderItemInput{
			{Name: "prato", ProdutoID: &prato, Quantity: 2, UnitPrice: 30},
			{Name: "avulso", Quantity: 1, UnitPrice: 5}, // sem produto: não baixa nada
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range []string{"preparing", "ready"} {
		if _, err := st.UpdateStatus(ctx, &UpdateCKOrderStatusInput{ID: pedido.ID, Status: s}); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
	if estoque(tomate) != 1000 {
		t.Fatal("baixa antes da entrega")
	}

	o, err := st.UpdateStatus(ctx, &UpdateCKOrderStatusInput{ID: pedido.ID, Status: "delivered"})
	if err != nil || o.Status != "delivered" {
		t.Fatalf("delivered: %v %+v", err, o)
	}
	// 2 pratos × 2 molhos × 200 g = 800 g de tomate; 2 × 100 g = 200 g de massa (só há 150: zera).
	if got := estoque(tomate); got != 200 {
		t.Errorf("tomate = %v, quer 200", got)
	}
	if got := estoque(massa); got != 0 {
		t.Errorf("massa = %v, quer 0 (nunca negativo)", got)
	}
	if got := estoque(alheio); got != 500 {
		t.Errorf("insumo de outra empresa baixou: %v", got)
	}

	// Repetir delivered é no-op: não baixa de novo.
	if _, err := st.UpdateStatus(ctx, &UpdateCKOrderStatusInput{ID: pedido.ID, Status: "delivered"}); err != nil {
		t.Fatal(err)
	}
	if got := estoque(tomate); got != 200 {
		t.Errorf("baixa repetida: tomate = %v", got)
	}
	var n int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM cozinha_pass.movimentacao_estoque WHERE order_id=$1 AND tipo='baixa'`, pedido.ID).Scan(&n)
	if n != 2 {
		t.Errorf("movimentações de baixa = %d, quer 2", n)
	}

	// Registrar manual de uma baixa que já existe devolve a existente, sem erro
	// (antes falhava com ON CONFLICT ON CONSTRAINT num índice parcial).
	mov, err := ckmov.NewPGCKMovimentacaoStore(execution.NewPGExecutor(pool)).Registrar(ctx,
		&ckmov.RegistrarMovimentacaoInput{InsumoID: tomate, OrderID: &pedido.ID, Tipo: "baixa", Quantidade: 1})
	if err != nil || mov.Quantidade != 800 {
		t.Fatalf("Registrar baixa repetida: %v %+v", err, mov)
	}

	// Saída de status final é recusada e não mexe em nada.
	_, err = st.UpdateStatus(ctx, &UpdateCKOrderStatusInput{ID: pedido.ID, Status: "ready"})
	if !errors.Is(err, ErrTransicaoInvalida) {
		t.Fatalf("delivered → ready: err = %v, quer ErrTransicaoInvalida", err)
	}

	// Cancelado não baixa estoque e também é final.
	p2, _ := st.Create(ctx, &CreateCKOrderInput{CompanyID: co, Ticket: "#2", ClientName: "c", Channel: "manual",
		Items: []CreateCKOrderItemInput{{Name: "prato", ProdutoID: &prato, Quantity: 1, UnitPrice: 30}}})
	if _, err := st.UpdateStatus(ctx, &UpdateCKOrderStatusInput{ID: p2.ID, Status: "cancelled"}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.UpdateStatus(ctx, &UpdateCKOrderStatusInput{ID: p2.ID, Status: "delivered"}); !errors.Is(err, ErrTransicaoInvalida) {
		t.Fatalf("cancelled → delivered: err = %v", err)
	}
	if got := estoque(tomate); got != 200 {
		t.Errorf("cancelado baixou: tomate = %v", got)
	}
}
