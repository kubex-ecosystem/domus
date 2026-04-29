# Domus

Portuguese (Brazil) version: [docs/README.pt-BR.md](./docs/README.pt-BR.md)

## Table of Contents

- [Overview](#overview)
- [Current Product Scope](#current-product-scope)
- [Current Operational State](#current-operational-state)
- [Core Capabilities](#core-capabilities)
- [Architecture Overview](#architecture-overview)
- [Repository Layout](#repository-layout)
- [Runtime Model](#runtime-model)
- [Configuration and Runtime Home](#configuration-and-runtime-home)
- [Primary Commands](#primary-commands)
- [Embedded Schema](#embedded-schema)
- [Typed Store Surface](#typed-store-surface)
- [External Metadata Registry](#external-metadata-registry)
- [Role in the Ecosystem](#role-in-the-ecosystem)
- [Current Limitations](#current-limitations)
- [Documentation and Notes](#documentation-and-notes)
- [Screenshots](#screenshots)

## Overview

`Domus` is the data-service substrate of the Kubex ecosystem.

It is responsible for:

- provisioning local data infrastructure
- bootstrapping and migrating the core schema
- exposing typed datastore clients and stores
- hosting the active PostgreSQL runtime used by other projects
- serving as the persistent substrate for multi-tenant access and external metadata fronts

`Domus` should be read as a platform component, not as a standalone end-user product.

## Current Product Scope

At the current stage, `Domus` is strongest in:

- PostgreSQL provisioning and migration bootstrap
- Docker-backed local runtime orchestration
- embedded SQL schema management
- typed store exposure for a subset of core entities
- multi-tenant access schema foundations
- session and invitation-related persistence structures
- external metadata registry support for new integration fronts

Architectural support exists for other backends such as MongoDB, Redis, and RabbitMQ, but PostgreSQL remains the most mature and operational runtime path today.

## Current Operational State

The repeatedly used local flow today is centered on:

```bash
domus database migrate -C ./configs/config.json
```

Operational truths today:

- PostgreSQL is the active and proven backend path
- Docker-managed runtime is part of the normal local flow
- embedded bootstrap creates and evolves the current core schema
- `Domus` is already used by `GNyx` as a real data-service boundary
- new metadata-oriented fronts now also rely on Domus as the database substrate

## Core Capabilities

Current concrete capabilities include:

- local database provisioning
- embedded migration/bootstrap orchestration
- typed store factory exposure
- reusable client package for downstream Go consumers
- foundational tables for users, roles, memberships, invites, sessions, and related entities
- additive schema evolution through embedded SQL steps
- registry support for externally loaded metadata datasets

## Architecture Overview

`Domus` is organized around a few major layers:

- `cmd/` for CLI entrypoints
- `internal/bootstrap/embedded/` for schema bootstrap steps
- `internal/backends/` for backend-specific runtime orchestration
- `internal/datastore/` for store interfaces, factories, and implementations
- `client/` for consumer-facing client access
- `configs/` for runtime configuration

The current architecture is deliberately broader than the typed store surface already exposed to consumers.

## Repository Layout

```text
cmd/                               CLI entrypoints
client/                            public client and store exports
configs/                           Domus runtime config
internal/backends/                 backend orchestration, including dockerstack
internal/bootstrap/embedded/       embedded SQL stages and manifest
internal/datastore/                store interfaces and implementations
```

## Runtime Model

The normal local posture is:

- Docker-based local stack
- PostgreSQL as the primary active database
- runtime home under `~/.kubex/domus`
- additive bootstrap/migration through embedded SQL

This makes Domus both a runtime substrate and a schema carrier for higher-level products.

## Configuration and Runtime Home

Domus uses a runtime-home model under:

```text
~/.kubex/domus/
```

This runtime area should be treated as durable operational state.

Important practical rule:

- generated operational state should not be destructively overwritten in repeated or parallel runs

A repo-local configuration flow still drives most local usage:

```bash
./configs/config.json
```

## Primary Commands

Run migrations:

```bash
go run ./cmd/main.go database migrate -C ./configs/config.json
```

Build:

```bash
go build ./...
```

## Embedded Schema

The embedded schema now covers more than the original access core.

Important active areas include:

- users and auth sessions
- roles, permissions, and membership structures
- invitation and pending-access related tables
- broader tenant and business-domain tables
- a registry table for external metadata datasets

Recent additive evolution includes:

- `public.external_metadata_registry`

This table supports registry/governance for external metadata that is loaded into the active PostgreSQL runtime by other tools.

## Typed Store Surface

The typed store surface remains intentionally narrower than the full schema.

It is already meaningful for:

- user-related stores
- invitation-related stores
- company / tenant-related access paths
- pending access structures
- session repository behavior
- external metadata registry access

This narrower typed surface is not a flaw by itself. It is a pragmatic boundary between schema breadth and consumer-facing stability.

## External Metadata Registry

A recent addition to `Domus` is the generic registry for externally loaded metadata datasets.

Current purpose:

- record datasets loaded from external systems
- track where they were materialized in PostgreSQL
- capture readiness, status, and manifest-oriented metadata
- support product/runtime features that depend on grounded metadata availability

Current real use:

- Sankhya BI catalog ingestion into the `sankhya_catalog` schema
- readiness checks consumed by `GNyx` for the BI generation flow

This keeps external ingestion governance inside the active data substrate without forcing Domus itself to become the ingestion tool.

## Role in the Ecosystem

Today `Domus` is the active substrate for:

- `GNyx` access and session-related persistence
- multi-tenant and RBAC schema foundations
- external metadata readiness and registry data
- future expansion of grounded, data-backed product features

It is not just “a future database layer”. It is already in the critical path.

## Current Limitations

Current limitations and constraints include:

- PostgreSQL is far more mature than the other backend paths
- the typed store surface is still smaller than the full schema surface
- some consumers still combine typed stores with targeted SQL composition
- broader platform ambitions exist, but not every path is equally hardened yet

## Documentation and Notes

Useful docs include:

- [`docs/README.pt-BR.md`](./docs/README.pt-BR.md)
- analysis notes in the `GNyx` repository under `.notes/analyzis/`

## Screenshots

Placeholder suggestions:

- `[Screenshot Placeholder: migration output]`
- `[Screenshot Placeholder: PostgreSQL schema view]`
- `[Screenshot Placeholder: external_metadata_registry rows]`
