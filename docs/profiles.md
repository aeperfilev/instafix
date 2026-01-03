# Profiles Config Proposal (Registry + Profiles, With Auto Format)

**Goals:**

- Human-friendly, small number of concepts.
- Registry lookups for shared styles (backgrounds, watermarks, formats).
- Profiles only compose by references + small layout primitives.
- Clear validation rules for auto vs fixed formats.

**Key decisions:**

1. Separate registries: backgrounds, watermarks, formats.
2. Profiles reference registries via *_ref fields.
3. Formats are typed: fixed (width/height) or auto (from_list).
4. Add no_upscale to avoid enlarging small images.
5. Consistent snake_case naming.

**Example config (TOML):**

```toml
# --- Global settings ---
[settings]
jpeg_quality = 100
assets_path = "./assets"

# --- Registry: Backgrounds ---
[backgrounds.solid_black]
type = "solid"
color = "#000000"

[backgrounds.solid_white]
type = "solid"
color = "#ffffff"

[backgrounds.blur_soft]
type = "blur"
blur_radius = 20.0

[backgrounds.blur_hard]
type = "blur"
blur_radius = 60.0
darken = 0.4

# --- Registry: Watermarks (style only, text at runtime) ---
[watermarks.standard]
font = "Roboto-Bold.ttf"
size = 24
color = "#ffffff"
outline = true
outline_color = "#0f0f0f"
opacity = 0.5
align = "bottom-center"
offset_y = 5

# --- Registry: Formats ---
[formats.square]
type = "fixed"
width = 1080
height = 1080

[formats.portrait]
type = "fixed"
width = 1080
height = 1350

[formats.story]
type = "fixed"
width = 1080
height = 1920

[formats.landscape]
type = "fixed"
width = 1080 # px
height = 566 # px

[formats.auto]
type = "auto"
from_list = ["square", "portrait", "story"]

# --- Profiles ---
[profiles.default]
background_ref = "solid_black"
watermark_ref = "standard"
format_ref = "auto"
padding_percent = 5.0
no_upscale = true

[profiles.blur_story]
background_ref = "blur_hard"
watermark_ref = "standard"
format_ref = "story"
padding_percent = 0.0
no_upscale = true

[profiles.framed_blur]
background_ref = "blur_soft"
format_ref = "square"
padding_percent = 10.0
border_width = 2
border_color = "#ffffff"
no_upscale = true
```

**Validation rules (simple):**

1. All *_ref fields must exist in corresponding registry.
2. format_ref must resolve to formats.*.
3. formats.* validation:
   - type="fixed" -> width/height required, from_list forbidden.
   - type="auto" -> from_list required, width/height forbidden.
4. no_upscale=true: if source is smaller than target, keep original size and only apply padding/background/watermark.

**Suggested Go structs (shape only):**

```go
type Config struct {
    Settings    Settings              `toml:"settings"`
    Backgrounds map[string]Background `toml:"backgrounds"`
    Watermarks  map[string]Watermark  `toml:"watermarks"`
    Formats     map[string]Format     `toml:"formats"`
    Profiles    map[string]Profile    `toml:"profiles"`
}

type Settings struct {
    JpegQuality int    `toml:"jpeg_quality"`
    AssetsPath  string `toml:"assets_path"`
}

type Profile struct {
    BackgroundRef  string   `toml:"background_ref"`
    WatermarkRef   string   `toml:"watermark_ref"`
    FormatRef      string   `toml:"format_ref"`
    PaddingPercent float64  `toml:"padding_percent"`
    BorderWidth    int      `toml:"border_width"`
    BorderColor    string   `toml:"border_color"`
    NoUpscale      bool     `toml:"no_upscale"`
}

type Format struct {
    Type     string   `toml:"type"`
    Width    int      `toml:"width"`
    Height   int      `toml:"height"`
    FromList []string `toml:"from_list"`
}

type Background struct {
    Type       string  `toml:"type"` # solid, blur, stretch, average
    Color      string  `toml:"color"`
    BlurRadius float64 `toml:"blur_radius"`
    Darken     float64 `toml:"darken"`
}

type Watermark struct {
    Font         string  `toml:"font"`
    Size         float64 `toml:"size"`
    Color        string  `toml:"color"`
    Opacity      float64 `toml:"opacity"`
    Align        string  `toml:"align"`
    OffsetX      float64 `toml:"offset_x"`
    OffsetY      float64 `toml:"offset_y"`
    Outline      bool    `toml:"outline"`
    OutlineColor string  `toml:"outline_color"`
    OutlineWidth float64 `toml:"outline_width"`
}
```

**Notes:**

- Watermark text is provided at runtime (CLI or HTTP); registry stores style only.
- Solid backgrounds can also be inlined later if you want fewer registry entries,
  but the registry-only approach is the most explicit and easiest to validate.
