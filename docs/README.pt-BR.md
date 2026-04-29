# Domus

English version: [../README.md](../README.md)

## Sumário

- [Visão Geral](#visão-geral)
- [Escopo Atual do Produto](#escopo-atual-do-produto)
- [Estado Operacional Atual](#estado-operacional-atual)
- [Capacidades Principais](#capacidades-principais)
- [Visão Geral da Arquitetura](#visão-geral-da-arquitetura)
- [Estrutura do Repositório](#estrutura-do-repositório)
- [Modelo de Runtime](#modelo-de-runtime)
- [Configuração e Runtime Home](#configuração-e-runtime-home)
- [Comandos Principais](#comandos-principais)
- [Schema Embarcado](#schema-embarcado)
- [Superfície de Stores Tipados](#superfície-de-stores-tipados)
- [Registry de Metadados Externos](#registry-de-metadados-externos)
- [Papel no Ecossistema](#papel-no-ecossistema)
- [Limitações Atuais](#limitações-atuais)
- [Documentação e Notas](#documentação-e-notas)
- [Screenshots](#screenshots)

## Visão Geral

`Domus` é o substrate de data-service do ecossistema Kubex.

Ele é responsável por:

- provisionar infraestrutura local de dados
- bootstrapping e migração do schema principal
- expor clients e stores tipados
- hospedar o runtime PostgreSQL ativo usado por outros projetos
- servir como substrate persistente para frentes multi-tenant e de metadados externos

`Domus` deve ser lido como um componente de plataforma, não como um produto final para usuário.

## Escopo Atual do Produto

No estágio atual, o `Domus` é mais forte em:

- provisionamento de PostgreSQL e bootstrap de migrações
- orquestração local baseada em Docker
- gerenciamento de schema por SQL embarcado
- exposição de stores tipados para um subconjunto de entidades centrais
- fundações de schema para acesso multi-tenant
- estruturas persistentes de sessão e convites
- suporte a registry de metadados externos para novas frentes de integração

Existe suporte arquitetural para outros backends como MongoDB, Redis e RabbitMQ, mas o PostgreSQL continua sendo o caminho de runtime mais maduro e operacional hoje.

## Estado Operacional Atual

O fluxo local mais repetido hoje está centrado em:

```bash
domus database migrate -C ./configs/config.json
```

Verdades operacionais atuais:

- PostgreSQL é o backend ativo e comprovado
- o runtime local gerenciado por Docker faz parte do fluxo normal
- o bootstrap embarcado cria e evolui o schema principal atual
- o `Domus` já é usado pelo `GNyx` como boundary real de data-service
- novas frentes orientadas por metadados também passaram a depender do `Domus` como substrate de banco

## Capacidades Principais

Capacidades concretas atuais incluem:

- provisionamento local de banco
- orquestração de migração/bootstrap embarcada
- exposição de factories de stores tipados
- package client reutilizável para consumidores Go
- tabelas fundacionais para usuários, roles, memberships, convites, sessões e entidades relacionadas
- evolução aditiva de schema por etapas SQL embarcadas
- suporte a registry de datasets de metadados externos

## Visão Geral da Arquitetura

O `Domus` se organiza em algumas camadas principais:

- `cmd/` para entrypoints da CLI
- `internal/bootstrap/embedded/` para etapas de bootstrap do schema
- `internal/backends/` para orquestração específica de backend
- `internal/datastore/` para interfaces, factories e implementações de stores
- `client/` para acesso consumível por outras aplicações
- `configs/` para configuração de runtime

A arquitetura atual é propositalmente mais ampla do que a superfície de stores tipados já exposta aos consumidores.

## Estrutura do Repositório

```text
cmd/                               entrypoints da CLI
client/                            client público e exports de stores
configs/                           config de runtime do Domus
internal/backends/                 orquestração de backends, incluindo dockerstack
internal/bootstrap/embedded/       etapas SQL embarcadas e manifest
internal/datastore/                interfaces e implementações de stores
```

## Modelo de Runtime

A postura local normal é:

- stack local baseada em Docker
- PostgreSQL como banco principal ativo
- runtime home em `~/.kubex/domus`
- bootstrap/migração aditiva por SQL embarcado

Isso faz do `Domus` ao mesmo tempo um substrate de runtime e um portador de schema para produtos de nível mais alto.

## Configuração e Runtime Home

O `Domus` usa um modelo de runtime home em:

```text
~/.kubex/domus/
```

Essa área de runtime deve ser tratada como estado operacional durável.

Regra prática importante:

- o estado operacional gerado não deve ser sobrescrito destrutivamente em execuções repetidas ou paralelas

Um fluxo de config local do repositório ainda dirige a maior parte do uso local:

```bash
./configs/config.json
```

## Comandos Principais

Rodar migrações:

```bash
go run ./cmd/main.go database migrate -C ./configs/config.json
```

Compilar:

```bash
go build ./...
```

## Schema Embarcado

O schema embarcado hoje cobre mais do que o núcleo original de acesso.

Áreas ativas importantes incluem:

- users e auth sessions
- roles, permissions e memberships
- tabelas relacionadas a convites e pending access
- tabelas mais amplas de tenant e domínio de negócio
- uma tabela de registry para datasets de metadados externos

Evolução aditiva recente inclui:

- `public.external_metadata_registry`

Essa tabela dá suporte à governança de metadados externos carregados no PostgreSQL ativo por outras ferramentas.

## Superfície de Stores Tipados

A superfície de stores tipados continua intencionalmente menor que a largura total do schema.

Ela já é significativa para:

- stores relacionados a usuários
- stores relacionados a convites
- caminhos de acesso relacionados a company / tenant
- estruturas de pending access
- comportamento do session repository
- acesso ao external metadata registry

Essa superfície tipada mais estreita não é um problema por si só. Ela é um boundary pragmático entre a largura do schema e a estabilidade voltada ao consumidor.

## Registry de Metadados Externos

Uma adição recente ao `Domus` é o registry genérico para datasets de metadados externos.

Propósito atual:

- registrar datasets carregados de sistemas externos
- rastrear onde eles foram materializados no PostgreSQL
- capturar status, readiness e metadados orientados a manifest
- suportar features de produto/runtime que dependem da disponibilidade de metadados grounded

Uso real atual:

- ingestão do catálogo BI do Sankhya no schema `sankhya_catalog`
- checks de readiness consumidos pelo `GNyx` para o fluxo de geração BI

Isso mantém a governança de ingestão externa dentro do substrate ativo de dados sem forçar o `Domus` a virar a própria ferramenta de ingestão.

## Papel no Ecossistema

Hoje o `Domus` é o substrate ativo para:

- persistência de acesso e sessão do `GNyx`
- fundações de schema para multi-tenant e RBAC
- readiness e registry de metadados externos
- futura expansão de features grounded e sustentadas por dados

Ele não é apenas “uma camada de banco futura”. Ele já está no caminho crítico.

## Limitações Atuais

Limitações e restrições atuais incluem:

- PostgreSQL é muito mais maduro do que os outros backends
- a superfície de stores tipados ainda é menor que a superfície total do schema
- alguns consumidores ainda combinam stores tipados com composição de SQL pontual
- ambições mais amplas de plataforma existem, mas nem todos os caminhos estão igualmente endurecidos ainda

## Documentação e Notas

Docs úteis incluem:

- [`../README.md`](../README.md)
- notas de análise no repositório `GNyx` sob `.notes/analyzis/`

## Screenshots

Sugestões de placeholders:

- `[Screenshot Placeholder: saída de migração]`
- `[Screenshot Placeholder: visão do schema no PostgreSQL]`
- `[Screenshot Placeholder: linhas de external_metadata_registry]`
