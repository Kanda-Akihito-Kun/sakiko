<p align="center">
  <img src="sakiko-wails/frontend/public/sakiko.png" alt="Sakiko" width="128" />
</p>

<p align="center">
  <a href="./README.en.md">English</a>
</p>

# Sakiko

Sakiko —— 桌面端机场梯子测试工具。

Web Demo: http://43.167.194.92:8080/

<img src="docs/img/hidden-profile_full_2026-05-10_15-17-57_Light.png" alt="hidden-profile_full_2026-05-10_15-17-57_Light.png" style="zoom: 33%;" />


## 特色:

- 桌面端导出测试瓜条
- 支持自选 Speedtest 的测速目标服务器
- 默认开启隐藏敏感信息(机场名字和国内入口信息)
- 远程操控(开发中的实验性内容)
- 更多自定义选项


## 下载指南

在 Release 中有打好的包, 因为是基于 wails3 开发的, 所以支持跨平台 (但必须是有 webview2 的系统, 如 win10 及以上)


## 项目结构

主要分为两部分:

- `sakiko-core`: Go 内核, 负责所有核心逻辑, 如果觉得本项目的 UI 太丑, 可以直接把 core 拉走
- `sakiko-wails`：Wails v3 桌面端, 处理可视化和交互

PS: Wails3 支持 server mode 模式, 可以直接部署到 web 上, 很适合做 demo


## 其他信息与详细文档

### 重要信息:

- 协议库用的 Mihomo
- 代码架构上分为三个部分: sakiko-core / sakiko-wails / sakiko-cli(未开发)

### 详细文档:

- 中文开发总览：[docs/README.zh-CN.md](./docs/README.zh-CN.md)
- English developer overview: [docs/README.en.md](./docs/README.en.md)
- Agent / 二开 skill：[docs/sakiko.skill](./docs/sakiko.skill)


## 许可协议

MIT
