package schemas

import (
	"github.com/kubex-ecosystem/domus/internal/extensions"
)

func init() {
	extensions.Global.Register(&CozinhaPassTenancySchema{})
}

// CozinhaPassTenancySchema liga cada empresa do KDS ao tenant dono dela
// (cozinha_pass.company.tenant_id -> public.tenant). É uma extensão separada
// porque cozinha_pass_schema já foi aplicada em bancos existentes e o
// gerenciador de extensões não reaplica uma extensão marcada como aplicada.
//
// A coluna é aditiva e anulável: bancos com empresas antigas continuam
// subindo. O gateway trata empresa sem tenant como invisível (fail-closed),
// então dados antigos só voltam a aparecer depois de receberem um tenant_id.
// Espelhado no fim de `bootstrap/embedded/core/etapa_25_cozinha_pass.sql`.
type CozinhaPassTenancySchema struct{}

func (s *CozinhaPassTenancySchema) ID() string { return "cozinha_pass_tenancy" }

func (s *CozinhaPassTenancySchema) Dependencies() []string {
	return []string{"cozinha_pass_schema"}
}

func (s *CozinhaPassTenancySchema) Statements() []string {
	return []string{
		`ALTER TABLE cozinha_pass.company
			ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES public.tenant(id) ON DELETE CASCADE`,
		`CREATE INDEX IF NOT EXISTS idx_ck_company_tenant ON cozinha_pass.company (tenant_id)`,
	}
}
