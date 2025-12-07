# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

img2ppt is a Go backend service that converts images to structured PowerPoint presentations using AI.

**Pipeline:** Image → Gemini (content extraction) → Image Generation → PPT Rendering → Storage → URL

## Build & Run

```bash
# Install dependencies
go mod tidy

# Run server
go run cmd/server/main.go

# Build binary
go build -o img2ppt cmd/server/main.go

# Run with config
CONFIG_PATH=config.yaml ./img2ppt
```

## Configuration

Environment variables (or config.yaml):

- `GEMINI_API_KEY` - Google Gemini API key (required)
- `IMAGEGEN_API_KEY` - Image generation API key (uses same Gemini key)
- `SERVER_ADDR` - Server address (default: `:8080`)
- `STORAGE_TYPE` - Storage type: `local`, `s3`, `gcs` (default: `local`)

## Architecture

```text
cmd/server/main.go          # Entry point
internal/
  api/                      # HTTP handlers, router, DTOs
  service/
    orchestrator/           # Pipeline coordinator
    gemini/                 # Image analysis (Gemini API)
    imagegen/               # Image generation (Gemini Image API)
    ppt/                    # PPTX rendering (unioffice)
    storage/                # File storage (local/S3/GCS)
  infra/
    config/                 # YAML + env config
    logger/                 # Zap logger wrapper
    httpclient/             # HTTP client with retry
    limiter/                # Rate limiting
pkg/
  errors/                   # Custom error types
```

## API

### POST /v1/image-to-ppt

```json
{
  "image_base64": "...",
  "language": "zh-CN",
  "style": "consulting_minimal"
}
```

Response:

```json
{
  "request_id": "uuid",
  "status": "SUCCEEDED",
  "ppt_url": "/files/uuid.pptx",
  "meta": { "title": "..." }
}
```

## Key Dependencies

- `gin-gonic/gin` - HTTP framework
- `unidoc/unioffice` - PPTX generation
- `go.uber.org/zap` - Logging
- `golang.org/x/time/rate` - Rate limiting
