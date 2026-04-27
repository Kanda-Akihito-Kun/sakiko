<p align="center">
  <img src="sakiko-wails/frontend/public/sakiko.png" alt="Sakiko 图标" width="128" />
</p>

<p align="center">
  <a href="./README.en.md">
    <img src="https://img.shields.io/badge/README-English-111111?style=for-the-badge" alt="README in English" />
  </a>
</p>

# Sakiko

Web Demo：`http://43.167.194.92:8080`

桌面端机场订阅测试工具。

版本：`v0.1.0`

## 特色

- 支持基础的全协议测试，以及 `ping`、测速、出入口检测、NAT 检测、流媒体解锁检测。
- 支持结果图片导出。
- 支持自选 Speedtest 节点作为目标测速节点，以及其他可配置测试选项。
- 支持自选参与测速的节点。

## 构成

- 协议库：`mihomo`
- 参考项目：`miaospeed`、`clash-verge-rev`、`Speed-Stair`、`RegionRestrictionCheck`、`SSRSpeedN`

## 开发者介绍

- 技术框架：
  - `sakiko-core`：可复用 Go 内核
  - `sakiko-wails`：Wails v3 桌面端
  - `frontend`：React + TypeScript + MUI
- 运行依赖：
  - Go `1.26+`
  - `pnpm`
  - Wails v3 工具链
- 目录结构：

```text
sakiko/
  sakiko-core/
  sakiko-wails/
    frontend/
```

## 开发运行

```powershell
cd .\sakiko-wails\frontend
pnpm install

cd ..
wails3 dev -config .\build\config.yml
```

## 桌面端构建

```powershell
cd .\sakiko-wails
wails3 task build PRODUCTION=true
```

## Windows Server Mode 打包

快捷命令：

```powershell
cd .\sakiko-wails
wails3 task build:server
```

手动打包：

```powershell
cd .\sakiko-wails\frontend
pnpm install
pnpm build

cd ..
go build -tags server -o .\bin\Sakiko-server.exe
```

运行：

```powershell
$env:SAKIKO_SERVER_HOST="0.0.0.0"
$env:SAKIKO_SERVER_PORT="8080"
.\bin\Sakiko-server.exe
```

## License

MIT
