<div align="center">

<img src="assets/logo.png" alt="Play Console CLI" width="150">

# Play Console CLI

### Ship Android apps from your terminal

[![Latest Release](https://img.shields.io/github/v/release/AndroidPoet/playconsole-cli?style=for-the-badge&color=3DDC84&logo=android)](https://github.com/AndroidPoet/playconsole-cli/releases/latest)
[![Downloads](https://img.shields.io/github/downloads/AndroidPoet/playconsole-cli/total?style=for-the-badge&color=blue)](https://github.com/AndroidPoet/playconsole-cli/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/AndroidPoet/playconsole-cli/ci.yml?style=for-the-badge&label=CI)](https://github.com/AndroidPoet/playconsole-cli/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge)](LICENSE)

**No browser. No clicking. Just ship.**

```bash
gpc bundles upload --file app.aab --track production
```

[Install](#-installation) · [Quick Start](#-quick-start) · [Commands](#-commands) · [CI/CD](#-cicd-integration)

</div>

---

## ✨ Why Developers Love GPC

| 😤 The Old Way | 🚀 The GPC Way |
|---------------|----------------|
| Open browser, navigate menus, wait... | `gpc bundles upload --track internal` |
| Copy-paste release notes manually | `gpc listings sync --dir ./metadata/` |
| Check reviews one by one | `gpc reviews list --min-rating 1 \| jq` |
| Complex CI/CD setup | Single binary + env vars |
| "Is it uploaded yet?" | **Instant feedback** |

> *Inspired by [App Store Connect CLI](https://github.com/rudrankriyam/App-Store-Connect-CLI) — the same philosophy, now for Android.*

---

## 📦 Installation

```bash
# Homebrew (recommended)
brew tap AndroidPoet/tap && brew install playconsole-cli

# Install script (Linux/macOS)
curl -fsSL https://raw.githubusercontent.com/AndroidPoet/playconsole-cli/main/install.sh | bash

# Build from source
git clone https://github.com/AndroidPoet/playconsole-cli.git
cd playconsole-cli && make build
```

After install, you can use either `playconsole-cli` or the shorter alias `gpc`.

---

## ⚡ Quick Start

**1. Create a service account** ([Google Cloud Console](https://console.cloud.google.com/iam-admin/serviceaccounts))
```bash
mkdir -p ~/.config/gpc
mv ~/Downloads/your-key.json ~/.config/gpc/service-account.json
chmod 600 ~/.config/gpc/service-account.json
```

**2. Enable the API** → [Enable Google Play Android Developer API](https://console.cloud.google.com/apis/library/androidpublisher.googleapis.com)

**3. Grant access** in [Play Console API Settings](https://play.google.com/console/developers/api-access)

**4. Configure & verify**
```bash
gpc auth login --credentials ~/.config/gpc/service-account.json
gpc apps list  # See your apps
```

**5. Deploy!** 🎉
```bash
gpc bundles upload --file app.aab --track internal
gpc tracks promote --from internal --to production --rollout 10
```

---

## 🎯 Commands

**30 command groups, 80+ subcommands.** [Full reference →](docs/commands.md)

### 📤 Release Management

```bash
gpc bundles upload --file app.aab --track internal    # Upload
gpc bundles find --version-code 42                     # Find by version code
gpc bundles wait --version-code 42                     # Wait for processing
gpc tracks list                                        # List tracks
gpc tracks promote --from internal --to beta           # Promote
gpc tracks update --track production --rollout 50     # Staged rollout
gpc tracks halt --track production                    # Emergency halt
gpc deobfuscation upload --version-code 42 --file mapping.txt  # Crash symbolication
```

### 🏪 Store Presence

```bash
gpc listings sync --dir ./metadata/                   # Sync all listings
gpc listings update --locale en-US --title "My App"   # Update listing
gpc images sync --dir ./screenshots/                  # Sync screenshots
gpc availability list --track production              # Country targeting
```

### ⭐ Reviews

```bash
gpc reviews list --min-rating 1 --max-rating 2        # Negative reviews
gpc reviews reply --review-id "gp:..." --text "Thanks!"
gpc reviews list | jq '[.[] | select(.rating == 5)]' # Filter with jq
```

### 💰 Monetization

```bash
gpc products list                                      # In-app products
gpc subscriptions list                                 # Subscriptions
gpc offers list --product-id sub_id --base-plan monthly # Subscription offers
gpc purchases verify --token "..." --product-id premium
gpc orders get --order-id GPA.1234                     # Order details
gpc orders refund --order-id GPA.1234 --confirm        # Issue refund
gpc external-transactions create --file tx.json        # Alternative billing
```

### 📊 Analytics & Vitals

```bash
gpc vitals overview                                    # Health summary
gpc vitals crashes --days 7                            # Crash rate
gpc vitals anr --days 28                               # ANR rate
gpc vitals slow-start --days 28                        # Slow startup rate
gpc vitals slow-rendering --days 28                    # Frame rendering
gpc vitals wakeups --days 28                           # Battery: wakeup alarms
gpc vitals wakelocks --days 28                         # Battery: stuck wakelocks
gpc vitals memory --days 28                            # Low memory killer rate
gpc vitals errors issues                               # Grouped error issues
```

### 📱 Devices

```bash
gpc devices list                                       # Supported devices
gpc devices stats                                      # Device distribution
gpc device-tiers list                                  # Device tier configs
```

### 📈 Reports

```bash
gpc reports list                                       # Available reports
gpc reports types                                      # Report type info
```

### 🧪 Testing

```bash
gpc testing internal-sharing upload --file app.aab   # Instant test link
gpc testing testers add --track beta --email "dev@company.com"
```

### 👥 Team

```bash
gpc users list
gpc users grant --email "dev@company.com" --role releaseManager
```

### 🛠️ Utilities

```bash
gpc doctor                                             # Validate setup
gpc init --package com.example.app                     # Create project config
gpc diff                                               # Compare draft vs live
gpc recovery list                                      # App recovery actions
gpc completion zsh > "${fpath[1]}/_gpc"                # Shell completions
```

---

## 🔄 CI/CD Integration

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
          gpc bundles upload --file app/build/outputs/bundle/release/app-release.aab --track internal
          gpc tracks promote --from internal --to production --rollout 10
```

### Encode Credentials for CI

```bash
base64 < service-account.json | pbcopy  # macOS
base64 < service-account.json | xclip   # Linux
# Add as GPC_CREDENTIALS_B64 secret
```

---

## ⚙️ Environment Variables

| Variable | Description |
|----------|-------------|
| `GPC_CREDENTIALS_PATH` | Path to service account JSON |
| `GPC_CREDENTIALS_B64` | Base64-encoded credentials (CI) |
| `GPC_PACKAGE` | Default package name |
| `GPC_PROFILE` | Auth profile to use |
| `GPC_OUTPUT` | Format: `json` \| `table` \| `tsv` \| `csv` \| `yaml` |

---

## 🎨 Output Formats

```bash
gpc tracks list                    # JSON (default, for scripting)
gpc tracks list --pretty           # Pretty JSON
gpc tracks list -o table           # ASCII table
gpc tracks list -o tsv             # Tab-separated values
gpc tracks list -o csv             # Comma-separated values
gpc tracks list -o yaml            # YAML
gpc tracks list -o minimal         # First field only (piping)
```

---

## 🔒 Security

- Credentials stored with `0600` permissions
- Service account keys never logged
- Base64 encoding for CI/CD secrets
- No credentials in command history

---

## 🏆 Wall of Apps

**Apps shipped using Play Console CLI:**

<!-- WALL_OF_APPS_START -->
| App | Creator |
|-----|---------|
| [Wally](https://play.google.com/store/apps/details?id=com.androidpoet.wally) | [@AndroidPoet](https://github.com/AndroidPoet) |
<!-- WALL_OF_APPS_END -->

<details>
<summary><b>🚀 Add your app to the Wall!</b></summary>

Using GPC to ship your app? Get featured here!

1. Fork this repo
2. Edit [`docs/wall-of-apps.json`](docs/wall-of-apps.json)
3. Add your app:
```json
{
  "app": "Your App Name",
  "link": "https://play.google.com/store/apps/details?id=your.package",
  "creator": "YourGitHubUsername"
}
```
4. Submit a PR!

</details>

---

## 🤖 Agent Skills

Supercharge your coding agent with Play Console CLI skills. Works with [Claude Code](https://claude.ai/claude-code), [Cursor](https://cursor.sh), [Codex](https://openai.com/codex), and [37+ more agents](https://skills.sh).

### Install All Skills

```bash
npx skills add AndroidPoet/playconsole-cli
```

### Install Specific Skills

```bash
# Release management (upload, tracks, promote, rollouts)
npx skills add AndroidPoet/playconsole-cli --skill gpc-release

# Store presence (listings, images, availability)
npx skills add AndroidPoet/playconsole-cli --skill gpc-store

# Reviews (list, filter, reply)
npx skills add AndroidPoet/playconsole-cli --skill gpc-reviews

# Monetization (products, subscriptions, offers, orders)
npx skills add AndroidPoet/playconsole-cli --skill gpc-monetization

# Vitals & analytics (crashes, ANR, performance, devices)
npx skills add AndroidPoet/playconsole-cli --skill gpc-vitals

# Testing (internal sharing, testers, test tracks)
npx skills add AndroidPoet/playconsole-cli --skill gpc-testing

# Setup & auth (authentication, config, doctor, team)
npx skills add AndroidPoet/playconsole-cli --skill gpc-setup
```

### Install to a Specific Agent

```bash
npx skills add AndroidPoet/playconsole-cli -a claude-code
npx skills add AndroidPoet/playconsole-cli -a cursor
npx skills add AndroidPoet/playconsole-cli -a codex
```

### Install Globally

```bash
npx skills add AndroidPoet/playconsole-cli -g
```

### Available Skills

| Skill | Description |
|-------|-------------|
| `gpc-release` | Upload bundles, manage tracks, promote releases, staged rollouts |
| `gpc-store` | Store listings, screenshots, images, country availability |
| `gpc-reviews` | List, filter, and reply to user reviews |
| `gpc-monetization` | In-app products, subscriptions, offers, orders, refunds |
| `gpc-vitals` | Crash rates, ANR, performance, battery, device stats |
| `gpc-testing` | Internal sharing, testers, tester groups, test tracks |
| `gpc-setup` | Authentication, project config, doctor, team management |

---

## 🤝 Contributing

PRs welcome! Please open an issue first to discuss major changes.

```bash
make build    # Build
make test     # Test
make lint     # Lint
```

---

## 📄 License

MIT

---

<div align="center">

**[⬆ Back to top](#-play-console-cli)**

<sub>Not affiliated with Google. Google Play is a trademark of Google LLC.</sub>

### If GPC saved you time, [give it a ⭐](https://github.com/AndroidPoet/playconsole-cli/stargazers)

</div>
