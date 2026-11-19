package bootstrap

import (
	"encoding/json"
	"path"
	"testing"
)

// TestManifestExecutionOrder garante que o bootstrap.manifest.json embedado é
// aplicável em banco vazio: todo arquivo listado existe no embed e toda
// dependência declarada roda antes da etapa que depende dela.
func TestManifestExecutionOrder(t *testing.T) {
	data, err := MigrationFiles.ReadFile("embedded/bootstrap.manifest.json")
	if err != nil {
		t.Fatalf("falha ao ler manifesto embedado: %v", err)
	}

	var mf struct {
		ExecutionOrder []struct {
			Step         int      `json:"step"`
			File         string   `json:"file"`
			Dependencies []string `json:"dependencies"`
		} `json:"execution_order"`
	}
	if err := json.Unmarshal(data, &mf); err != nil {
		t.Fatalf("falha ao decodificar manifesto: %v", err)
	}
	if len(mf.ExecutionOrder) == 0 {
		t.Fatal("execution_order vazio")
	}

	seen := make(map[string]bool, len(mf.ExecutionOrder))
	for _, step := range mf.ExecutionOrder {
		if _, err := MigrationFiles.ReadFile(path.Join("embedded", step.File)); err != nil {
			t.Errorf("etapa %d: arquivo %q não está no embed: %v", step.Step, step.File, err)
		}
		for _, dep := range step.Dependencies {
			if !seen[dep] {
				t.Errorf("etapa %d (%s) depende de %q, que não roda antes dela", step.Step, step.File, dep)
			}
		}
		seen[path.Base(step.File)] = true
	}

	// A etapa 25 (cozinha_pass) precisa estar no caminho do `domus database migrate`.
	if !seen["etapa_25_cozinha_pass.sql"] {
		t.Error("etapa_25_cozinha_pass.sql ausente do execution_order")
	}
}
