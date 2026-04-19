<p align="center">
  <img src="sakiko-wails/frontend/public/sakiko.png" alt="Sakiko icon" width="128" />
</p>

<p align="center">
  <a href="./README.zh-CN.md">
    <img src="https://img.shields.io/badge/README-%E7%AE%80%E4%BD%93%E4%B8%AD%E6%96%87-1677ff?style=for-the-badge" alt="README in Simplified Chinese" />
  </a>
</p>

# Sakiko

Sakiko is a desktop proxy benchmarking project built around a reusable Go kernel.
The goal is to build a practical benchmark workflow, while keeping the execution core reusable for desktop and future web delivery.

Version: `v0.1.0`

Web Demo Trail: 43.167.194.92:8080

## What Sakiko Func

The current MVP already covers the mainly function:

1. manage subscription
2. inspect nodes
3. run a bunch of benchmark task and check proccess
4archive results and export into picture

Current task presets include:

- `ping`
- `geo`
- `speed`

## Architecture

- `sakiko-core`
  Reusable Go kernel for profile parsing, task execution, result archiving, and report generation
- `sakiko-wails`
  Wails v3 desktop client that consumes `sakiko-core`

The execution model stays centered on, learned from `miaospeed` :

`Vendor -> Macro -> Matrix`

That boundary matters. Business logic lives in Go, not in frontend pages. The desktop app is a consumer of the kernel, not the place where the core benchmarking logic is implemented.

## Repository Layout

```text
sakiko/
  sakiko-core/   reusable kernel
  sakiko-wails/  desktop app
```

### Requirements

- Go `1.26` or newer, and pnpm
- Wails v3 toolchain

The workspace is managed with `go.work` and currently includes:

- `sakiko-core`
- `sakiko-wails`

### Run The Desktop App

```powershell
cd .\sakiko-wails\frontend
pnpm install

cd ..
wails3 dev -config .\build\config.yml
```

## Local Data

The desktop app stores its local data under your OS user config directory in a `sakiko` folder.
That currently includes:

- `profiles.yaml` for the profile index
- `profiles/<profile-id>.yaml` for the original profile content snapshots
- `results/<task-id>.json` for full archived task results
- `results/<task-id>.meta.json` for history list summaries
- `settings.json` for desktop settings

Archived results are meant to be reusable data, not just one-time UI exports.

## Server Mode

`sakiko-wails` also has a `server` build tag for a simple single-machine demo mode, although this is not the main product target.

Build it like this:

```powershell
cd .\sakiko-wails\frontend
pnpm install
pnpm build

cd ..
go build -tags server -o .\bin\Sakiko-server.exe
```

Important caveats:

- this is a local or single-machine deployment mode, not a multi-user web service
- profiles, settings, and result history are stored on the machine that runs the process

## Project References

Sakiko openly learns from several existing projects, but it is not intended to be a one-to-one clone of any single project.

- Backend architecture and execution abstractions reference `miaospeed`
- Frontend information architecture and interaction flow reference `clash-verge-rev`
- Streaming unlock checks and parts of the benchmarking implementation reference `Speed-Stair` and `RegionRestrictionCheck`
- Protocol library is supported by `mihomo-core`

## License

MIT

## Author

鼠鼠今天吃嘉然
