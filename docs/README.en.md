# Sakiko Developer Documents

This directory is for developers, fork maintainers, and agents. It is not the primary entry for ordinary end users.

## Document Layout

- Chinese overview: [README.zh-CN.md](./README.zh-CN.md)
- English skill: [sakiko.skill](./sakiko.skill)

## Suggested Reading Order

1. root `README.md`
2. this document
3. `sakiko.skill`

## Current Project Rules

- `sakiko-core` is the only business core
- `sakiko-wails` is a desktop consumer layer
- remote capability is centralized in `sakiko-core/cluster`
- result archives are stable product data, not temporary cache

## Main Development Areas

- profile import and node management
- local task execution and archives
- remote `Master / Knight`
- desktop bridge and result export
