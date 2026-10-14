# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`mdview` is a single-binary Go CLI that converts a markdown file to styled HTML and opens it in the
default browser. It's essentially one file, `main.go`, plus a handful of embedded assets and
`main_test.go`.

## Build & run

Requires Go (see `go.mod`/`mise.toml` for version) and, for packaging, `just`, `pandoc`, and
[`goreleaser`](https://goreleaser.com/).

- Quick local build: `go build` (produces `./mdview`)
- Run directly without building: `go run . <filename.md>`
- Full cross-platform build + packaging via `just` (uses `justfile`, which thinly wraps
  `goreleaser`; config lives in `.goreleaser.yaml`):
  - `just` or `just build` — cross-compiles all platform binaries via `goreleaser build --snapshot` (linux amd64/arm64/386, windows amd64, darwin amd64/arm64, freebsd amd64)
  - `just release-snapshot` — full local dry run producing archives + `.deb`/`.rpm` packages in `dist/` without publishing
  - `just release` — the real release (`goreleaser release --clean`); publishes a draft GitHub release, run from CI on tag push
  - `just check` — validates `.goreleaser.yaml`
  - `just snap` — build snap package via `snapcraft` (kept separate; goreleaser doesn't build snaps)
  - `just clean` — remove build artifacts (`dist/`, `mdview.1`, `*.snap`)
  - `just manpage` regenerates `mdview.1` from `mdview.1.md` via pandoc; goreleaser runs this automatically as a `before.hooks` step.
  - Versioning is read directly from git tags by goreleaser (embedded via `-ldflags -X main.appVersion={{.Version}}`) — there's no `VERSION` env var to pass anymore.
- Run tests with `go test ./...`. There is no lint target configured; `go vet ./...` and `gofmt` are reasonable sanity checks before committing.

## Architecture

Everything happens in `main()` in `main.go`, in a straight-line pipeline:

1. **Flag parsing** — `-o`, `-v/-version`, `-h/-help`, `-b/-bare`. First positional arg is the input markdown file; a single `-` reads the markdown from stdin instead, in which case `baseDir` for image resolution is the working directory rather than the input file's directory.
2. **Image inlining** (`newMarkdown` → `imageTransformer` + `htmlImageRenderer` → `embedImage`/`imageToDataURI`) — rewrites relative image references into base64 `data:` URIs so the output HTML is fully self-contained/offline-viewable. Markdown images (inline and reference-style) are rewritten by an AST transformer, so images inside code spans and code blocks are left alone. Raw `<img>` tags (`src` and `srcset`) are rewritten by a custom renderer for `HTMLBlock`/`RawHTML` nodes. `embedImage` strips query/fragment and decodes percent-encoding. `getMimeType` uses the extension, then content sniffing, and skips files that are not images. This includes path-traversal guards (caps `..` traversal depth) and a 10MB per-image size cap.
3. **Markdown parsing** via Goldmark (built by `newMarkdown`), configured with:
   - `extension.GFM` (tables, task lists, strikethrough, etc.)
   - `extension.Typographer` (smart quotes/dashes)
   - `go.abhg.dev/goldmark/mermaid` with `NoScript: true` (diagram rendering is handled ourselves)
   - `html.WithUnsafe()` (raw HTML passthrough is required for the image-tag handling above)
4. **Title extraction** (`getTitleFromAST`/`extractText`) — walks the parsed AST for the first H1 to use as the `<title>`.
5. **Mermaid script embedding** (`embedMermaidScript`) — only injects the embedded `mermaid.min.js` + an init script (with light/dark theme detection via `prefers-color-scheme`) if the rendered HTML actually contains a mermaid block, keeping non-diagram output lean.
6. **Templating & output** — `template.html` and `github-markdown.css` are embedded via `//go:embed` and combined with the rendered content via `fmt.Fprintf`. `-bare` skips the CSS. Output goes to `-o <path>` if given, otherwise a random temp filename (see below).
7. **Launch** — opens the resulting HTML file via `github.com/pkg/browser`.

### Temp file / Snap sandboxing (`getTempDir`, `isSnap`)

Output location resolution order: `MDVIEW_DIR` env var → (if running as a Snap, per `SNAP_USER_COMMON` env var) `~/mdview-temp` → OS default temp dir. This exists because Snap-sandboxed browsers (e.g. Snap Firefox) can't read from `/tmp`, only from the user's home directory — `check()` also prints a Snap-specific hint on file errors for this reason.

### Packaging

`.goreleaser.yaml` drives cross-platform binary builds, archives (`.tar.gz`/`.zip`), and `.deb`/
`.rpm` packages (via goreleaser's built-in `nfpm` integration — no more hand-rolled `control` file
or `dpkg-deb` invocation). `justfile` wraps the `goreleaser` CLI so `just <target>` stays the single
local entry point. `mdview.1.md` (man page source) and `snap/` (built separately via `just snap`,
outside goreleaser) round out packaging.

`.github/workflows/release.yml` triggers on tag push, runs `just release`, and publishes a draft
GitHub release (goreleaser reads the version straight from the pushed git tag — no more
previous-tag lookup step). `.github/workflows/build.yml` runs `just build` (a `goreleaser build
--snapshot`) on every push as a CI sanity check (no tests to run).
