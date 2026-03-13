// Package provider utilities for DSN manipulation and endpoint management.
package provider

import (
	"strconv"

	"github.com/kubex-ecosystem/domus/internal/module/kbx"
	"github.com/kubex-ecosystem/domus/internal/types"

	kbxGet "github.com/kubex-ecosystem/kbx/get"
)

// BuildEndpoint creates an Endpoint from a DBConfig with DSN generation and redaction.
func BuildEndpoint(dbConfig *kbx.DBConfig) Endpoint {
	return Endpoint{
		DSN: types.NewDSNFromDBConfig[types.Driver](*dbConfig),
	}
}

// ConvertRootConfigToStartSpec translates a RootConfig into a StartSpec.
// This helper is useful for CLI commands that need to convert high-level
// configuration into provider-specific startup specifications.
func ConvertRootConfigToStartSpec(rootConfig *kbx.RootConfig) StartSpec {
	spec := StartSpec{
		Services:      []ServiceRef{},
		PreferredPort: map[string]int{},
		Secrets:       map[string]string{},
		Labels:        map[string]string{},
		Configs:       map[string]kbx.DBConfig{},
	}

	if rootConfig == nil {
		return spec
	}

	for _, db := range rootConfig.Databases {
		// Skip disabled databases
		if !*kbxGet.ValOrType(db.Enabled, new(bool)) {
			continue
		}

		// Map database type to engine
		var engine Engine
		switch db.Protocol {
		case "postgresql", "postgres":
			engine = EnginePostgres
		case "mongodb", "mongo":
			engine = EngineMongo
		case "redis":
			engine = EngineRedis
		case "rabbitmq":
			engine = EngineRabbit
		default:
			continue // Skip unknown types
		}

		// Add service reference
		spec.Services = append(spec.Services, ServiceRef{
			Name:   db.Name,
			Engine: engine,
		})

		// Parse and add preferred port
		if port, err := strconv.Atoi(db.Port); err == nil {
			spec.PreferredPort[kbxGet.ValOrType(db.ID, db.Name)] = port
		}

		// Add secrets (passwords)
		if db.Pass != "" {
			spec.Secrets[kbxGet.ValOrType(db.ID, db.Name)+"_pass"] = db.Pass
		}

		// Add labels
		spec.Labels["db_"+kbxGet.ValOrType(db.ID, db.Name)] = string(db.Protocol)

		spec.Configs[kbxGet.ValOrType(db.ID, db.Name)] = db
	}

	return spec
}

func GetConfigListByService(spec StartSpec, serviceName string) []kbx.DBConfig {
	var configs []kbx.DBConfig
	for _, svc := range spec.Services {
		if svc.Name == serviceName {
			for _, config := range spec.Configs {
				if config.Name == svc.Name {
					configs = append(configs, config)
				}
			}
		}
	}
	return configs
}
