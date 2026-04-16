# Sakiko

`Sakiko` is a personal project by Kanda Akihito (`鼠鼠今天吃嘉然`).
The project started as a desktop-first MVP after `miaospeed` had gone quiet for a long time, with the goal of building a reusable proxy benchmarking kernel and a practical desktop client instead of a one-off clone.

## Project Layout

- `sakiko-core`: reusable Go kernel for subscription parsing, task execution, result archiving, and report generation
- `sakiko-wails`: Wails3 desktop client and the current primary delivery target

## Reference Projects

`Sakiko` openly references several existing projects in different layers of the stack:

- architecture and execution model reference `miaospeed`
- page layout and interaction ideas reference `clash-verge-rev`
- media unlock capability references `RegionRestrictionCheck`

These references guide the design direction, but `Sakiko` keeps its own kernel boundary, report model, and desktop workflow.

## Web Demo Deployment

The desktop client is still the main target.
The web version is currently only a simple demo built on `sakiko-wails` server mode.

Build the demo package locally:

```powershell
cd sakiko-wails\frontend
pnpm install
pnpm build

cd ..
go build -tags server -o bin\Sakiko-server.exe
```

Upload `sakiko-wails/bin/Sakiko-server.exe` to your Windows Server VPS, then start it with:

```powershell
$env:SAKIKO_SERVER_HOST="0.0.0.0"
$env:SAKIKO_SERVER_PORT="8080"
.\Sakiko-server.exe
```

Open the port in Windows Firewall:

```powershell
New-NetFirewallRule -DisplayName "Sakiko Demo 8080" -Direction Inbound -Action Allow -Protocol TCP -LocalPort 8080
```

Then verify:

- `http://127.0.0.1:8080/health`
- `http://your-server-ip:8080/`

Notes:

- this is a single-machine demo, not a multi-user web service
- profiles, settings, and result history are stored on the server machine
- for a public demo, putting `Caddy` or `IIS` in front of it for HTTPS is recommended

## Author

- Author: Kanda Akihito
- Alias: 鼠鼠今天吃嘉然
- GitHub: https://github.com/Kanda-Akihito-Kun
