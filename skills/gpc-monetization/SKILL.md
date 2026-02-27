---
name: gpc-monetization
description: Manage in-app products, subscriptions, offers, purchases, orders, and external transactions using the Play Console CLI (gpc).
---

# GPC Monetization

Use this skill when the user wants to manage in-app purchases, subscriptions, subscription offers, verify purchases, handle orders/refunds, or manage external transactions for alternative billing.

## Prerequisites

- `gpc` CLI installed and authenticated (`gpc auth login`)
- Package name configured via `--package`, `.gpc.yaml`, or `GPC_PACKAGE` env var

## In-App Products (One-Time Purchases)

```bash
gpc products list
gpc products get --product-id coins_100
gpc products create --product-id coins_100 --file product.json
gpc products update --product-id coins_100 --file product.json
gpc products delete --product-id coins_100
```

## Subscriptions

```bash
gpc subscriptions list
gpc subscriptions get --product-id monthly_pro
gpc subscriptions create --product-id monthly_pro --file subscription.json

# Base plans
gpc subscriptions base-plans list --product-id monthly_pro

# Pricing
gpc subscriptions pricing --product-id monthly_pro --base-plan monthly
```

## Subscription Offers

```bash
gpc offers list --product-id monthly_pro --base-plan monthly
gpc offers get --product-id monthly_pro --base-plan monthly --offer-id free_trial
gpc offers create --product-id monthly_pro --base-plan monthly --file offer.json
gpc offers update --product-id monthly_pro --base-plan monthly --offer-id free_trial --file offer.json
gpc offers delete --product-id monthly_pro --base-plan monthly --offer-id free_trial
gpc offers activate --product-id monthly_pro --base-plan monthly --offer-id free_trial
gpc offers deactivate --product-id monthly_pro --base-plan monthly --offer-id free_trial
```

## Purchase Verification

```bash
# Verify a one-time purchase
gpc purchases verify --product-id premium --token "purchase_token_here"

# Check subscription status
gpc purchases subscription-status --product-id monthly_pro --token "sub_token_here"

# Acknowledge a purchase
gpc purchases acknowledge --product-id premium --token "purchase_token_here"

# List voided purchases
gpc purchases voided
```

## Orders

```bash
gpc orders get --order-id GPA.1234-5678
gpc orders batch-get --order-ids GPA.1234,GPA.5678
gpc orders refund --order-id GPA.1234-5678 --confirm
```

The `--confirm` flag is required for refunds as a safety measure.

## External Transactions (Alternative Billing)

Alias: `ext-tx`

```bash
gpc external-transactions create --file tx.json
gpc external-transactions get --external-transaction-id "ext_123"
gpc external-transactions refund --external-transaction-id "ext_123" --file refund.json
```

## Global Flags

All commands support: `--package/-p`, `--output/-o` (json/table/tsv/csv/yaml/minimal), `--pretty`, `--quiet/-q`, `--debug`, `--dry-run`, `--timeout`, `--profile`.
