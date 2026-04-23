# Sakiko Patch Notes

This local copy of `github.com/wailsapp/wails/v3` is pinned to `v3.0.0-alpha.76`.

It exists to keep Windows `server` mode buildable for `sakiko-wails`.

Patched areas:

- Windows GUI-only files under `pkg/application` are excluded from `server` builds with `&& !server`
- `pkg/application/browser_window.go` adds a no-op `SetScreen` implementation so the server-side browser window satisfies the `Window` interface

This patch should be treated as a temporary compatibility layer until upstream Windows `server` mode is fixed.
