# sakiko-core

`sakiko-core` is the reusable Go kernel of the `Sakiko` project.
It is responsible for node parsing, task scheduling, proxy dialing, test execution, result extraction, and external communication.

This document records the current MVP architecture and implementation status.

`Sakiko` began as a personal MVP after `miaospeed` had gone quiet for a long time.
The current implementation still keeps that original goal: ship a reusable desktop-first kernel for proxy benchmarking without cloning the old project blindly.

## 0. Current MVP Status (2026-04-07)

The codebase now includes a working MVP kernel implementation for:

- task execution (`ping` / `geo` / `udp` / `speed` / `full`), task lifecycle query, and task result retrieval
- in-memory runtime task registry
- profile management in `api/`:
  - import profile from subscription URL
  - list and get profile
  - refresh profile from source URL
- local profile persistence via `profiles.yaml` (YAML file storage)
- local result persistence via `results/<task-id>.json`
- report-oriented archive output with `speed_table`, `topology_table`, and `udp_nat_table`
- node-level automatic retry based on matrix failure detection
- sanitized archive snapshots that exclude sensitive node payload content

Current `task.status` response includes:

- task lifecycle state
- per-node result list after task exit
- task exit code for client-side status recovery

Current result archive includes:

- archived task metadata and config snapshot
- sanitized node list for sharing
- raw per-node results
- report sections for desktop UI and future web rendering

Current profile parser focuses on Clash-compatible YAML input with a `proxies` list.
Each proxy entry is normalized into `interfaces.Node`.

## 1. Positioning

`sakiko-core` is not a UI project and not a thin wrapper around `mihomo`.
It is the execution kernel shared by:

- `sakiko-wails`: local desktop application
- future server-mode consumers
- future CLI / worker / remote agent forms

Core principles:

- business logic stays in Go, not in frontend state
- UI and service layers talk to the kernel through stable application interfaces
- test execution is decomposed into reusable abstractions instead of hard-coded flows
- protocol details should be replaceable without rewriting the scheduler

## 2. Reference Projects And What To Learn

### From `miaospeed`

Learn and reuse these ideas:

- `Vendor`: unified proxy connection interface
- `Macro`: execution action, such as ping or speed
- `Matrix`: result extraction from macro output
- task queue + concurrency scheduling
- per-node progress reporting

Do not copy blindly:

- private protocol coupling
- monolithic server assumptions
- product-specific API semantics

### From `clash-verge-rev`

Learn and reuse these ideas:

- clear module boundaries
- separate config, runtime, feature, and core lifecycle responsibilities
- keep integration entrypoints thin and push logic downward

### From `RegionRestrictionCheck`

Learn and reuse these ideas:

- media unlock probing as a dedicated capability instead of UI-only decoration
- explicit platform-by-platform result modeling for report generation
- keep unlock checks separable from the rest of the execution pipeline

### From `GUI.for.Clash`

Learn and reuse these ideas:

- bridge/service style boundary between UI and backend
- backend owns side effects, process control, and resource access
- frontend should consume stable service contracts instead of internal structs

## 3. First-Phase Goal

The first stage only targets a usable kernel MVP.

Scope:

- parse node payloads and build vendors
- run ping and speed tasks
- expose task lifecycle and result streaming
- import / refresh / list / get profiles through `api`
- be embeddable by desktop app and later reusable by web service

Out of scope for the first stage:

- multi-user tenancy
- distributed workers
- full script marketplace
- complex dashboard UI concerns

## 4. Proposed Architecture

The kernel is planned as layered modules:

### 4.1 Domain And Contracts

- `interfaces/`
  - stable cross-module contracts
  - task, node, result, vendor, macro, matrix, transport interfaces
- `profiles/`
  - subscription, profile, node domain models
  - parsing and normalization
This layer should stay simple and dependency-light.

### 4.2 Execution Kernel

- `executor/`
  - task scheduling
  - serial / parallel execution engine
  - progress callbacks
  - queue isolation
- `executor/taskpoll/`
  - worker queues
  - weighted scheduling
  - timeout and cancellation primitives
- `kernel/`
  - application-facing orchestration service
  - lifecycle, task registry, runtime status

This is the heart of `sakiko-core`.

### 4.3 Test Capability Layer

- `vendors/`
  - `mihomo`
  - `local`
  - later extensible vendor adapters
- `macro/`
  - `ping`
  - `speed`
  - later `udp`, `geo`, `script`
- `matrix/`
  - `rttping`
  - `httpping`
  - `averagespeed`
  - `maxspeed`
  - `persecondspeed`
- `netx/`
  - low-level HTTP / TCP / UDP helpers
  - reusable transport utilities

This layer should follow the `Vendor -> Macro -> Matrix` model from `miaospeed`, but keep naming and contracts aligned with `Sakiko`.

### 4.4 External Access Layer

- `api/`
  - application service facade
  - submit tasks, query progress, list runtime status
  - import / list / get / refresh profiles
Important rule:

- external callers should not call `executor`, `vendor`, or `macro` directly
- all outward access goes through `api`

## 5. Suggested Directory Layout

Planned structure:

```text
sakiko-core/
  README.md
  go.mod
  api/
  executor/
    taskpoll/
  interfaces/
  kernel/
  macro/
    ping/
    speed/
    udp/
    geo/
    script/
  matrix/
    rttping/
    httpping/
    averagespeed/
    maxspeed/
    persecondspeed/
  netx/
  profiles/
  storage/
  vendors/
    mihomo/
    local/
    invalid/
```

Notes:

- `storage/` is reserved early even if phase 1 only writes local YAML / JSON
- current phase writes `profiles.yaml` using the `storage.ProfileStore`
- `api/` is the stable facade that both desktop and future server-mode consumers should depend on

## 6. Communication Plan

Current MVP uses the local desktop facade in `api/`.
If server mode is added later, the transport and authentication model should be designed around the same application boundary instead of leaking scheduler internals.

## 7. Runtime Model

The kernel should separate task types by runtime cost:

- connection-like tasks
  - ping
  - HTTP delay
  - simple capability probes
- heavy throughput tasks
  - download speed
  - later upload speed

Why:

- speed tests are much heavier than ping tests
- they should not starve lightweight requests
- desktop and web deployment need different concurrency knobs

Planned runtime knobs:

- connection concurrency
- speed concurrency
- speed interval
- task timeout
- vendor-level timeout
- queue length limit

## 8. Data Model Direction

The stable data model should revolve around these entities:

- `Profile`
  - subscription source or local collection
- `Node`
  - normalized proxy node
- `Task`
  - one submitted test job
- `EntryResult`
  - per-node execution result
- `MatrixResult`
  - extracted metric value
- `TaskState`
  - queueing / running / finished / failed / canceled

Persistence direction:

- profiles stored as local YAML
- runtime snapshots stored as JSON
- current result storage format is one file per task under `results/`, named by `task-id`
- later history storage can upgrade to sqlite without changing core interfaces

## 9. Phase Plan

### Phase 1: Kernel MVP

- basic module skeleton
- `mihomo` vendor
- `local` vendor
- ping matrices
- speed matrices
- in-memory task registry
### Phase 2: Rich Test Capabilities

- UDP tests
- richer geo and outbound identity tests
- script-driven probes
- historical result querying and richer report sections
- better observability and structured logs

## 12. What Is Already Landed

Implemented and in daily use now:

- result archive model in `interfaces/result_archive.go`
- filesystem result store in `storage/result_store.go`
- archive save-on-exit in `kernel/service.go`
- result list / get APIs in `api/service.go`
- `geo` report rows that merge task error and matrix-level geo errors
- organization-name fallback that prefers more specific ISP data over generic values like `Private Customer`
- executor retry flow in `executor/poll_item.go`

This means `sakiko-core` is no longer only a transient execution kernel.
It now also owns the first stable version of the project's result history model.

### Phase 3: Service-Grade Expansion

- multi-session management
- stronger quotas and isolation
- worker mode
- web-oriented orchestration

## 10. Decisions To Keep Stable Early

These decisions should be locked early to avoid churn:

- `Vendor -> Macro -> Matrix` remains the test abstraction model
- `api/` is the only official outward application boundary
- future server-mode access should stay behind a dedicated transport layer rather than leaking executor internals
- `sakiko-wails` is a consumer of `sakiko-core`, not a place for duplicated business logic

## 11. Immediate Next Step After Approval

If this README direction is approved, implementation should begin in this order:

1. initialize `go.mod` and base package layout
2. define `interfaces` and task/report contracts first
3. implement `executor` and `kernel` lifecycle
4. implement `vendors/mihomo` and `netx`
5. implement `macro/ping`, `macro/speed`, and matching matrices
6. expose `api`

This order keeps protocol, scheduling, and capability boundaries stable before adding more features. 
