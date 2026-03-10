# Domus

Domus is the data-service foundation of the Kubex ecosystem.

It is responsible for provisioning local data infrastructure, bootstrapping the core database schema, exposing a typed datastore client for consuming applications, and centralizing the data model conventions used by projects such as `GNyx`.

Today, Domus is not just a folder of SQL files or a set of models. It is a combined CLI, bootstrap engine, provider abstraction, and datastore layer that already powers real local development and integration flows.

> Current posture: functional and operational, but still evolving. PostgreSQL is the most mature runtime path today. MongoDB, Redis, and RabbitMQ already exist in the architecture and local stack model, but their typed datastore surface is not yet at the same maturity level as PostgreSQL.

## Table of Contents

- [What Domus Is](#what-domus-is)
- [Current Product Scope](#current-product-scope)
- [Current Operational Status](#current-operational-status)
- [Core Capabilities](#core-capabilities)
- [Architecture Overview](#architecture-overview)
- [Repository Layout](#repository-layout)
- [Runtime Model](#runtime-model)
- [Configuration and Runtime Home](#configuration-and-runtime-home)
- [CLI Overview](#cli-overview)
- [Real Local Development Flow](#real-local-development-flow)
- [Embedded Schema and Data Domains](#embedded-schema-and-data-domains)
- [Typed Store Surface](#typed-store-surface)
- [Using Domus as a Go Dependency](#using-domus-as-a-go-dependency)
- [Ecosystem Role](#ecosystem-role)
- [Current Limitations](#current-limitations)
- [Roadmap Direction](#roadmap-direction)
- [Screenshots](#screenshots)
- [License](#license)

## What Domus Is

Domus is the data layer runtime for Kubex projects.

In practical terms, it currently covers four responsibilities:

1. Provision local infrastructure for data backends, currently centered on Docker-based flows.
2. Bootstrap and migrate the core schema used by the ecosystem.
3. Expose a typed datastore client (`client/`) that other applications can consume.
4. Provide a domain-aligned schema foundation for multi-tenant access, invitations, sessions, integrations, and business entities.

Domus should be read as a platform component, not as a standalone SaaS product UI. Its main consumers are backend applications and integration services that need a stable data-service boundary.

## Current Product Scope

As of the current codebase state, Domus is strongest in these areas:

- local PostgreSQL provisioning and migration bootstrap
- embedded SQL schema orchestration
- typed datastore access for a subset of core entities
- Docker-based runtime orchestration via the `dockerstack` backend
- foundational support for multi-tenant access structures and integration-oriented schema

Domus also contains architectural support for:

- MongoDB
- Redis
- RabbitMQ
- provider-style backend abstraction
- adaptive services and repository adapters

However, these broader capabilities are not all equally consolidated yet. The actual typed datastore surface exposed to consumers is still intentionally narrower than the full schema or infrastructure ambition.

## Current Operational Status

Operationally, the real and repeatedly used flow today is centered on:

- `domus database migrate -C ./configs/config.json`
- a runtime home under `~/.kubex/domus`
- PostgreSQL as the active data backend
- Docker-managed local services
- embedded schema bootstrap that creates the current core tables if needed

That means Domus is already functioning as a real development and integration foundation, even though some parts of the architecture still represent forward-looking expansion rather than equally mature surface area.

## Core Capabilities

### 1. Local data infrastructure bootstrap

Domus can start and manage local backend services through its provider/backend model, currently with `dockerstack` as the effective operational path.

Supported at the infrastructure layer:

- PostgreSQL
- MongoDB
- Redis
- RabbitMQ

### 2. Embedded schema bootstrap

The repository ships embedded SQL bootstrap stages under `internal/bootstrap/embedded/core/`, including:

- multi-tenancy primitives (`org`, `tenant`, `team`)
- users and RBAC (`user`, `role`, `permission`, `role_permission`)
- memberships (`tenant_membership`, `team_membership`)
- invitations (`partner_invitation`, `internal_invitation`)
- auth sessions (`auth_sessions`)
- pending access requests (`pending_access_requests`)
- integration engine tables (`integration_config`, `sync_job`, `integration_query`)

### 3. Typed datastore client

The `client/` package exposes a DS client that can be embedded in other Go applications to:

- load a Domus root config
- initialize backend connections
- retrieve typed stores
- create repository adapters

### 4. Store registry and factory model

Domus maintains an internal store registry keyed by driver and store name. This allows consumer applications to resolve stores dynamically while still using typed helpers for common domains.

### 5. Ecosystem-aligned data modeling

The project already encodes the access and tenancy model used by `GNyx` and related services, even where the consuming applications still complete some workflows with local SQL composition.

## Architecture Overview

At a high level, Domus is organized like this:

```text
CLI / Module Layer
  -> cmd/, internal/module/

Bootstrap + Provider Layer
  -> internal/bootstrap/
  -> internal/provider/
  -> internal/backends/dockerstack/

Connection / Engine Layer
  -> internal/engine/
  -> internal/types/
  -> internal/execution/

Datastore Layer
  -> internal/datastore/
  -> client/

Embedded Schema + Models
  -> internal/bootstrap/embedded/core/
  -> internal/model/
  -> types/models/
```

The current execution model is pragmatic:

- use a JSON root config to describe enabled backends
- initialize the Docker-based backend provider
- bring the required services up
- wait for readiness
- run embedded migrations
- expose typed connection/store access to consuming Go services

## Repository Layout

```text
cmd/                           Cobra CLI entrypoints
client/                        Public DS client and typed store helpers
config/                        Project-local config artifacts
configs/                       Example/used runtime config files
internal/bootstrap/            Embedded bootstrap orchestration and SQL stages
internal/backends/dockerstack/ Active local backend provider
internal/provider/             Provider abstractions and contracts
internal/engine/               Connection manager and config bootstrap logic
internal/datastore/            Store registry, factories, and concrete stores
internal/model/                Internal domain models
types/                         Exported/shared type helpers
support/                       Build, install, validation, and docs scripts
```

## Runtime Model

Domus currently operates in two complementary modes.

### CLI/runtime mode

This is the operational mode used in local development and integration setup:

- read config
- provision/attach local services
- run migrations
- leave the data layer ready for consuming applications

### Library/client mode

This is the embedded mode used by other Go applications:

- initialize a DS client from config
- resolve connections by logical database name
- request typed stores or repository adapters
- use Domus as a single integration point for data access

## Configuration and Runtime Home

The active runtime home is expected to live under:

```text
~/.kubex/domus
```

Typical structure:

```text
~/.kubex/domus/
├── config/
│   └── config.json
└── volumes/
    └── postgresql/
        └── init/
```

Important operational expectations:

- `~/.kubex/domus` should be treated as active runtime state.
- Missing config should be materialized there when bootstrapping a fresh environment.
- Existing runtime files should not be destructively overwritten during repeated or parallel executions.
- Volume directories inside `~/.kubex/domus/volumes` are runtime state and should generally be treated as managed operational storage, not as documentation targets.

Relevant default path constants currently point to:

- `KUBEX_DOMUS_CONFIG_PATH`
- `$HOME/.kubex/domus/config/config.json`

There are also older fallback traces in the codebase referencing `.gnyx`-style config paths. The intended active runtime home for current work should be read as `~/.kubex/domus`.

## CLI Overview

The root CLI is implemented with Cobra and currently exposes these top-level command groups:

- `domus docker`
- `domus database`
- `domus utils`
- `domus ssh`
- `domus config`
- `domus version`

The most relevant path today is `database`.

### Database commands

Representative commands:

```bash
domus database migrate -C ./configs/config.json
domus database start --config-file config.yaml
domus database stop --config-file config.yaml
domus database status --config-file config.yaml
```

### Docker commands

Representative commands:

```bash
domus docker start
domus docker restart
domus docker logs
domus docker list
domus docker list-volumes
```

Not every CLI area has the same maturity level. The most reliable and operationally relevant command path today is still the migration/bootstrap flow.

## Real Local Development Flow

The repeatedly used local flow currently looks like this:

```bash
go fmt ./...
go vet ./...
go build -v ./...
go mod tidy
make build-dev
domus database migrate -C ./configs/config.json
```

In practice, this does the following:

1. validates and builds the project
2. uses the configured backend stack
3. initializes Docker-backed services when needed
4. waits for database readiness
5. executes the migration/bootstrap pipeline
6. leaves the schema ready for consumers such as `GNyx`

If the target schema already exists, the pipeline behaves conservatively and skips redundant bootstrap work.

## Embedded Schema and Data Domains

The embedded core schema is organized in numbered stages.

### Foundational access and tenancy

- `etapa_1_extensions_tenancy.sql`
- `etapa_2_users_rbac.sql`
- `etapa_3_memberships.sql`
- `etapa_4_invites.sql`
- `etapa_9_auth_sessions.sql`
- `etapa_10_pending_access_requests.sql`

This gives Domus a real foundation for:

- organizations and tenants
- teams
- users
- RBAC roles and permissions
- tenant and team memberships
- invitation flows
- refresh-session storage
- pending access workflows

### Business and integration layer

Additional bootstrap stages cover:

- business entities
- indices and triggers
- seed data
- integration engine tables

The integration engine stage currently introduces:

- `integration_config`
- `sync_job`
- `integration_query`

This is important because it means Domus is already carrying the schema foundation for the integration-oriented features being consumed or planned elsewhere in the ecosystem.

## Typed Store Surface

The current typed store surface exposed through the public DS client is intentionally narrower than the full schema.

Today, the strongest PostgreSQL-backed stores are:

- `user`
- `invite`
- `company`
- `pending_access_request`

The client package provides typed helpers such as:

- `GetUserStore(...)`
- `GetInviteStore(...)`
- `GetCompanyStore(...)`
- `GetPendingAccessStore(...)`

This distinction matters:

- the schema is broader than the typed store API
- the infrastructure ambition is broader than the current typed store API
- the consumer-safe surface is therefore smaller than the total architectural footprint

That is not a bug in the README. It is the current reality of the codebase.

## Using Domus as a Go Dependency

Consumer applications can use Domus as a library.

Minimal example:

```go
package main

import (
    "context"

    dsclient "github.com/kubex-ecosystem/domus/client"
    logz "github.com/kubex-ecosystem/logz"
)

func main() {
    ctx := context.Background()
    logger := logz.GetLoggerZ("example")

    cfg := dsclient.NewDSClientConfig(
        "domus",
        "$HOME/.kubex/domus/config/config.json",
    )

    client := dsclient.NewDSClient(ctx, cfg.FilePath, cfg, logger)
    if err := client.Init(ctx); err != nil {
        panic(err)
    }
    defer client.Close(ctx)

    userStore, err := client.GetUserStore(ctx, "domus")
    if err != nil {
        panic(err)
    }

    _ = userStore
}
```

Domus also exposes adapter helpers for mixed store/ORM repository patterns where consumer applications need a transition layer instead of a pure store-only integration.

## Ecosystem Role

Domus sits inside a broader system, but it should be understood on its own terms.

### In relation to GNyx

`GNyx` uses Domus as its data-service boundary for core domains such as:

- users
- invitations
- pending access
- company-like tenancy references

At the moment, `GNyx` still completes some access workflows locally with direct SQL around memberships and role lookups. That means the Domus schema is already foundational, while parts of the full application behavior are still being consolidated at the service boundary.

### In relation to Kbx

`Kbx` supplies cross-project helpers, utilities, security primitives, and shared infrastructure abstractions. Domus depends on this layer for parts of configuration, defaults, and supporting services.

### In relation to Logz

`Logz` is the shared logging foundation used throughout the runtime.

## Current Limitations

Domus is functional, but the boundary between what is architecturally present and what is fully consolidated must be stated clearly.

Current limitations include:

- PostgreSQL is the only clearly consolidated typed datastore path today.
- MongoDB, Redis, and RabbitMQ exist in the provider/runtime model, but not with an equivalently mature typed store surface.
- Some config/default path logic still carries older path conventions alongside the intended `~/.kubex/domus` runtime home.
- The typed store registry does not yet expose the full breadth of the multi-tenant/RBAC schema.
- Some CLI areas exist but are not as mature or operationally important as the migration/bootstrap flow.
- The `company` surface does not fully eliminate broader tenancy-model ambiguity in downstream consumers.

## Roadmap Direction

The practical direction for Domus is clear.

### Near-term consolidation

- strengthen the DS client as the preferred integration point
- keep PostgreSQL bootstrap and store surface stable
- expand typed store coverage for access and tenancy domains where needed
- keep local runtime behavior safe and non-destructive in `~/.kubex/domus`

### Mid-term expansion

- deepen the multi-tenant and RBAC service boundary
- formalize more of the schema as consumer-safe stores
- support broader PostgreSQL-compatible deployments such as Supabase-backed Postgres

### Future infrastructure expansion

- treat Redis and RabbitMQ as explicit, use-case-driven expansions
- avoid expanding backend abstractions without a clear consumer path
- continue separating “available in architecture” from “consolidated in public surface”

## Screenshots

Placeholders for future documentation assets:

- `[Placeholder] CLI migration flow screenshot`
- `[Placeholder] Runtime home structure screenshot`
- `[Placeholder] Schema/bootstrap flow diagram`
- `[Placeholder] DS client usage example diagram`

## License

This repository is licensed under the [MIT License](./LICENSE).
