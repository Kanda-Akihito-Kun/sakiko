<p align="center">
  <img src="sakiko-wails/frontend/public/sakiko.png" alt="Sakiko icon" width="128" />
</p>

<p align="center">
  <a href="./README.md">
    <img src="https://img.shields.io/badge/README-%E7%AE%80%E4%BD%93%E4%B8%AD%E6%96%87-1677ff?style=for-the-badge" alt="README in Simplified Chinese" />
  </a>
</p>

# Sakiko

Web Demo: `http://43.167.194.92:8080`

Desktop proxy subscription testing tool.

Version: `v0.1.0`

## Features

- Full protocol support with `ping`, speed test, inbound/outbound geo detection, NAT detection, and media unlock detection.
- Export test results as images.
- Support custom Speedtest targets and other configurable test options.
- Support selecting which nodes participate in a test.

## Composition

- Protocol library: `mihomo`
- References: `miaospeed`, `clash-verge-rev`, `Speed-Stair`, `RegionRestrictionCheck`, `SSRSpeedN`

## Developer Notes

- Stack:
  - `sakiko-core`: reusable Go kernel
  - `sakiko-wails`: Wails v3 desktop app
  - `frontend`: React + TypeScript + MUI
- Runtime dependencies:
  - Go `1.26+`
  - `pnpm`
  - Wails v3 toolchain
- Layout:

```text
sakiko/
  sakiko-core/
  sakiko-wails/
    frontend/
```

## Development

```powershell
cd .\sakiko-wails\frontend
pnpm install

cd ..
wails3 dev -config .\build\config.yml
```

## Desktop Build

```powershell
cd .\sakiko-wails
wails3 task build PRODUCTION=true
```

## Windows Server Mode Build

Shortcut:

```powershell
cd .\sakiko-wails
wails3 task build:server
```

Manual build:

```powershell
cd .\sakiko-wails\frontend
pnpm install
pnpm build

cd ..
go build -tags server -o .\bin\Sakiko-server.exe
```

Run:

```powershell
$env:SAKIKO_SERVER_HOST="0.0.0.0"
$env:SAKIKO_SERVER_PORT="8080"
.\bin\Sakiko-server.exe
```

## License

MIT
