package types

import (
	"context"
	"reflect"
)



// SchemaType define a interface base para todos os stores/repositories.
// Fornece métodos para introspecção e validação em runtime.
type SchemaType interface {
	// GetName retorna o nome identificador do store (ex: "pg_user_store").
	GetName() string

	// Validate valida o estado interno do store (ex: executor não-nil).
	// Retorna erro se o store está em estado inválido.
	Validate() error

	// Close libera recursos do store se necessário.
	// Implementações que não precisam liberar recursos devem retornar nil.
	Close() error
}

// StoreType define a interface base que todos os stores/repositories devem implementar.
// Fornece métodos para introspecção e validação em runtime.
type StoreType interface {
	SchemaType

	// GetType retorna o reflect.Type da entidade gerenciada, nome do store e erro (se houver).
	// Usado para discovery e validação em runtime.
	GetType() (reflect.Type, string, error)
}

// StoreMetadata contém metadados sobre um store registrado.
type StoreMetadata struct {
	Name         string       // Nome do store (ex: "user")
	Type         reflect.Type // Tipo da entidade (ex: reflect.TypeOf(User{}))
	DriverName   string       // Nome do driver (ex: "postgres")
	Description  string       // Descrição opcional
	Version      string       // Versão do schema/store
	Capabilities []string     // Capacidades (ex: ["transactions", "full_text_search"])
}

// StoreFactory define interface para criar stores de forma dinâmica.
type StoreFactory interface {
	// Create cria uma instância do store por nome.
	Create(ctx context.Context, name string) (StoreType, error)

	// ListAvailable retorna lista de nomes de stores disponíveis.
	ListAvailable() []string

	// GetMetadata retorna metadados de um store específico.
	GetMetadata(name string) (*StoreMetadata, error)
}

// StoreRegistry define interface para registro global de stores.
type StoreRegistry interface {
	// Register registra um store para um driver específico.
	Register(driverName, storeName string, metadata *StoreMetadata, factory func(context.Context, Driver) (StoreType, error)) error

	// Get retorna a factory function para um store específico.
	Get(driverName, storeName string) (func(context.Context, Driver) (StoreType, error), bool)

	// ListByDriver retorna todos os stores registrados para um driver.
	ListByDriver(driverName string) []string

	// GetMetadata retorna metadados de um store.
	GetMetadata(driverName, storeName string) (*StoreMetadata, error)

	// AllDrivers retorna lista de drivers com stores registrados.
	AllDrivers() []string
}
