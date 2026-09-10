---
name: gpc-setup
description: Set up authentication, configure projects, validate credentials, and manage auth profiles for the Play Console CLI (gpc).
---

# GPC Setup & Configuration

Use this skill when the user wants to set up the gpc CLI, authenticate, manage auth profiles, initialize a project config, troubleshoot issues, or manage team access.

## Prerequisites

- `gpc` CLI installed (via Homebrew: `brew tap AndroidPoet/tap && brew install playconsole-cli`)
- A Google Cloud service account with Play Console API access

## Initial Setup

### Interactive Setup Wizard

```bash
gpc setup
```

### Manual Setup

1. Create service account at [Google Cloud Console](https://console.cloud.google.com/iam-admin/serviceaccounts)
2. Enable API at [Google Play Android Developer API](https://console.cloud.google.com/apis/library/androidpublisher.googleapis.com)
3. Grant access in [Play Console API Settings](https://play.google.com/console/developers/api-access)
4. Store credentials securely:

```bash
mkdir -p ~/.config/gpc
mv ~/Downloads/your-key.json ~/.config/gpc/service-account.json
chmod 600 ~/.config/gpc/service-account.json
```

## Authentication

### Login with Credentials

```bash
gpc auth login --credentials ~/.config/gpc/service-account.json
```

### Login with Base64 Credentials (CI/CD)

```bash
gpc auth login --credentials-b64 "$GPC_CREDENTIALS_B64"
```

### List Auth Profiles

```bash
gpc auth list
```

### Show Current Profile

```bash
gpc auth current
```

### Switch Profile

```bash
gpc auth switch --profile <name>
```

### Delete a Profile

```bash
gpc auth delete --profile <name>
```

## Project Configuration

### Initialize Project Config

```bash
gpc init --package com.example.app
gpc init --package com.example.app --force  # Overwrite existing
```

Creates `.gpc.yaml` in the current directory.

### Validate Setup

```bash
gpc doctor           # Quick check
gpc doctor --verbose # Detailed diagnostics
```

## Team Management

```bash
gpc users list
gpc users grant --email "dev@company.com" --role releaseManager
gpc users revoke --email "dev@company.com"
```

Roles: `admin`, `releaseManager`, `appOwner`.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GPC_CREDENTIALS_PATH` | Path to service account JSON |
| `GPC_CREDENTIALS_B64` | Base64-encoded credentials (for CI) |
| `GPC_PACKAGE` | Default package name |
| `GPC_PROFILE` | Auth profile to use |
| `GPC_OUTPUT` | Default output format |

## Shell Completions

```bash
gpc completion bash > /etc/bash_completion.d/gpc
gpc completion zsh > "${fpath[1]}/_gpc"
gpc completion fish > ~/.config/fish/completions/gpc.fish
gpc completion powershell > gpc.ps1
```

## App Recovery

```bash
gpc recovery list
gpc recovery create --file recovery-action.json
gpc recovery deploy --action-id <id>
gpc recovery cancel --action-id <id>
```

## Edit Sessions (Advanced)

```bash
gpc edits create
gpc edits get
gpc edits validate
gpc edits commit
gpc edits delete
```

## Global Flags

All commands support: `--package/-p`, `--output/-o` (json/table/tsv/csv/yaml/minimal), `--pretty`, `--quiet/-q`, `--debug`, `--dry-run`, `--timeout`, `--config`, `--profile`.
