# playconsole-cli

A fast, lightweight, and scriptable CLI for Google Play Console.

Inspired by [App Store Connect CLI](https://github.com/rudrankriyam/App-Store-Connect-CLI).

## Installation

### Homebrew

```bash
brew tap AndroidPoet/tap
brew install playconsole-cli
```

### Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/AndroidPoet/playconsole-cli/main/install.sh | bash
```

Installs to `~/.local/bin` by default (ensure it's in your PATH).

### From Source

```bash
git clone https://github.com/AndroidPoet/playconsole-cli.git
cd playconsole-cli
make build
./bin/playconsole-cli --help
```

## Setup

### 1. Create a Service Account

1. Go to [Google Cloud Console](https://console.cloud.google.com/iam-admin/serviceaccounts)
2. Create a service account and download the JSON key
3. Enable the **Google Play Developer API**

### 2. Grant Access in Play Console

1. Open [Play Console → API Access](https://play.google.com/console/developers/api-access)
2. Link your Google Cloud project
3. Grant your service account access to apps

### 3. Configure playconsole-cli

```bash
playconsole-cli auth login --name default --credentials /path/to/service-account.json
```

Or use environment variables:

```bash
export GPC_CREDENTIALS_PATH=/path/to/service-account.json
export GPC_PACKAGE=com.example.app
```

## Commands

| Command | Description |
|---------|-------------|
| `auth` | Manage authentication profiles |
| `apps` | Application details |
| `tracks` | Release tracks (internal/alpha/beta/production) |
| `bundles` | Upload and manage App Bundles |
| `apks` | Upload APKs (legacy) |
| `listings` | Store metadata and descriptions |
| `images` | Screenshots and graphics |
| `reviews` | View and reply to reviews |
| `products` | In-app products |
| `subscriptions` | Subscription management |
| `purchases` | Purchase verification |
| `testing` | Internal sharing and testers |
| `users` | Access control |
| `edits` | Low-level edit sessions |

## Usage Examples

### Upload a Release

```bash
# Upload bundle to internal track
playconsole-cli bundles upload --file app.aab --track internal

# Promote to beta
playconsole-cli tracks promote --from internal --to beta

# Staged rollout to production (10%)
playconsole-cli tracks update --track production --version-code 42 --rollout-percentage 10

# Complete rollout
playconsole-cli tracks complete --track production
```

### Manage Store Listing

```bash
# List localizations
playconsole-cli listings list

# Update listing
playconsole-cli listings update --locale en-US \
  --title "My App" \
  --short-description "A great app"

# Sync from directory (fastlane-compatible)
playconsole-cli listings sync --dir ./metadata/
```

### Screenshots

```bash
# Upload screenshot
playconsole-cli images upload --locale en-US --type phoneScreenshots --file screenshot.png

# Sync all images
playconsole-cli images sync --dir ./screenshots/
```

### Reviews

```bash
# List negative reviews
playconsole-cli reviews list --min-rating 1 --max-rating 3

# Reply to a review
playconsole-cli reviews reply --review-id "abc123" --text "Thank you for your feedback!"
```

### In-App Products

```bash
# List products
playconsole-cli products list

# Create product
playconsole-cli products create --sku premium --title "Premium" --price-usd 4.99
```

## Output Formats

```bash
playconsole-cli tracks list                    # JSON (default)
playconsole-cli tracks list --pretty           # Pretty JSON
playconsole-cli tracks list --output table     # Table format
playconsole-cli tracks list --output tsv       # TSV for scripting
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GPC_CREDENTIALS_PATH` | Path to service account JSON |
| `GPC_CREDENTIALS_B64` | Base64-encoded credentials |
| `GPC_PACKAGE` | Default package name |
| `GPC_PROFILE` | Default auth profile |
| `GPC_OUTPUT` | Default output format |
| `GPC_DEBUG` | Enable debug logging |

## CI/CD

### GitHub Actions

```yaml
- name: Deploy to Play Store
  env:
    GPC_CREDENTIALS_B64: ${{ secrets.PLAY_STORE_CREDENTIALS }}
  run: |
    playconsole-cli bundles upload \
      --package com.example.app \
      --file app.aab \
      --track internal
```

## Design Philosophy

- **JSON-first** — Machine-readable output by default
- **Explicit flags** — No magic, no cryptic shortcuts
- **No prompts** — Fully scriptable, CI/CD-ready
- **Clean exit codes** — 0 success, 1 error, 2 validation

## License

MIT
