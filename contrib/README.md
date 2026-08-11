# contrib: Style Inspector + themed viewer

Companion tools that let `mdview` users render markdown in the visual style of
any website — extract a site's design tokens with a Chrome extension, save them
as a theme file, and apply that theme to a markdown viewer.

## What's inside

| Path | What it is |
|------|------------|
| `chrome-extension/` | **Style Inspector** — a Chrome extension (Manifest V3) that extracts a site's logo, colors, fonts and paddings, and exports them as JSON |
| `mdview-themed` | A self-contained bash markdown viewer (marked + mermaid, fully offline) that renders `.md` files using an installable theme |
| `install-mdview-themed.sh` | Installer: copies the script to `~/.local/bin` and fetches marked/mermaid |
| `mdview-theme.example.json` | Example theme file |

## Workflow

1. Install the extension: `chrome://extensions` → Developer mode → *Load unpacked* → `contrib/chrome-extension`.
2. Open any website, click the extension icon, press **Inject into mdview** —
   the extension derives a theme from the page (background, text color, the two
   most saturated colors as accents, font stack, logo) and downloads
   `mdview-theme.json`.
3. Apply it:

   ```sh
   ./contrib/install-mdview-themed.sh
   mdview-themed --install-theme ~/Downloads/mdview-theme.json
   mdview-themed README.md
   ```

Every markdown file now opens styled like the site: its colors, its fonts, its
logo in the header. Mermaid diagrams are re-colored to match.

`--reset-theme` restores the built-in dark theme, `--show-theme` prints the
active one.

## Theme file format

```json
{
  "bg": "#FFFFFF",        // page background
  "text": "#1A1A1A",      // main text color
  "accent": "#7C88FC",    // links, table headers, rules
  "accent2": "#FF8562",   // hover, blockquote border
  "surface": "#F5F5F7",   // code blocks, quotes, even table rows
  "line": "#E6E6E6",      // borders
  "fonts": ["Manrope", "Roboto", "Arial"],
  "logo": "https://…/logo.svg",   // URL or data URI; downloaded on install
  "site": "example.com"           // logo links here
}
```

Only `bg` and `text` are required. The extension's **Export style** / **Import
style** buttons additionally let you save and share the full extracted token
set between machines and teammates.
