<div align="center">

<img src="assets/logo.png" alt="Play Console CLI" width="200">

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

### 📤 Release Management

```bash
gpc bundles upload --file app.aab --track internal    # Upload
gpc tracks list                                        # List tracks
gpc tracks promote --from internal --to beta           # Promote
gpc tracks update --track production --rollout 50     # Staged rollout
gpc tracks halt --track production                    # Emergency halt
```

### 🏪 Store Presence

```bash
gpc listings sync --dir ./metadata/                   # Sync all listings
gpc listings update --locale en-US --title "My App"   # Update listing
gpc images sync --dir ./screenshots/                  # Sync screenshots
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
gpc purchases verify --token "..." --product-id premium
```

### 📊 Analytics & Vitals

```bash
gpc vitals overview                                    # Health summary
gpc vitals crashes --days 7                            # Crash metrics
gpc vitals anr --days 30                               # ANR metrics
```

### 📱 Devices

```bash
gpc devices list                                       # Supported devices
gpc devices stats                                      # Device distribution
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
| `GPC_OUTPUT` | Format: `json` \| `table` \| `tsv` |

---

## 🎨 Output Formats

```bash
gpc tracks list                    # JSON (default, for scripting)
gpc tracks list --pretty           # Pretty JSON
gpc tracks list --output table     # ASCII table
gpc tracks list --output tsv       # Tab-separated
```

---

## 🧠 Design Philosophy

1. **Explicit over clever** — No magic, clear intent
2. **JSON-first** — Pipe to jq, grep, or your scripts
3. **No prompts** — Works in CI without interaction
4. **Clean exit codes** — 0 success, 1 error, 2 validation

---

## 🔒 Security

- Credentials stored with `0600` permissions
- Service account keys never logged
- Base64 encoding for CI/CD secrets
- No credentials in command history

> ⚠️ **SCAM WARNING:** There is **NO official cryptocurrency, token, or NFT** associated with Play Console CLI. Any tokens using this project's name are unauthorized scams. This is a free, open-source CLI tool — nothing more.

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
