---
name: gpc-reviews
description: List, filter, and reply to Google Play user reviews using the Play Console CLI (gpc).
---

# GPC Reviews Management

Use this skill when the user wants to read app reviews, filter by rating, or reply to user reviews on Google Play.

## Prerequisites

- `gpc` CLI installed and authenticated (`gpc auth login`)
- Package name configured via `--package`, `.gpc.yaml`, or `GPC_PACKAGE` env var

## Commands

### List Reviews

```bash
gpc reviews list
```

Options:
- `--min-rating <n>`: Filter reviews with rating >= n (1-5)
- `--max-rating <n>`: Filter reviews with rating <= n (1-5)
- `--translation-lang <code>`: Get translated review text (e.g., `en`)
- `--max-results <n>`: Limit number of results

### Get a Specific Review

```bash
gpc reviews get --review-id "gp:AOqpT..."
```

### Reply to a Review

```bash
gpc reviews reply --review-id "gp:AOqpT..." --text "Thank you for the feedback!"
```

## Common Patterns

### Find Negative Reviews

```bash
gpc reviews list --min-rating 1 --max-rating 2
```

### Find 5-Star Reviews

```bash
gpc reviews list --min-rating 5
```

### Filter Reviews with jq

```bash
gpc reviews list | jq '[.[] | select(.rating == 1)]'
```

### Reply to All 1-Star Reviews (Scripting)

```bash
gpc reviews list --min-rating 1 --max-rating 1 -o json | \
  jq -r '.[].reviewId' | \
  while read id; do
    gpc reviews reply --review-id "$id" --text "We're sorry about your experience. Please contact support@example.com"
  done
```

## Global Flags

All commands support: `--package/-p`, `--output/-o` (json/table/tsv/csv/yaml/minimal), `--pretty`, `--quiet/-q`, `--debug`, `--dry-run`, `--timeout`, `--profile`.
