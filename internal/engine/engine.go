// Package engine fornece uma camada de orquestração de alto nível
// para o Data Service (DS). Ele junta:
//
//   - a configuração consolidada (DatabaseConfigManagerImpl)
//   - o lifecycle real de conexões (db.Manager / DatabaseLifecycleManager)
//
// v1: zero mágica extra. Só sobe, expõe conexões e desliga direito.
// Nada de provider, DockerStack, etc. ainda.
package engine

import (
	"context"
	"fmt"

	"github.com/kubex-ecosystem/domus/internal/types"

	logz "github.com/kubex-ecosystem/logz"
)

// Runtime é o “núcleo vivo” do DS em runtime:
// - segura o config consolidado
// - segura o manager de conexões
// - oferece helpers de uso direto pro resto do serviço.
type Runtime struct {
	cfg    *types.DBConfig
	mgr    *DatabaseManager
	logger *logz.LoggerZ
}

// Options permite você plugar futuros knobs (provider, hooks, etc.)
// sem quebrar a assinatura do Bootstrap.
type Options struct {
	// Nome lógico do app, só pra log/telemetria se quiser.
	AppName  string
	FilePath string

	// Futuro: aqui pode entrar provider/backend, hooks de migration,
	// métricas custom, etc.
}

// Bootstrap inicializa o runtime do DS em cima de uma configuração já carregada.
//
// Contrato mental:
//   - cfg já veio de arquivo/env/whatever (fonte da verdade do DS).
//   - aqui a gente só valida + instancia o Manager + dá Init.
//   - se der erro aqui, o DS não tem que subir.
func Bootstrap(ctx context.Context,
	cfg *types.DBConfig,
	logger *logz.LoggerZ,
	opts *Options,
) (Runtime, error) {
	if cfg == nil && (opts == nil || opts.FilePath == "") {
		return Runtime{}, fmt.Errorf("dsruntime: configuração de banco não pode ser nula")
	}

	if logger == nil {
		logger = logz.GetLoggerZ("dsruntime")
	}

	if opts == nil {
		opts = &Options{}
	}

	if opts != nil && opts.AppName != "" {
		opts.AppName = "domus"
	}

	logger.Infof("[dsruntime] bootstrap iniciado (app=%s)", opts.AppName)

	// Instancia lifecycle manager real
	mgr := NewDatabaseManager(logger)

	// Carrega/bootstraps config raiz (se necessário)
	if cfg == nil {
		var err error
		rootConfig, err := mgr.LoadOrBootstrap(opts.FilePath)
		if err != nil {
			return Runtime{}, logger.Errorf("[dsruntime] falha ao carregar/bootstraps config raiz: %v", err)
		}
		if err := mgr.InitFromRootConfig(ctx, &rootConfig); err != nil {
			return Runtime{}, logger.Errorf("[dsruntime] falha ao inicializar config raiz: %v", err)
		}
	}

	logger.Success("[dsruntime] runtime inicializado com sucesso")

	return Runtime{
		cfg:    cfg,
		mgr:    mgr,
		logger: logger,
	}, nil
}

// Config retorna a configuração consolidada usada pelo runtime.
// Útil pra introspecção, debug, endpoints de /debug/config (sem segredos, claro), etc.
func (r *Runtime) Config() *types.DBConfig {
	return r.cfg
}

// Manager expõe o Manager bruto, caso você queira acessar algo
// mais “baixo nível” que não está na fachada.
func (r *Runtime) Manager() *DatabaseManager {
	return r.mgr
}

// HealthCheck roda o health check de todas as conexões configuradas.
func (r *Runtime) HealthCheck(ctx context.Context) error {
	if r == nil || r.mgr == nil {
		return fmt.Errorf("dsruntime: runtime não inicializado")
	}
	return r.mgr.HealthCheck(ctx)
}

// Conn devolve uma conexão segura para o banco especificado.
// Usa exatamente a semântica do SecureConn do Manager:
// - verifica se o manager foi inicializado
// - tenta reconectar se o ping falhar.
func (r *Runtime) Conn(ctx context.Context, dbName string) (*types.DBConnection, error) {
	if r == nil || r.mgr == nil {
		return nil, fmt.Errorf("dsruntime: runtime não inicializado")
	}
	return r.mgr.SecureConn(ctx, dbName)
}

// Shutdown encerra todas as conexões de forma ordenada.
// É o que você chama no graceful shutdown do DS (signal handler, etc.).
func (r *Runtime) Shutdown(ctx context.Context) error {
	if r == nil || r.mgr == nil {
		return fmt.Errorf("dsruntime: runtime não inicializado")
	}
	return r.mgr.Shutdown(ctx)
}
