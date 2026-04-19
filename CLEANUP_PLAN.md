# Sakiko Cleanup Plan

本文档整理当前 `sakiko` 仓库中建议优先清理和优化的内容。
目标不是一次性“大重构”，而是先把会影响可维护性、可构建性和数据一致性的点收口，再逐步处理次级问题。

## 优先级说明

- `P0`：会直接影响仓库可用性、协作稳定性或源码完整性，应该最先处理
- `P1`：不会立刻阻塞开发，但会持续制造隐患、技术债或维护成本
- `P2`：结构与工程体验优化，适合在前两层完成后持续推进

## P0

### 1. 修正根 `.gitignore` 对真实源码目录的误伤

- 问题：
  根 `.gitignore` 里存在 `**/results/` 规则，会误伤任何名为 `results` 的目录。
  当前 `sakiko-wails/frontend/src/components/results/` 就被整目录忽略了。
- 风险：
  真实前端源码未进入 git，换机器克隆或协作时会直接缺文件。
- 改造范围：
  - `/.gitignore`
  - `sakiko-wails/frontend/src/components/results/`
- 建议：
  - 只忽略运行期结果目录，不要用会误伤源码的广泛通配规则
  - 改成更精确的路径级忽略，例如仅忽略 `profiles.yaml` 同级生成的 `results/`

### 2. 明确 `third_party` 策略，解决当前构建链路依赖未提交本地目录的问题

- 问题：
  `sakiko-wails/go.mod` 通过 `replace` 指向 `../third_party/wails-v3-alpha76`，但 `third_party/` 当前整体未纳入版本控制。
- 风险：
  fresh clone 后无法直接构建；构建结果依赖本机状态，不可复现。
- 改造范围：
  - `sakiko-wails/go.mod`
  - `/third_party/`
  - 相关 README / 协作说明
- 建议：
  只保留一种策略，不要混用：
  - 方案 A：正式 vendoring，提交 `third_party/wails-v3-alpha76`
  - 方案 B：移除 `replace`，回到远程模块版本管理
  - 如果继续 vendoring，旧版 `wails-v3-alpha74` 应同步清理

### 3. 把当前已被正式引用的测试/构建入口纳入版本控制

- 问题：
  `tests/*.ps1`、`sakiko-wails/build/tools/`、`sakiko-wails/build/ios/` 当前是未跟踪状态，但其中一部分已经被 `README` 或 `Taskfile` 正式引用。
- 风险：
  文档和任务定义与仓库实际内容不一致，后续维护者会遇到“文档里有、仓库里没有”的情况。
- 改造范围：
  - `/tests/`
  - `sakiko-wails/build/tools/`
  - `sakiko-wails/build/ios/`
  - 相关 README
- 建议：
  - 明确哪些属于源码资产，哪些属于本机生成物
  - 被 README / Taskfile 依赖的文件应提交
  - 纯本机产物则应加到 `.gitignore`

## P1

### 4. 统一文件落盘策略，补齐原子写入和一致性收口

- 问题：
  当前 `appsettings` 已使用临时文件 + rename 的原子写法，但 `profiles.yaml`、`profiles/<id>.yaml`、`results/<task-id>.json`、`results/<task-id>.meta.json` 仍是直接 `WriteFile`。
- 风险：
  进程中断、磁盘异常或写入失败时，容易留下半文件、空文件或索引与内容不一致状态。
- 改造范围：
  - `sakiko-core/storage/profiles_yaml.go`
  - `sakiko-core/storage/profile_content_store.go`
  - `sakiko-core/storage/result_store.go`
  - `sakiko-wails/appsettings.go`
- 建议：
  - 抽出统一的原子写 helper
  - 所有关键持久化文件统一走同一套写入策略
  - 让“索引文件 + 内容文件”的一致性约束下沉到 storage 层，而不是只在 manager 层补救

### 5. 补齐 `core/api.Service` 的 nil-guard 一致性

- 问题：
  同一个 API 服务里，大多数方法有“未初始化”保护，但 `ListTasks`、`GetTask`、`RuntimeStatus` 风格不一致，直接访问 `s.kernel`。
- 风险：
  当前虽然主要由上层 `ensureReady()` 遮住，但 API 自身不稳，未来复用时容易埋下 panic 点。
- 改造范围：
  - `sakiko-core/api/service.go`
- 建议：
  - 所有 public 方法统一 nil-guard 语义
  - 明确“未初始化时返回空值”还是“返回错误”，不要混用

### 6. 为前端任务轮询增加单飞保护，避免叠加请求

- 问题：
  当前任务页在 `running` 状态下每 500ms 调一次 `syncActiveTask()`，但没有“上一轮尚未结束则跳过”的保护。
- 风险：
  当一次同步超过 500ms 时，会并发发起多轮 `GetTask + ListTasks`，造成无谓负载和状态抖动风险。
- 改造范围：
  - `sakiko-wails/frontend/src/hooks/useDashboardLifecycle.ts`
  - `sakiko-wails/frontend/src/store/dashboardStore.ts`
- 建议：
  - 加轮询中的互斥标记或 promise in-flight 保护
  - 保持“只对 running 任务持续轮询”的现有约束不变

### 7. 统一“什么应该提交，什么不该提交”的仓库边界

- 问题：
  目前 `frontend/bindings` 作为生成文件被提交，但 `tests/*.ps1`、`build/tools` 这类真实工程入口反而没提交；同时本地缓存、Wails 生成物、可执行文件的忽略规则也不够统一。
- 风险：
  长期会让仓库边界越来越混乱，新增文件时很难判断该提交还是该忽略。
- 改造范围：
  - 根 `.gitignore`
  - `sakiko-wails/.gitignore`
  - `frontend/bindings` 策略
  - `README` / 协作文档
- 建议：
  - 明确分成四类：源码、可重复生成物、必须提交的构建资产、本机缓存
  - 每类给出统一规则，不靠临时判断

## P2

### 8. 清理失效和重复的第三方残留

- 问题：
  当前 `third_party/` 下同时存在旧版 `wails-v3-alpha74`、新版 `wails-v3-alpha76`，还有未被当前构建链路引用的 `go.uber.org` 目录。
- 风险：
  增加仓库体积和认知负担，后续升级时容易误用旧目录。
- 改造范围：
  - `/third_party/`
- 建议：
  - 只保留当前实际使用的第三方目录
  - 保留前先确认是否仍有隐式引用

### 9. 收口重复的路径解析和本地状态初始化逻辑

- 问题：
  `profilesPath`、`settingsPath`、前端初始化刷新、下载目标刷新、历史刷新等流程分散在多个入口，阅读成本较高。
- 风险：
  后续加入自动更新、更多设置项或新页面时，初始化路径和状态刷新容易继续扩散。
- 改造范围：
  - `sakiko-wails/sakikoservice.go`
  - `sakiko-wails/frontend/src/hooks/`
  - `sakiko-wails/frontend/src/store/dashboardStore.ts`
- 建议：
  - 不要重做架构，但可以把初始化职责和轮询职责做更明确的边界划分

### 10. 补一轮工程命名和文档一致性整理

- 问题：
  当前 README、Taskfile、目录结构和 git 跟踪状态之间有轻微漂移。
- 风险：
  随着自动更新、发布和签名链路加入，文档漂移会更快。
- 改造范围：
  - 根 README
  - `sakiko-wails/README.md`
  - `tests/README.md`
  - 相关 Taskfile 注释
- 建议：
  - 保持“文档描述的文件和仓库真实存在的文件”一致
  - 文档只写当前真实可执行的流程

## 推荐执行顺序

1. 先处理 `P0`，保证仓库源码完整、可克隆、可构建
2. 再处理 `P1` 中的持久化和轮询稳健性
3. 最后做 `P2` 的目录瘦身和结构整理

## 本轮不建议直接做的大动作

- 不建议现在就全面重构 `sakiko-core` / `sakiko-wails` 边界
- 不建议把前端状态管理整体替换
- 不建议为“看起来整洁”而一次性搬大量目录

更稳妥的做法是先把仓库卫生、文件一致性和运行期稳健性修好，再考虑更深层的结构优化。
