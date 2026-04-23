# Tests

`Sakiko` keeps Go unit tests next to the package they verify.
That is an intentional choice, not leftover clutter:

- Go package tests need direct access to package-local APIs and fixtures
- moving all `_test.go` files into a single directory would weaken package boundaries and make imports noisier
- frontend tests can be centralized more safely, so `sakiko-wails/frontend/tests` is the preferred location for Vitest files

## Current Layout

- `sakiko-core/**/_test.go`: Go unit tests for the reusable kernel
- `sakiko-wails/*_test.go`: Go bridge tests for the Wails host layer
- `sakiko-wails/frontend/tests`: frontend unit tests

## Run

Run the Go test suites from their module roots:

```powershell
cd .\sakiko-core
go test ./...
```

```powershell
cd .\sakiko-wails
go test ./...
```

Run the frontend tests from the frontend workspace:

```powershell
cd .\sakiko-wails\frontend
pnpm test
```

## Scope

This directory is reserved for future integration tests, release smoke tests, or other cross-package test assets.
If future end-to-end or release smoke tests are added, they should live under `tests/` directly.
