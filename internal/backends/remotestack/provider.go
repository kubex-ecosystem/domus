// Package remotestack provides a local Docker-based stack implementation
package remotestack

// Provider is an alias to the real implementation in adapter.go
// This file exists for backward compatibility
type Provider = RemoteStackProvider

// New creates a new remotestack provider instance.
func New() *Provider {
	return NewRemoteStackProvider()
}
