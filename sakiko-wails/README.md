# sakiko-wails

`sakiko-wails` is the first desktop client of the `sakiko` project.
It uses `Wails3 + React + MUI`, while all core task, profile, archive, and report logic stays in `sakiko-core`.

This README records the current desktop-side implementation status instead of the default Wails template.

## Current Status (2026-04-07)

The desktop app already supports:

- import profile from subscription URL
- refresh and delete profile
- browse imported nodes
- launch `ping / geo / speed / full` tasks
- inspect runtime task results
- browse archived result history
- lazy-load result details by `taskId`
- export archived results as PNG images

The main MVP loop is now complete:

1. import profile
2. run task
3. inspect live result
4. archive task result
5. browse archived history
6. export shareable image

## UI Decisions Already Made

- task-specific shared config no longer stays in the task launcher
- common task options such as `ping url`, `timeout`, `download url`, `duration`, and `thread count` have been moved into Settings
- the task launcher is now focused on preset selection and starting tasks
- the sidebar now contains a dedicated `Results` page under task-related navigation

## Result History

Archived results come from `sakiko-core` and are not front-end-only exports.

Current result page behavior:

- summary list is paged by scroll
- 10 records are loaded at a time
- each record is rendered as a collapsible item
- collapsed state shows:
  - test time
  - profile name
  - tested metrics
  - node count
- expanded state lazy-loads the full archive and shows:
  - task metadata
  - task config snapshot
  - tested node list
  - report sections
  - raw per-node results

## Image Export

Each archived result currently has an `Export` action.

Current export strategy:

- implemented on the frontend with canvas rendering
- layout references the local `sample/` result images
- white-table report style instead of poster/card style
- supports multi-section export for `full` results
- uses the archived report model from `sakiko-core`

Current naming and data rules:

- the old `Type` column has been renamed to `Protocol`
- exported topology rows mask IP values for safer sharing
- export uses the archived report and does not re-run the test

## Architecture Boundary

Keep these boundaries stable:

- `sakiko-core` owns profile parsing, task execution, result archive, and report model
- `sakiko-wails` owns page composition, interaction flow, result browsing, and image export
- frontend state is a consumer of backend contracts, not the source of truth for task history

## Development

Frontend:

```bash
cd frontend
pnpm install
pnpm test
pnpm run build
```

Desktop dev:

```bash
wails3 dev -config ./build/config.yml
```

## Key Files

- `main.go`: Wails app bootstrap
- `sakikoservice.go`: desktop bridge to `sakiko-core`
- `frontend/src/store/dashboardStore.ts`: frontend state and backend calls
- `frontend/src/pages/DashboardPage.tsx`: main desktop shell
- `frontend/src/components/results/ResultsArchivePanel.tsx`: archived result page
- `frontend/src/utils/resultExport.ts`: PNG export renderer
