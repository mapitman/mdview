# Style Inspector — Chrome Extension

A lightweight Chrome extension that instantly extracts the design system of any website: fonts, colors, paddings, and logo URL — all in one click.

## Demo

[![Style Inspector Demo](https://img.youtube.com/vi/NnzUSkLEZ3M/0.jpg)](https://www.youtube.com/watch?v=NnzUSkLEZ3M)

## What it does

Open any website, click the extension icon — and you'll see:

- **Logo** — image preview with a direct link (falls back to favicon if no logo found)
- **Background colors** — HEX chips, click any chip to copy the code
- **Text colors** — same
- **Fonts** — list with a live preview (`Aa Бб`) in the actual font
- **Paddings** — all unique padding values used on the page

At the bottom there's a **ready-to-copy text summary**, e.g.:

```
Site: business-pad.com
Logo: https://business-pad.com/assets/logo.svg
Background colors: #FFFFFF, #F5F7FA, #1A1A2E
Text colors: #333333, #666666, #0056D2
Fonts: Inter, Roboto
Paddings: 8px, 16px, 24px, 32px, 48px
```

Paste it into a brief, a design chat, or a Notion doc — done.

## Export / Import / mdview

Three buttons in the popup toolbar:

- **Export style** — saves everything the inspector collected as a JSON file
  (`style-<domain>.json`). Share it with a teammate or archive it.
- **Import style** — loads a previously exported JSON file and renders it in the
  popup, even on pages the extension can't access (e.g. `chrome://` pages).
- **Inject into mdview** — derives a ready-to-use theme (`mdview-theme.json`)
  from the page's palette: background, text, the two most saturated colors as
  accents, surface/border tones, and the font stack. Apply it to the
  [mdview](https://github.com/neo37/mdview) markdown viewer:

  ```
  mdview --install-theme ~/Downloads/mdview-theme.json
  ```

  From then on every markdown file you open with `mdview` is rendered in the
  site's corporate style, logo included. `mdview --reset-theme` restores the
  default theme.

## Installation (Developer mode)

1. Download or clone this repository
2. Open Chrome → `chrome://extensions/`
3. Enable **Developer mode** (top right toggle)
4. Click **"Load unpacked"**
5. Select the `style-inspector` folder

No build step required — it's pure HTML + JS.

## How it works

When you click the popup, the extension injects `content.js` into the active tab using the Chrome Scripting API. The script scans computed styles of key elements (`header`, `nav`, `h1–h4`, `p`, `button`, `a`, etc.), deduplicates values by frequency, and returns the top results to the popup.

Logo detection checks:
- `<img>` with "logo" in `src`, `alt`, `class`, or `id`
- First `<img>` inside `header` or `nav`
- Inline SVG elements with a logo-related class
- `<link rel="icon">` as a fallback

## Files

| File | Purpose |
|------|---------|
| `manifest.json` | Extension config (Manifest V3) |
| `content.js` | Page analysis — extracts styles and logo |
| `popup.html` | Extension popup UI |
| `popup.js` | Renders results, handles copy |

## Permissions

- `activeTab` — access the current tab when the popup is opened
- `scripting` — inject the analysis script into the page

No data is sent anywhere. Everything runs locally.

## License

MIT
