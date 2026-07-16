---
name: gpc-store
description: Manage Google Play store listings, screenshots, images, and country availability using the Play Console CLI (gpc).
---

# GPC Store Presence

Use this skill when the user wants to manage app store listings (title, description, release notes), upload or sync screenshots and images, or configure country/region availability.

## Prerequisites

- `gpc` CLI installed and authenticated (`gpc auth login`)
- Package name configured via `--package`, `.gpc.yaml`, or `GPC_PACKAGE` env var

## Listings Commands

### List All Locale Listings

```bash
gpc listings list
```

### Get a Specific Locale Listing

```bash
gpc listings get --locale en-US
```

### Update a Listing

```bash
gpc listings update --locale en-US --title "My App" --short-description "Short desc" --full-description "Full desc"
```

### Sync Listings from Directory

```bash
gpc listings sync --dir ./metadata/
```

Expected directory structure:
```
metadata/
├── en-US/
│   ├── title.txt
│   ├── short_description.txt
│   ├── full_description.txt
│   └── changelogs/
│       └── default.txt
├── es-ES/
│   └── ...
```

## Image Commands

### List Images

```bash
gpc images list --locale en-US --type phoneScreenshots
```

### Upload an Image

```bash
gpc images upload --locale en-US --type phoneScreenshots --file screenshot.png
```

Image types: `phoneScreenshots`, `sevenInchScreenshots`, `tenInchScreenshots`, `tvScreenshots`, `wearScreenshots`, `featureGraphic`, `previewGraphic`, `header`, `icon`.

### Delete an Image

```bash
gpc images delete --locale en-US --type phoneScreenshots --image-id <id>
```

### Delete All Images of a Type

```bash
gpc images delete-all --locale en-US --type phoneScreenshots
```

### Sync Images from Directory

```bash
gpc images sync --dir ./screenshots/
```

Expected directory structure:
```
screenshots/
├── en-US/
│   ├── phoneScreenshots/
│   │   ├── 01.png
│   │   └── 02.png
│   └── featureGraphic/
│       └── feature.png
```

## Availability Commands

### List Country Availability

```bash
gpc availability list --track production
```

### Update Country Targeting

```bash
gpc availability update --track production --countries US,GB,DE,FR --confirm
```

## Diff (Compare Draft vs Live)

```bash
gpc diff                       # All sections
gpc diff --section listings    # Just listings
```

## Global Flags

All commands support: `--package/-p`, `--output/-o` (json/table/tsv/csv/yaml/minimal), `--pretty`, `--quiet/-q`, `--debug`, `--dry-run`, `--timeout`, `--profile`.
