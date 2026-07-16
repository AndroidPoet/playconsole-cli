---
name: gpc-vitals
description: Monitor Android app health with crash rates, ANR rates, startup times, rendering performance, battery metrics, and error reports using the Play Console CLI (gpc).
---

# GPC App Vitals & Analytics

Use this skill when the user wants to check app health, crash rates, ANR rates, performance metrics, device statistics, or error reports from Google Play.

## Prerequisites

- `gpc` CLI installed and authenticated (`gpc auth login`)
- Package name configured via `--package`, `.gpc.yaml`, or `GPC_PACKAGE` env var

## Vitals Commands

### Health Overview

```bash
gpc vitals overview
```

Combined summary of crash rate and ANR rate.

### Crash Rate

```bash
gpc vitals crashes --days 7
gpc vitals crashes --days 28
```

### ANR Rate

```bash
gpc vitals anr --days 7
gpc vitals anr --days 28
```

### Startup Performance

```bash
gpc vitals slow-start --days 28
```

### Rendering Performance

```bash
gpc vitals slow-rendering --days 28
```

### Battery Metrics

```bash
gpc vitals wakeups --days 28       # Excessive wakeup alarms
gpc vitals wakelocks --days 28     # Stuck partial wakelocks
```

### Memory

```bash
gpc vitals memory --days 28        # Low memory killer rate
```

### Error Reports

```bash
gpc vitals errors issues           # Grouped error issues with root causes
```

## Device Commands

### List Supported Devices

```bash
gpc devices list
```

### Device Usage Distribution

```bash
gpc devices stats
```

### Device Tier Configurations

```bash
gpc device-tiers list
gpc device-tiers get --config-id <id>
gpc device-tiers create --file tier-config.json
```

## Reports

```bash
gpc reports list                   # Available reports
gpc reports types                  # Report type metadata
```

## Common Patterns

### Quick Health Check

```bash
gpc vitals overview --pretty
```

### Monitor After Release

```bash
gpc vitals crashes --days 7 --pretty
gpc vitals anr --days 7 --pretty
```

### Export Metrics

```bash
gpc vitals crashes --days 28 -o csv > crashes.csv
gpc devices stats -o csv > device_distribution.csv
```

## Global Flags

All commands support: `--package/-p`, `--output/-o` (json/table/tsv/csv/yaml/minimal), `--pretty`, `--quiet/-q`, `--debug`, `--dry-run`, `--timeout`, `--profile`.
