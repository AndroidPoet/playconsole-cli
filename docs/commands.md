# Command Reference

Complete reference for all `gpc` commands.

---

## Table of Contents

- [Release Management](#release-management)
  - [bundles](#bundles) · [apks](#apks) · [tracks](#tracks) · [deobfuscation](#deobfuscation)
- [Store Presence](#store-presence)
  - [listings](#listings) · [images](#images)
- [Reviews](#reviews)
- [Monetization](#monetization)
  - [products](#products) · [subscriptions](#subscriptions) · [offers](#offers) · [purchases](#purchases) · [orders](#orders) · [external-transactions](#external-transactions)
- [Quality & Vitals](#quality--vitals)
  - [vitals](#vitals) · [devices](#devices) · [reports](#reports)
- [Testing](#testing)
- [Team & Access](#team--access)
  - [users](#users)
- [App Configuration](#app-configuration)
  - [edits](#edits) · [availability](#availability) · [device-tiers](#device-tiers) · [recovery](#recovery) · [diff](#diff)
- [Setup & Utilities](#setup--utilities)
  - [auth](#auth) · [setup](#setup) · [doctor](#doctor) · [init](#init) · [completion](#completion) · [version](#version) · [stats](#stats)

---

## Release Management

### bundles

Upload and manage Android App Bundles.

```bash
gpc bundles upload --file app.aab --track internal    # Upload AAB to track
gpc bundles list                                       # List uploaded bundles
```

### apks

Manage APKs (legacy — prefer bundles).

```bash
gpc apks upload --file app.apk                         # Upload APK (deprecated)
gpc apks list                                          # List APKs
```

### tracks

Manage release tracks (internal, alpha, beta, production).

```bash
gpc tracks list                                        # List all tracks
gpc tracks get --track production                      # Get track details
gpc tracks update --track production --rollout 50      # Staged rollout
gpc tracks promote --from internal --to beta           # Promote release
gpc tracks halt --track production                     # Halt rollout
gpc tracks complete --track production                 # Complete to 100%
```

### deobfuscation

Upload ProGuard/R8 mapping files or native debug symbols for crash symbolication.

```bash
gpc deobfuscation upload --version-code 42 --file mapping.txt --type proguard
gpc deobfuscation upload --version-code 42 --file symbols.zip --type native-code
```

**Types:**
| Type | Description |
|------|-------------|
| `proguard` | ProGuard/R8 `mapping.txt` file |
| `native-code` | Native debug symbols ZIP (`symbols.zip`) |

---

## Store Presence

### listings

Manage localized store listings.

```bash
gpc listings list                                      # List all locales
gpc listings get --locale en-US                        # Get specific locale
gpc listings update --locale en-US --title "My App"    # Update listing
gpc listings sync --dir ./metadata/                    # Sync from directory
```

### images

Manage screenshots and promotional graphics.

```bash
gpc images list --locale en-US --type phoneScreenshots
gpc images upload --locale en-US --type phoneScreenshots --file screenshot.png
gpc images delete --locale en-US --type phoneScreenshots --sha1 abc123
gpc images delete-all --locale en-US --type phoneScreenshots
gpc images sync --dir ./screenshots/
```

---

## Reviews

Manage app reviews and respond to user feedback.

```bash
gpc reviews list                                       # All reviews
gpc reviews list --min-rating 1 --max-rating 2         # Negative reviews
gpc reviews get --review-id "gp:AOqpT..."              # Single review
gpc reviews reply --review-id "gp:..." --text "Thanks!"
```

---

## Monetization

### products

Manage in-app products (one-time purchases).

```bash
gpc products list                                      # List products
gpc products get --product-id premium_unlock            # Get details
gpc products create --product-id coins_100 --file product.json
gpc products update --product-id coins_100 --file product.json
gpc products delete --product-id coins_100 --confirm
```

### subscriptions

Manage subscription products and base plans.

```bash
gpc subscriptions list                                 # List subscriptions
gpc subscriptions get --product-id monthly_pro         # Get details
gpc subscriptions create --product-id annual_pro --file sub.json
gpc subscriptions base-plans list --product-id monthly_pro
gpc subscriptions base-plans create --product-id monthly_pro --file plan.json
gpc subscriptions pricing get --product-id monthly_pro --base-plan monthly
```

### offers

Manage subscription offers (introductory pricing, free trials, promotions).

```bash
gpc offers list --product-id monthly_pro --base-plan monthly
gpc offers get --product-id monthly_pro --base-plan monthly --offer-id free_trial
gpc offers create --product-id monthly_pro --base-plan monthly --file offer.json
gpc offers update --product-id monthly_pro --base-plan monthly --offer-id free_trial --file offer.json
gpc offers delete --product-id monthly_pro --base-plan monthly --offer-id free_trial --confirm
gpc offers activate --product-id monthly_pro --base-plan monthly --offer-id free_trial
gpc offers deactivate --product-id monthly_pro --base-plan monthly --offer-id free_trial
```

**Example `offer.json`:**
```json
{
  "offerId": "free_trial_7d",
  "phases": [
    {
      "duration": "P7D",
      "recurrenceCount": 1,
      "otherRegionsConfig": {
        "otherRegionsNewSubscriberAvailability": true
      }
    }
  ],
  "offerTags": [{"tag": "trial"}]
}
```

### purchases

Verify and manage purchases.

```bash
gpc purchases verify --product-id premium --token "purchase_token..."
gpc purchases subscription-status --product-id monthly --token "token..."
gpc purchases acknowledge --product-id premium --token "token..."
gpc purchases voided list
```

### orders

View order details and issue refunds.

```bash
gpc orders get --order-id GPA.1234-5678-9012
gpc orders refund --order-id GPA.1234-5678-9012 --confirm
gpc orders batch-get --order-ids GPA.1234,GPA.5678
```

### external-transactions

Manage transactions processed outside Google Play Billing (alternative billing compliance).

```bash
gpc external-transactions create --file tx.json
gpc external-transactions get --name "apps/com.example/externalTransactions/TX_ID"
gpc external-transactions refund --name "apps/com.example/externalTransactions/TX_ID" --confirm
```

Alias: `gpc ext-tx`

---

## Quality & Vitals

### vitals

Access Android Vitals data from the Play Developer Reporting API.

```bash
# Core metrics
gpc vitals overview                     # Health summary (crash + ANR rates)
gpc vitals crashes --days 7             # Crash rate metrics
gpc vitals anr --days 28               # ANR rate metrics

# Performance metrics
gpc vitals slow-start --days 28         # Slow app startup rate
gpc vitals slow-rendering --days 28     # Slow frame rendering rate

# Battery metrics
gpc vitals wakeups --days 28            # Excessive wakeup alarm rate
gpc vitals wakelocks --days 28          # Stuck background wakelock rate

# Memory metrics
gpc vitals memory --days 28             # Low memory killer (LMK) rate

# Error tracking
gpc vitals errors --days 28             # Aggregated error counts
gpc vitals errors issues                # Grouped error issues with causes
```

### devices

View device catalog and compatibility.

```bash
gpc devices list                        # List supported form factors
gpc devices stats                       # Device usage statistics
```

### reports

View available report types.

```bash
gpc reports list                        # List available reports
gpc reports types                       # Show all report types with details
```

---

## Testing

Manage testing tracks and testers.

```bash
gpc testing internal list               # List internal test builds
gpc testing internal-sharing upload --file app.aab   # Get instant test link
gpc testing testers list --track beta   # List testers
gpc testing testers add --track beta --email "dev@company.com"
gpc testing testers add --track beta --file testers.txt
gpc testing testers remove --track beta --email "dev@company.com"
gpc testing tester-groups list          # List tester groups
```

---

## Team & Access

### users

Manage user access and permissions.

```bash
gpc users list                          # List team members
gpc users grant --email "dev@co.com" --role releaseManager
gpc users revoke --email "dev@co.com"
```

**Roles:** `admin`, `releaseManager`, `appOwner`

---

## App Configuration

### edits

Manage edit sessions (advanced — most commands handle edits internally).

```bash
gpc edits create                        # Start new edit session
gpc edits get --edit-id EDIT_ID         # Get existing edit
gpc edits validate --edit-id EDIT_ID    # Validate changes
gpc edits commit --edit-id EDIT_ID      # Commit edit (go live)
gpc edits delete --edit-id EDIT_ID      # Discard edit
```

### availability

Manage country targeting per release track.

```bash
gpc availability list --track production
gpc availability update --track production --countries US,GB,DE,FR --confirm
gpc availability update --track production --countries US --include-rest=false --confirm
```

### device-tiers

Manage device tier configurations for targeted content delivery.

```bash
gpc device-tiers list                   # List device tier configs
gpc device-tiers get --config-id 123    # Get config details
gpc device-tiers create --file config.json
```

Alias: `gpc dt`

### recovery

Manage app recovery actions for production incidents.

```bash
gpc recovery list                       # List recovery actions
gpc recovery create --file recovery-action.json
gpc recovery deploy --recovery-id 123 --confirm
gpc recovery cancel --recovery-id 123 --confirm
gpc recovery add-targeting --recovery-id 123 --file targeting.json
```

### diff

Compare draft edit state against live version.

```bash
gpc diff                                # Full diff (listings + tracks)
gpc diff --section listings             # Listings only
gpc diff --section tracks               # Tracks only
gpc diff --edit-id EDIT_ID              # Diff specific edit
```

---

## Setup & Utilities

### auth

Manage authentication profiles.

```bash
gpc auth login --credentials path/to/service-account.json
gpc auth login --credentials-b64 "base64_string"
gpc auth list                           # List profiles
gpc auth current                        # Show active profile
gpc auth switch --profile production    # Switch profile
gpc auth delete --profile old-profile
```

### setup

Interactive setup wizard.

```bash
gpc setup                               # Guided first-time setup
```

### doctor

Validate CLI setup and credentials.

```bash
gpc doctor                              # Run all diagnostic checks
gpc doctor --verbose                    # Detailed output
```

**Checks performed:**
1. Configuration file valid
2. Credentials available
3. Service account JSON valid
4. Package name configured
5. Android Publisher API reachable
6. Reporting API reachable

### init

Initialize project configuration.

```bash
gpc init                                # Create .gpc.yaml with defaults
gpc init --package com.example.app      # Set package name
gpc init --track production --output table
gpc init --force                        # Overwrite existing
```

Creates `.gpc.yaml` in the current directory. The CLI auto-detects this file in the current or parent directories.

### completion

Generate shell completion scripts.

```bash
# Bash
source <(gpc completion bash)
gpc completion bash > /etc/bash_completion.d/gpc              # Linux
gpc completion bash > $(brew --prefix)/etc/bash_completion.d/gpc  # macOS

# Zsh
source <(gpc completion zsh)
gpc completion zsh > "${fpath[1]}/_gpc"

# Fish
gpc completion fish | source
gpc completion fish > ~/.config/fish/completions/gpc.fish

# PowerShell
gpc completion powershell | Out-String | Invoke-Expression
```

### version

Print version information.

```bash
gpc version
```

### stats

View CLI download statistics.

```bash
gpc stats downloads                     # Download counts by release
gpc stats sources                       # Download sources
```

---

## Global Flags

Available on every command:

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--package` | `-p` | App package name | `GPC_PACKAGE` env |
| `--output` | `-o` | Output format | `json` |
| `--pretty` | | Pretty-print JSON | `false` |
| `--quiet` | `-q` | Suppress non-essential output | `false` |
| `--debug` | | Show API requests/responses | `false` |
| `--dry-run` | | Preview without applying | `false` |
| `--timeout` | | Request timeout | `60s` |
| `--config` | | Config file path | `~/.playconsole-cli/config.json` |
| `--profile` | | Auth profile name | `GPC_PROFILE` env |

## Output Formats

```bash
gpc tracks list                         # JSON (default)
gpc tracks list --pretty                # Pretty-printed JSON
gpc tracks list -o table                # ASCII table
gpc tracks list -o tsv                  # Tab-separated values
gpc tracks list -o csv                  # Comma-separated values
gpc tracks list -o yaml                 # YAML
gpc tracks list -o minimal             # First field only (for scripting)
```
