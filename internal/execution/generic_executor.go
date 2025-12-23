package execution

import (
	"context"
)

// GenericExecutor define o caminho genérico para executar operações
// baseadas em QueryModel.
//
// Por padrão, a implementação mínima apenas retorna ErrUnsupported.
// A ideia é você expandir isso conforme for introduzindo
// backends genéricos (ex: DS remoto via HTTP, ou um router interno).
type GenericExecutor interface {
	Execute(ctx context.Context, qm QueryModel) (any, error)
}

// genericExecutor é a implementação básica.
//
// Ela é segura: não faz nada além de recusar operações com ErrUnsupported.
// Você pode evoluir isso no futuro para:
//
//   - chamar um HTTP client baseado em qm.Command;
//   - rotear para outro serviço interno;
//   - serializar QueryModel em mensageria, etc.
type genericExecutor struct{}

// NewGenericExecutor cria o GenericExecutor padrão.
func NewGenericExecutor() GenericExecutor {
	return &genericExecutor{}
}

func (g *genericExecutor) Execute(ctx context.Context, qm QueryModel) (any, error) {
	// Implementação "no-op" inicial. Ela é intencionalmente limitada
	// para evitar comportamento mágico prematuro.
	return nil, ErrUnsupported
}
