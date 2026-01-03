# Technical Overview

## Package API

The core package is `github.com/aeperfilev/instafix/pkg/instafix`:

- `NewProcessor(cfg config.Config) (*Processor, error)`
  Validates config and returns a processor instance.

- `(*Processor) Process(src image.Image, profileName, watermarkText string) (image.Image, int, error)`
  Applies a profile and returns the resulting image and JPEG quality.

Config package: `github.com/aeperfilev/instafix/config`
- `Load(path string) (Config, error)`
- `LoadDefault() (Config, string, error)`
- `FindDefaultPath() (string, error)`
- `Config.ResolveProfile(name string) (ResolvedProfile, error)`

**Config Resolution:**

1. Profiles reference registries by `*_ref` fields.
2. `format_ref` must point to a fixed or auto format.
3. Auto formats resolve to the closest fixed format by aspect ratio.
4. `watermark_ref` is optional; watermark text comes from runtime input.

**Processing Pipeline:**

1. Create canvas using the resolved target format size.
2. Render background:
   - solid: fill color
   - blur: fill + blur
   - stretch: resize to canvas size (distortion allowed)
   - average: compute average color and fill
3. Fit source image into the canvas while keeping aspect ratio.
   - If `no_upscale` is true and the source is smaller than available space,
     do not scale up.
4. Draw optional border around the fitted image.
5. Draw the fitted image.
6. Draw watermark text if provided and style is present.

**DNG/RAW Handling:**

- If the input filename ends with `.dng` or `.raw`, Instafix tries to extract the
  embedded JPEG preview and decodes it with EXIF orientation applied.
- This preserves Snapseed edits baked into the preview without external tools.

**Error Model:**

- Invalid request inputs (missing profile or watermark style) return `UserError`.
- Rendering errors (font missing, invalid config) are treated as server errors.

## HTTP API

Endpoint: `POST /fix`
- Multipart field: `image`
- Query params:
  - `profile` (default: `default`)
  - `watermark` (optional)
- Auth: `X-API-Key` header if `API_KEY` env var is set.

## Default Config Search

When `--config` is not provided, the search order is:
1. `INSTAFIX_CONFIG`
2. `./profiles.toml`
3. `./config/profiles.toml`
4. `profiles.toml` next to the executable

**Config Reference**

See `docs/profiles.md` for the full registry/profile schema and examples.
