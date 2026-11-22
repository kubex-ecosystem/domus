package schemas

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Integração contra Postgres descartável (DOMUS_CK_TEST_PG_DSN): parte do
// schema antigo (FKs RESTRICT), aplica a extensão duas vezes e confere que
// apagar a empresa funciona e apagar insumo em uso continua barrado.
func TestCozinhaPassFichaFK(t *testing.T) {
	dsn := os.Getenv("DOMUS_CK_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("DOMUS_CK_TEST_PG_DSN não definido")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	exec := func(sql string, args ...any) error {
		_, err := pool.Exec(ctx, sql, args...)
		return err
	}
	if err := exec(`DROP SCHEMA IF EXISTS cozinha_pass CASCADE;
		CREATE TABLE IF NOT EXISTS public.tenant (id UUID PRIMARY KEY);`); err != nil {
		t.Fatal(err)
	}
	for _, st := range (&CozinhaPassSchema{}).Statements() {
		if err := exec(st); err != nil {
			t.Fatalf("schema base: %v", err)
		}
	}
	ext := &CozinhaPassFichaFKSchema{}
	for i := 0; i < 2; i++ { // idempotente
		for _, st := range ext.Statements() {
			if err := exec(st); err != nil {
				t.Fatalf("extensão (rodada %d): %v", i+1, err)
			}
		}
	}

	id := func(sql string, args ...any) string {
		var v string
		if err := pool.QueryRow(ctx, sql, args...).Scan(&v); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
		return v
	}
	co := id(`INSERT INTO cozinha_pass.company (name) VALUES ('A') RETURNING id`)
	ins := id(`INSERT INTO cozinha_pass.insumo (company_id,nome,unidade) VALUES ($1,'tomate','g') RETURNING id`, co)
	mol := id(`INSERT INTO cozinha_pass.produto (company_id,nome,tipo) VALUES ($1,'molho','processado') RETURNING id`, co)
	pr := id(`INSERT INTO cozinha_pass.produto (company_id,nome,tipo) VALUES ($1,'prato','final') RETURNING id`, co)
	id(`INSERT INTO cozinha_pass.componente_ficha (produto_id,ref_tipo,insumo_ref_id,quantidade) VALUES ($1,'insumo',$2,1) RETURNING id`, mol, ins)
	id(`INSERT INTO cozinha_pass.componente_ficha (produto_id,ref_tipo,produto_ref_id,quantidade) VALUES ($1,'produto',$2,1) RETURNING id`, pr, mol)

	if err := exec(`DELETE FROM cozinha_pass.insumo WHERE id=$1`, ins); err == nil || !strings.Contains(err.Error(), "23503") {
		t.Fatalf("apagar insumo em uso: esperava violação de FK, veio %v", err)
	}
	if err := exec(`DELETE FROM cozinha_pass.produto WHERE id=$1`, mol); err == nil || !strings.Contains(err.Error(), "23503") {
		t.Fatalf("apagar produto em uso: esperava violação de FK, veio %v", err)
	}
	if err := exec(`DELETE FROM cozinha_pass.company WHERE id=$1`, co); err != nil {
		t.Fatalf("apagar empresa com ficha: %v", err)
	}
}
