// Package client provides structures and types for database client configuration and management.
// It defines configurations for multiple backend databases and the main DS client.
package client

import (
	kbx "github.com/kubex-ecosystem/domus/internal/module/kbx"
)

// Reference is an alias for the reference type.
type Reference = kbx.Reference

func NewReference(name string) Reference { return kbx.NewReference(name) }

// BackendConfig holds configuration for a single backend database.
type BackendConfig struct {
	// Engine is the database engine for the backend.
	Engine string `yaml:"engine,omitempty" json:"engine,omitempty" mapstructure:"engine"`
	// DBName is the name of the database for the backend.
	DBName string `yaml:"db_name,omitempty" json:"db_name,omitempty" mapstructure:"db_name"`
	// DBConfigFile is the path to the database configuration file for the backend.
	DBConfigFile string `yaml:"db_config_file,omitempty" json:"db_config_file,omitempty" mapstructure:"db_config_file"`
	// Options holds additional options for the backend configuration.
	Options map[string]any `yaml:"options,omitempty" json:"options,omitempty" mapstructure:"options"`
}

// NewBackendConfig creates a new BackendConfig instance.
func NewBackendConfig(engine, dbName, dbConfigFile string, options map[string]any) *BackendConfig {
	return &BackendConfig{
		Engine:       engine,
		DBName:       dbName,
		DBConfigFile: dbConfigFile,
		Options:      options,
	}
}

// DSClientConfig holds configuration for the DS client with multiple backends.
type DSClientConfig struct {
	kbx.Reference `yaml:",inline" json:",inline" mapstructure:",squash"`
	// FilePath is the path to the DS client configuration file.
	FilePath string `yaml:"file_path,omitempty" json:"file_path,omitempty" mapstructure:"file_path"`
	// Backends holds the configurations for multiple backends.
	Backends map[string]*BackendConfig `yaml:"backends,omitempty" json:"backends,omitempty" mapstructure:"backends"`
}

func NewDSClientConfig(name string, filePath string, backendConfig ...*BackendConfig) *DSClientConfig {
	ref := kbx.NewReference(name)

	// Build backends map
	backends := make(map[string]*BackendConfig)

	// Populate backends map with backend configurations
	for _, bc := range backendConfig {
		if bc == nil {
			continue
		}
		backends[bc.DBName] = bc
	}

	// Return new DSClientConfig instance
	return &DSClientConfig{
		Reference: ref,
		FilePath:  filePath,
		Backends:  backends,
	}
}
