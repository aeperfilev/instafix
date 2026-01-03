# Instafix

**Instafix** prepares images for Instagram formats with configurable backgrounds, padding, borders, and optional watermark text. The core processing logic is provided as a Go package, with a CLI and a web service on top.

## Features

- Resize to Instagram formats (square, portrait, landscape, story).
- Auto format selection by aspect ratio.
- Backgrounds: solid, blur, stretch, average.
- Padding and borders.
- Watermark styling (text provided at runtime).
- DNG/RAW preview support (uses embedded JPEG preview).
- Configurable profiles for reuse.

## Project Structure

- `config/` config schema and default profiles file.
- `cmd/cli/` CLI entrypoint.
- `cmd/service/` HTTP service entrypoint.
- `docs/` technical notes and package interfaces.
- `assets/` fonts used for watermarks.

### Config

Default config is `config/profiles.toml`. If `--config` is not provided, Instafix searches in this order:
1. `INSTAFIX_CONFIG` (if set)
2. `./profiles.toml`
3. `./config/profiles.toml`
4. `profiles.toml` next to the executable

Watermark text is not stored in config. You must pass it explicitly when calling CLI or HTTP API; if omitted, no watermark is drawn.

## CLI

**Build:**

```shell
go build -o instafix ./cmd/cli
```

**Usage:**

```shell
./instafix --profile default --watermark "@name" input.jpg
./instafix --config config/profiles.toml --profile white_passepartout --out output.jpg input.jpg
```

## Web Service

**Build:**

```shell
go build -o instafix-server ./cmd/service
```

**Run:**

```shell
API_KEY=secret ./instafix-server --config config/profiles.toml --addr :8080
```

**API:**

- `POST /fix` (multipart form field `image`)
- Query params:
  - `profile` (default: `default`)
  - `watermark` (optional)
- Header:
  - `X-API-Key` (required if `API_KEY` is set)

**Example:**

```shell
curl -X POST \
  -H "X-API-Key: secret" \
  -F "image=@input.jpg" \
  "http://localhost:8080/fix?profile=black&watermark=@name"
```

## Docs

- Technical details: `docs/technical.md`
- Config proposal: `docs/profiles.md`

## Tests

```shell
go test ./...
```
