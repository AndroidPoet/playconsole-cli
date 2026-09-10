---
name: gpc-testing
description: Manage test tracks, testers, tester groups, and internal app sharing using the Play Console CLI (gpc).
---

# GPC Testing

Use this skill when the user wants to manage app testing — upload for internal sharing, add/remove testers, manage tester groups, or work with test tracks.

## Prerequisites

- `gpc` CLI installed and authenticated (`gpc auth login`)
- Package name configured via `--package`, `.gpc.yaml`, or `GPC_PACKAGE` env var

## Internal App Sharing

Upload a bundle and get an instant shareable test link:

```bash
gpc testing internal-sharing upload --file app.aab
```

Returns a download URL that testers can use immediately without waiting for Play Store processing.

## Internal Test Track

```bash
gpc testing internal list
```

## Testers

### List Testers on a Track

```bash
gpc testing testers list --track beta
```

### Add a Tester

```bash
# Single email
gpc testing testers add --track beta --email "dev@company.com"

# Bulk from file (one email per line)
gpc testing testers add --track beta --file testers.txt
```

### Remove a Tester

```bash
gpc testing testers remove --track beta --email "dev@company.com"
```

## Tester Groups

```bash
gpc testing tester-groups list
```

## Common Patterns

### Quick Internal Test

```bash
gpc testing internal-sharing upload --file app.aab
# Share the returned URL with your team
```

### Set Up Beta Testing

```bash
gpc bundles upload --file app.aab --track beta
gpc testing testers add --track beta --file beta-testers.txt
```

### Promote from Internal to Beta

```bash
gpc tracks promote --from internal --to beta
```

## Global Flags

All commands support: `--package/-p`, `--output/-o` (json/table/tsv/csv/yaml/minimal), `--pretty`, `--quiet/-q`, `--debug`, `--dry-run`, `--timeout`, `--profile`.
