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
- `tests/*.ps1`: centralized local entrypoints for running the main test suites

## Run

From the repository root:

```powershell
.\tests\run-all.ps1
```

Or run each suite separately:

```powershell
.\tests\run-core-tests.ps1
.\tests\run-wails-tests.ps1
.\tests\run-frontend-tests.ps1
```

## Scope

This directory is the centralized test entrypoint, not a dumping ground for all unit test files.
If future end-to-end or release smoke tests are added, they should live under `tests/` directly.
