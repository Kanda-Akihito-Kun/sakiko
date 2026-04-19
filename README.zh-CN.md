<p align="center">
  <img src="sakiko-wails/frontend/public/sakiko.png" alt="Sakiko 图标" width="128" />
</p>

<p align="center">
  <a href="./README.md">
    <img src="https://img.shields.io/badge/README-English-111111?style=for-the-badge" alt="README in English" />
  </a>
</p>

# Sakiko

Sakiko 是一个以桌面端为主、围绕可复用 Go 内核构建的代理测速项目。
它的目标是提供一套实用的测速工作流，同时保证执行核心在桌面端和未来 Web 形态下都可以复用。

## Sakiko 目前能做什么

当前 MVP 已经覆盖了主要功能链路：

1. 管理订阅
2. 查看节点
3. 批量运行测速任务并查看执行进度
4. 归档结果并导出为图片

当前支持的任务预设包括：

- `ping`
- `geo`
- `speed`

## 架构

- `sakiko-core`
  可复用的 Go 内核，负责订阅解析、任务执行、结果归档和报告生成
- `sakiko-wails`
  基于 Wails v3 的桌面客户端，消费 `sakiko-core`

核心执行模型保持为，且这一部分主要参考自 `miaospeed`：

`Vendor -> Macro -> Matrix`

这条边界非常重要。业务逻辑放在 Go 中，而不是放在前端页面里。桌面端只是内核的消费方，不是核心测速逻辑的实现位置。

## 仓库结构

```text
sakiko/
  sakiko-core/   可复用内核
  sakiko-wails/  桌面应用
  tests/         预留给后续集成测试和发布烟雾测试
```

当前主要交付目标是 `sakiko-wails`。
后续会有 Web 形态的消费者，但它必须复用 `sakiko-core`，而不是重新复制一套业务逻辑。

### 环境要求

- Go `1.26` 或更高版本，以及 `pnpm`
- Wails v3 工具链

当前工作区通过 `go.work` 管理，包含：

- `sakiko-core`
- `sakiko-wails`

### 运行桌面端

```powershell
cd .\sakiko-wails\frontend
pnpm install

cd ..
wails3 dev -config .\build\config.yml
```

## 本地数据

桌面应用会将本地数据存放在当前操作系统用户配置目录下的 `sakiko` 文件夹中。
目前包括：

- `profiles.yaml`：订阅索引文件
- `profiles/<profile-id>.yaml`：原始订阅内容快照
- `results/<task-id>.json`：完整任务归档结果
- `results/<task-id>.meta.json`：历史列表摘要文件
- `settings.json`：桌面端设置

归档结果被设计为可复用数据，而不只是一次性的 UI 导出产物。

## Server Mode

`sakiko-wails` 也提供了 `server` build tag，可用于一个简单的单机 demo 模式，不过这并不是当前的主产品形态。

构建方式如下：

```powershell
cd .\sakiko-wails\frontend
pnpm install
pnpm build

cd ..
go build -tags server -o .\bin\Sakiko-server.exe
```

需要注意：

- 这只是本地或单机部署模式，不是多用户 Web 服务
- 订阅、设置和结果历史都保存在运行该进程的机器上
- 如果需要公网暴露，应该在前面加一个真正的 HTTPS 反向代理

## 项目参考来源

Sakiko 明确参考了多个现有项目，但它并不是任何单一项目的一比一复刻。

- 后端架构和执行抽象参考了 `miaospeed`
- 前端信息架构和交互流程参考了 `clash-verge-rev`
- 流媒体解锁检测和部分测速实现参考了 `Speed-Stair` 与 `RegionRestrictionCheck`
- 协议库能力由 `mihomo-core` 提供支持

这些参考会影响实现方式，但 Sakiko 仍然保持自己的内核边界、报告模型和桌面优先工作流。

## 许可证

MIT

## 作者

鼠鼠今天吃嘉然
