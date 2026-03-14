package execution

import "fmt"

// QueryModel é o "modo genérico" de descrever uma operação de dados.
//
// Ele existe para:
//   - suportar backends não-SQL (mongo, redis, http, etc);
//   - permitir que agentes/IA gerem operações abstratas;
//   - habilitar pipelines metaprogramados (ex dos antigos: Analyzer, GNyx Workspace, etc).
//
// Exemplos de Command:
//   - "pg:exec"
//   - "pg:query"
//   - "mongo:find"
//   - "redis:get"
//   - "http:post"
//   - "ds:invoke"
type QueryModel struct {
	// Command identifica a operação e o backend lógico.
	Command string

	// Query carrega o "payload" da operação:
	//   - pode ser string (SQL, JSON, etc.)
	//   - pode ser struct (filtros, projeções, etc.)
	Query any

	// Params são parâmetros posicionais ou extras contextuais.
	Params []any
}

var (
	// ErrUnsupported é o erro padrão para operações não implementadas
	// pelo GenericExecutor default.
	ErrUnsupported = fmt.Errorf("execution: unsupported generic operation")
)
