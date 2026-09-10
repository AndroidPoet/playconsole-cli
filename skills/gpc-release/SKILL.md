---
name: gpc-release
description: Upload Android app bundles, manage release tracks, promote releases, and handle staged rollouts using the Play Console CLI (gpc).
---

# GPC Release Management

Use this skill when the user wants to upload an Android app bundle, manage release tracks, promote a release between tracks, perform staged rollouts, halt a release, or upload deobfuscation/mapping files.

## Prerequisites

- `gpc` CLI installed and authenticated (`gpc auth login`)
- Package name configured via `--package`, `.gpc.yaml`, or `GPC_PACKAGE` env var

## Commands

### Upload a Bundle

```bash
gpc bundles upload --file <path-to-aab> --track <track>
```

- `--file` (required): Path to the `.aab` file
- `--track`: Target track (internal, alpha, beta, production). Defaults to internal if omitted.

### Find a Bundle by Version Code

```bash
gpc bundles find --version-code <code>
```

### Wait for Bundle Processing

```bash
gpc bundles wait --version-code <code> --timeout 5m
```

Polls until Google Play finishes processing. Useful in CI pipelines after upload.

### List Bundles

```bash
gpc bundles list
```

### List Tracks

```bash
gpc tracks list
```

### Get Track Details

```bash
gpc tracks get --track <track>
```

### Promote Between Tracks

```bash
gpc tracks promote --from <source-track> --to <target-track> --rollout <percentage>
```

- `--rollout`: Percentage (1-100) for staged rollout. Omit for full rollout.

### Update Rollout Percentage

```bash
gpc tracks update --track <track> --rollout <percentage>
```

### Halt a Release (Emergency)

```bash
gpc tracks halt --track <track>
```

Immediately stops the rollout on the specified track.

### Complete a Rollout

```bash
gpc tracks complete --track <track>
```

### Upload Deobfuscation Files

```bash
# ProGuard/R8 mapping file
gpc deobfuscation upload --version-code <code> --file mapping.txt

# Native debug symbols
gpc deobfuscation upload --version-code <code> --file native-debug-symbols.zip --type nativeCode
```

### Upload Legacy APK

```bash
gpc apks upload --file <path-to-apk>
```

## Typical Release Flow

1. Upload: `gpc bundles upload --file app.aab --track internal`
2. Wait: `gpc bundles wait --version-code 42`
3. Test internally, then promote: `gpc tracks promote --from internal --to beta`
4. Staged production: `gpc tracks promote --from beta --to production --rollout 10`
5. Increase rollout: `gpc tracks update --track production --rollout 50`
6. Complete: `gpc tracks complete --track production`

## Global Flags

All commands support: `--package/-p`, `--output/-o` (json/table/tsv/csv/yaml/minimal), `--pretty`, `--quiet/-q`, `--debug`, `--dry-run`, `--timeout`, `--profile`.
