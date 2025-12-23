package factory

import (
	prv "github.com/kubex-ecosystem/domus/internal/provider"
	flv "github.com/kubex-ecosystem/domus/internal/provider/flavors"
)

type ServiceRef = prv.ServiceRef
type Endpoint = prv.Endpoint
type Capabilities = prv.Capabilities
type StartSpec = prv.StartSpec
type Provider = prv.Provider

const (
	EnginePostgres = prv.EnginePostgres
	EngineMongo    = prv.EngineMongo
	EngineRedis    = prv.EngineRedis
	EngineRabbit   = prv.EngineRabbit
)

type Engine = prv.Engine

func Register(p Provider)              { flv.Register(p) }
func Get(name string) (Provider, bool) { return flv.Get(name) }
func All() []Provider                  { return flv.All() }
