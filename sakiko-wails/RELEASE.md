# Release Guide

This project is released with GitHub Actions using native runners:

- `windows-latest`: portable zip + NSIS installer
- `macos-latest`: universal `.app` bundle packed as zip
- `ubuntu-24.04`: AppImage

The workflow file is:

- [release.yml](../.github/workflows/release.yml)

## 1. Prerequisites

Before you publish, update the app metadata:

- [build/windows/info.json](./build/windows/info.json)
- [build/darwin/Info.plist](./build/darwin/Info.plist)
- [build/linux/nfpm/nfpm.yaml](./build/linux/nfpm/nfpm.yaml)

Also make sure your icon assets are final:

- [build/appicon.png](./build/appicon.png)

## 2. Release Trigger

The pipeline runs when you push a Git tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

It also supports manual trigger through `workflow_dispatch`.

## 3. Windows Signing

For Windows, you need a code-signing certificate in `.pfx` format.

Recommended path:

- Start with a standard OV code-signing certificate
- Upgrade to EV later if you want better SmartScreen reputation

Add these GitHub repository secrets:

- `WINDOWS_CERT_PFX_BASE64`
- `WINDOWS_CERT_PASSWORD`

Generate the base64 value locally:

```powershell
[Convert]::ToBase64String([IO.File]::ReadAllBytes("D:\path\to\your-cert.pfx"))
```

The workflow signs:

- `bin/Sakiko.exe`
- `bin/*-installer.exe`

If the secrets are absent, the workflow still builds unsigned packages.

## 4. macOS Signing

For macOS distribution outside the App Store, you need:

- Apple Developer Program membership
- A `Developer ID Application` certificate
- Notarization credentials

Add these GitHub repository secrets:

- `MACOS_CERT_P12_BASE64`
- `MACOS_CERT_PASSWORD`
- `MACOS_SIGN_IDENTITY`
- `APPLE_ID`
- `APPLE_APP_SPECIFIC_PASSWORD`
- `APPLE_TEAM_ID`

Generate the base64 value locally:

```bash
base64 -i /path/to/developer-id.p12 | pbcopy
```

Typical identity format:

```text
Developer ID Application: Your Name or Company (TEAMID)
```

The workflow will:

1. import the certificate into a temporary keychain
2. sign `bin/Sakiko.app`
3. notarize the packaged zip
4. staple the notarization ticket back onto the app
5. upload `Sakiko-macos-universal.zip`

If those secrets are absent, the workflow still exports an unsigned macOS zip for internal testing.

## 5. Linux Signing

Linux package signing is optional for the first public release.

Right now the workflow builds:

- `AppImage`

You can add package signing later with a GPG private key for `deb` and `rpm`.

## 6. What Users Download

Recommended release assets:

- Windows:
  - `Sakiko-amd64-installer.exe`
  - `Sakiko-portable-windows-amd64.zip`
- macOS:
  - `Sakiko-macos-universal.zip`
- Linux:
  - `Sakiko-*.AppImage`

## 7. First Dry Run

Before the first public tag:

1. push the workflow to GitHub
2. run it once with `workflow_dispatch`
3. confirm every platform builds
4. add signing secrets
5. tag a pre-release such as `v0.1.0-rc1`
6. verify downloads on real machines

## 8. Notes

- GitHub Actions `Artifacts` are always wrapped by GitHub into an outer download archive. The actual user-facing release files are the files inside that artifact archive, or the direct assets attached to a tag-based GitHub Release.
- macOS packaging is built on a native macOS runner instead of Windows cross-compilation.
- Linux packaging is built on a native Ubuntu runner instead of Docker cross-compilation.
- This keeps the pipeline more stable than trying to produce all targets from a single Windows machine.
