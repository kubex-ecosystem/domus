// Package backends registers all available backend providers
package backends

import (
	"github.com/faelmori/domus-pulse/internal/backends/dockerstack"
	"github.com/faelmori/domus-pulse/internal/backends/remotestack"
	"github.com/faelmori/domus-pulse/internal/interfaces"
	"github.com/faelmori/domus-pulse/internal/provider"
	"github.com/faelmori/domus-pulse/internal/provider/flavors"

	kbxGet "github.com/kubex-ecosystem/kbx/get"
)

func init() {
	var dockerService interfaces.IDockerService
	// TODO: Initialize dockerService with a concrete implementation if needed
	registerProviders(
		dockerstack.New(dockerService),
		remotestack.New(),
	)
}

func registerProviders(providers ...provider.Provider) {
	for _, p := range providers {
		if p == nil || p.Name() == "" {
			continue
		}
		if _, exists := flavors.Get(p.Name()); exists {
			continue
		}
		flavors.Register(p)
	}
}

func GetProvider(name string) (provider.Provider, bool) { return flavors.Get(name) }
func ListProviders() []provider.Provider                { return flavors.All() }
func defaultProviderName() string                       { return kbxGet.EnvOr("DOMUS_DEFAULT_PROVIDER", "dockerstack") }
func IsDefaultProvider(name string) bool                { return name == defaultProviderName() }

func DefaultProvider() provider.Provider {
	if p, exists := flavors.Get(defaultProviderName()); exists {
		return p
	}
	return nil
}
