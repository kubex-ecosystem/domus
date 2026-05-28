// Package schemasstore fornece o registro de schemas dinâmicos baseado no manifesto embedado.
package schemasstore

import (
	"encoding/json"
	"os"

	"github.com/kubex-ecosystem/domus/internal/bootstrap" // onde está o embed

	kbxget "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

// Registry que o seu BE vai consultar para saber o que existe

type SchemaRegistry struct{ Entities map[string]string }

type BootstrapManifest struct {
	ExecutionOrder []struct {
		Name    string   `json:"name"`
		Creates []string `json:"creates"` // Lista o que cada step cria
	} `json:"execution_order"`
}

func NewSchemaRegistry() (*SchemaRegistry, error) {
	bsFile := kbxget.ValOrType(
		os.ExpandEnv(kbxget.EnvOr("BOOTSTRAP_MANIFEST", "")),
		"embedded/bootstrap.manifest.json",
	)

	// 1. Lê o manifesto que já está embedado no binário
	data, err := bootstrap.MigrationFiles.ReadFile(bsFile)
	if err != nil {
		return nil, gl.Errorf("falha ao ler manifesto: %v", err)
	}

	// 2. Decodifica a estrutura que você já definiu no manifest
	var manifest BootstrapManifest

	if err := json.Unmarshal(data, &manifest); err != nil {
		gl.Errorf("falha ao decodificar manifesto: %v", err)
		return nil, err
	}

	// 3. Mapeia as entidades dinamicamente
	registry := &SchemaRegistry{Entities: make(map[string]string)}
	for _, step := range manifest.ExecutionOrder {
		for _, item := range step.Creates {

			// Filtra apenas o que for "table:" para o CRUD dinâmico
			if len(item) > 6 && item[:6] == "table:" {
				tableName := item[7:]
				registry.Entities[tableName] = step.Name
			}
		}
	}

	gl.Debugf("SchemaRegistry carregado com %d entidades", len(registry.Entities))

	// 4. Retorna o registry populado
	return registry, nil
}
