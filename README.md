# Play Console CLI

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-blue?style=for-the-badge" alt="Platform">
  <img src="https://img.shields.io/github/stars/AndroidPoet/playconsole-cli?style=for-the-badge" alt="Stars">
</p>

<p align="center">
  <b>Ship Android apps from your terminal. No browser. No clicking. Just code.</b>
</p>

<p align="center">
  <code>playconsole-cli bundles upload --file app.aab --track production</code>
</p>

---

A **fast**, **lightweight**, and **scriptable** CLI for Google Play Console. Built for developers who automate everything.

> Inspired by [App Store Connect CLI](https://github.com/rudrankriyam/App-Store-Connect-CLI) — the same philosophy, now for Android.

## Why This Exists

| The Old Way | The GPC Way |
|-------------|-------------|
| Open browser, click through menus | `gpc bundles upload --track internal` |
| Wait for slow web UI | Instant CLI responses |
| Copy-paste release notes manually | `gpc listings sync --dir ./metadata/` |
| Check review replies one by one | `gpc reviews list --min-rating 1 \| jq` |
| Complex CI/CD with multiple tools | Single binary, environment variables |

## 30-Second Demo

```bash
# Install
brew tap AndroidPoet/tap && brew install playconsole-cli

# Authenticate (one-time)
gpc auth login --credentials ~/.config/gpc/key.json

# Deploy to internal testing
gpc bundles upload --file app.aab --track internal

# Promote to production with 10% rollout
gpc tracks promote --from internal --to production --rollout 10

# Done. Go grab coffee.
```

## Features at a Glance

| Feature | Command |
|---------|---------|
| **Upload AAB/APK** | `gpc bundles upload` |
| **Staged Rollouts** | `gpc tracks update --rollout 25` |
| **Track Promotion** | `gpc tracks promote --from beta --to production` |
| **Store Listings** | `gpc listings sync --dir ./metadata/` |
| **Screenshots** | `gpc images sync --dir ./screenshots/` |
| **Reviews** | `gpc reviews list --min-rating 1` |
| **Reply to Reviews** | `gpc reviews reply --text "Thanks!"` |
| **In-App Products** | `gpc products create --sku premium` |
| **Subscriptions** | `gpc subscriptions list` |
| **Verify Purchases** | `gpc purchases verify --token xyz` |
| **Manage Testers** | `gpc testing testers add --email dev@co.com` |

## Table of Contents

- [Installation](#installation)
- [Setup (5 Minutes)](#setup-5-minutes)
- [Commands](#commands)
- [CI/CD Integration](#cicd-integration)
- [Environment Variables](#environment-variables)
- [Design Philosophy](#design-philosophy)
- [Contributing](#contributing)

## Installation

```bash
# Homebrew (recommended)
brew tap AndroidPoet/tap
brew install playconsole-cli

# Or install script
curl -fsSL https://raw.githubusercontent.com/AndroidPoet/playconsole-cli/main/install.sh | bash

# Or build from source
git clone https://github.com/AndroidPoet/playconsole-cli.git
cd playconsole-cli && make build
```

## Setup (5 Minutes)

### Step 1: Create Service Account

1. **Google Cloud Console** → [Create Service Account](https://console.cloud.google.com/iam-admin/serviceaccounts)
2. Name it `play-console-cli`
3. **Keys** tab → **Add Key** → **JSON** → Download

```bash
# Save it securely
mkdir -p ~/.config/gpc
mv ~/Downloads/your-key.json ~/.config/gpc/service-account.json
chmod 600 ~/.config/gpc/service-account.json
```

### Step 2: Enable the API

1. [Enable Google Play Android Developer API](https://console.cloud.google.com/apis/library/androidpublisher.googleapis.com)
2. Click **Enable**

### Step 3: Grant Play Console Access

1. **Play Console** → [Settings → API Access](https://play.google.com/console/developers/api-access)
2. **Link** your Google Cloud project
3. Find your service account → **Grant access**
4. Choose permissions:
   - **Admin** = full access
   - **Release manager** = uploads & releases only
5. Select which apps it can access
6. **Send invite**

> **Note**: Permissions take ~5 minutes to propagate.

### Step 4: Configure CLI

```bash
# Register credentials
gpc auth login --name default --credentials ~/.config/gpc/service-account.json

# Set default package (optional)
gpc auth login --name default \
  --credentials ~/.config/gpc/service-account.json \
  --default-package com.yourcompany.app

# Verify it works
gpc tracks list --package com.yourcompany.app
```

## Commands

### Release Management

```bash
# Upload app bundle
gpc bundles upload --file app.aab --track internal

# List tracks
gpc tracks list

# Promote between tracks
gpc tracks promote --from internal --to beta
gpc tracks promote --from beta --to production --rollout 10

# Update rollout percentage
gpc tracks update --track production --rollout 50

# Complete rollout (100%)
gpc tracks complete --track production

# Halt a bad release
gpc tracks halt --track production
```

### Store Presence

```bash
# Sync listings from fastlane-style directory
gpc listings sync --dir ./metadata/
# metadata/en-US/title.txt, short_description.txt, full_description.txt

# Update single listing
gpc listings update --locale en-US --title "My App" --short-description "Best app ever"

# Sync screenshots
gpc images sync --dir ./screenshots/
# screenshots/en-US/phoneScreenshots/1.png, 2.png...

# Upload single image
gpc images upload --locale en-US --type featureGraphic --file feature.png
```

### Reviews

```bash
# List negative reviews
gpc reviews list --min-rating 1 --max-rating 2

# Get translated reviews
gpc reviews list --translation-lang en

# Reply to a review
gpc reviews reply --review-id "gp:AOqp..." --text "Thanks for the feedback!"

# Pipe to jq for analysis
gpc reviews list | jq '[.[] | select(.rating <= 2)] | length'
```

### Monetization

```bash
# In-app products
gpc products list
gpc products create --product-id premium --title "Premium" --description "Unlock everything"
gpc products get --product-id premium

# Subscriptions
gpc subscriptions list
gpc subscriptions get --product-id monthly_pro
gpc subscriptions base-plans list --product-id monthly_pro

# Verify purchases (server-side validation)
gpc purchases verify --token "purchase-token" --product-id premium
gpc purchases subscription-status --token "sub-token" --product-id monthly_pro
```

### Testing

```bash
# Internal app sharing (instant install link)
gpc testing internal-sharing upload --file app.aab

# Manage testers
gpc testing testers list --track internal
gpc testing testers add --track beta --email "tester@company.com"
gpc testing testers remove --track beta --email "tester@company.com"

# Tester groups
gpc testing tester-groups list
```

### User Access

```bash
# List users
gpc users list

# Grant access
gpc users grant --email "dev@company.com" --role releaseManager

# Revoke access
gpc users revoke --email "contractor@example.com" --confirm
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Deploy to Play Store

on:
  push:
    tags: ['v*']

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build
        run: ./gradlew bundleRelease

      - name: Install GPC
        run: |
          curl -fsSL https://raw.githubusercontent.com/AndroidPoet/playconsole-cli/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: Deploy
        env:
          GPC_CREDENTIALS_B64: ${{ secrets.PLAY_CREDENTIALS }}
          GPC_PACKAGE: com.yourcompany.app
        run: |
          gpc bundles upload --file app.aab --track internal
          gpc tracks promote --from internal --to production --rollout 10
```

### GitLab CI

```yaml
deploy:
  script:
    - curl -fsSL https://raw.githubusercontent.com/AndroidPoet/playconsole-cli/main/install.sh | bash
    - export PATH="$HOME/.local/bin:$PATH"
    - gpc bundles upload --file app.aab --track production --rollout 20
  variables:
    GPC_CREDENTIALS_B64: $PLAY_CREDENTIALS
    GPC_PACKAGE: com.yourcompany.app
```

### Store Credentials Securely

```bash
# Encode your service account for CI secrets
base64 < service-account.json | pbcopy  # macOS
base64 < service-account.json | xclip   # Linux

# Then add as GPC_CREDENTIALS_B64 secret in your CI
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GPC_CREDENTIALS_PATH` | Path to service account JSON |
| `GPC_CREDENTIALS_B64` | Base64-encoded credentials (for CI) |
| `GPC_PACKAGE` | Default package name |
| `GPC_PROFILE` | Auth profile to use |
| `GPC_OUTPUT` | Output format: `json`, `table`, `tsv` |
| `GPC_DEBUG` | Enable debug logging |

## Output Formats

```bash
gpc tracks list                    # JSON (default, for scripting)
gpc tracks list --pretty           # Pretty JSON (for humans)
gpc tracks list --output table     # ASCII table
gpc tracks list --output tsv       # Tab-separated (for spreadsheets)
```

## Design Philosophy

**1. Explicit over clever**
```bash
# Clear intent, no magic
gpc tracks promote --from internal --to beta
```

**2. JSON-first**
```bash
# Pipe to jq, grep, or your scripts
gpc reviews list | jq '.[] | select(.rating == 5)'
```

**3. No prompts**
```bash
# Works in CI without interaction
gpc bundles upload --file app.aab --track internal
```

**4. Clean exit codes**
- `0` = success
- `1` = error
- `2` = validation failed

## Security

- Credentials stored in `~/.playconsole-cli/` with `0600` permissions
- Service account keys never logged
- Base64 encoding for CI/CD secrets
- No credentials in command history

## Contributing

PRs welcome! Please open an issue first to discuss major changes.

```bash
make build    # Build
make test     # Test
make lint     # Lint
```

## License

MIT

---

<p align="center">
  <sub>Not affiliated with Google. Google Play is a trademark of Google LLC.</sub>
</p>

<p align="center">
  <a href="https://github.com/AndroidPoet/playconsole-cli/stargazers">Star this repo</a> if it saved you time!
</p>
