package schemas

import (
	"github.com/kubex-ecosystem/domus/internal/extensions"
)

func init() {
	extensions.Global.Register(&CozinhaPassFichaFKSchema{})
}

// CozinhaPassFichaFKSchema troca as FKs da ficha técnica para insumos e
// produtos referenciados de ON DELETE RESTRICT para NO ACTION DEFERRABLE
// INITIALLY DEFERRED.
//
// Com RESTRICT, apagar uma empresa que tem ficha técnica falhava: o cascade
// company → insumo dispara a checagem antes de o cascade
// company → produto → componente_ficha remover as linhas que apontavam para o
// insumo. NO ACTION simples não basta (a checagem roda na fila de triggers,
// ainda antes do cascade de dois níveis). Adiada, a checagem roda no commit,
// quando a ficha já saiu junto. A proteção continua: apagar direto um insumo ou
// produto ainda usado em alguma ficha segue recusado (no commit do comando).
//
// Extensão separada porque cozinha_pass_schema já foi aplicada e o
// gerenciador não reaplica extensão marcada. Espelhado no fim de
// `bootstrap/embedded/core/etapa_25_cozinha_pass.sql`.
type CozinhaPassFichaFKSchema struct{}

func (s *CozinhaPassFichaFKSchema) ID() string { return "cozinha_pass_ficha_fk" }

func (s *CozinhaPassFichaFKSchema) Dependencies() []string {
	return []string{"cozinha_pass_schema"}
}

func (s *CozinhaPassFichaFKSchema) Statements() []string {
	return []string{
		`ALTER TABLE cozinha_pass.componente_ficha
			DROP CONSTRAINT IF EXISTS componente_ficha_insumo_ref_id_fkey,
			ADD CONSTRAINT componente_ficha_insumo_ref_id_fkey
				FOREIGN KEY (insumo_ref_id) REFERENCES cozinha_pass.insumo(id) ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED`,
		`ALTER TABLE cozinha_pass.componente_ficha
			DROP CONSTRAINT IF EXISTS componente_ficha_produto_ref_id_fkey,
			ADD CONSTRAINT componente_ficha_produto_ref_id_fkey
				FOREIGN KEY (produto_ref_id) REFERENCES cozinha_pass.produto(id) ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED`,
	}
}
