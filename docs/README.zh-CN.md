# Sakiko 开发文档

这个目录面向开发者、二开作者和 Agent，不面向普通用户。

## 文档结构

- 英文开发总览：[README.en.md](./README.en.md)
- 英文 skill：[sakiko.skill](./sakiko.skill)

## 先读什么

建议顺序：

1. 根目录 `README.md`
2. 本文档
3. `sakiko.skill`

## 当前约定

- `sakiko-core` 是唯一业务内核
- `sakiko-wails` 是桌面消费层
- 远程能力集中在 `sakiko-core/cluster`
- 结果归档是正式数据结构，不是临时缓存
- 版本号只人工维护根目录 `VERSION`，其他平台元数据由 `scripts/sync-version.mjs` 同步生成

## 当前主要开发面

- 订阅导入与节点管理
- 本地任务与结果归档
- 远程 `Master / Knight`
- 桌面端桥接与结果导出

## 版本与发布流程

版本号的唯一人工入口是根目录：

```text
VERSION
```

不要手动修改 Windows / macOS / Linux / iOS 的版本元数据，也不要直接改生成文件：

- `sakiko-core/app/version_generated.go`
- `sakiko-wails/frontend/src/constants/version.generated.ts`

升级版本时，在仓库根目录执行：

```powershell
node scripts/sync-version.mjs 0.2.0
```

或者先修改 `VERSION`，再执行：

```powershell
node scripts/sync-version.mjs
```

发布正式版本时推送 tag。GitHub Actions 会以 tag 为准同步版本并自动打包：

```powershell
git tag -a v0.2.0 -m "v0.2.0"
git push origin v0.2.0
```

自动发布流程在 `.github/workflows/release.yml`，tag 构建会产出 Windows portable / installer、Darwin universal zip 和 Linux AppImage。

改动版本流程后建议验证：

```powershell
node scripts/sync-version.mjs
cd sakiko-core
go test ./...
cd ..\sakiko-wails
go test ./...
cd frontend
pnpm test
pnpm run build
```
