# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`mdview` is a single-binary Go CLI that converts a markdown file to styled HTML and opens it in the
default browser. It's essentially one file, `main.go`, plus a handful of embedded assets. There are
no other Go source files and no test files currently in the repo.

## Build & run

Requires Go (see `go.mod`/`mise.toml` for version) and, for packaging targets, `just` and `pandoc`.

- Quick local build: `go build` (produces `./mdview`)
- Run directly without building: `go run . <filename.md>`
- Full cross-platform build via `just` (uses `justfile`, the project's task runner — replaced the old Makefile):
  - `just` or `just linux` — Linux amd64/arm64/i386 builds
  - `just all` — linux + windows + darwin + freebsd
  - `just deb` — build `.deb` packages (requires a prior linux build)
  - `just snap` — build snap package via `snapcraft`
  - `just clean` — remove build artifacts
  - Pass `VERSION=x.y.z` as an env var to any target to stamp the version (embedded via `-ldflags -X main.appVersion=...`); CI derives this from the previous git tag.
  - `just manpage` regenerates `mdview.1` from `mdview.1.md` via pandoc.
- There is no test suite and no lint target configured; `go vet ./...` and `gofmt` are reasonable sanity checks before committing.

## Architecture

Everything happens in `main()` in `main.go`, in a straight-line pipeline:

1. **Flag parsing** — `-o`, `-v/-version`, `-h/-help`, `-b/-bare`. First positional arg is the input markdown file.
2. **Image inlining** (`processMarkdownImages` → `processHTMLImages`/markdown image regex + `imageToDataURI`) — rewrites relative image references (both `![]()`  markdown syntax and raw `<img src=...>` HTML) into base64 `data:` URIs *before* markdown parsing, so the output HTML is fully self-contained/offline-viewable. This includes path-traversal guards (caps `..` traversal depth) and a 10MB per-image size cap.
3. **Markdown parsing** via Goldmark, configured with:
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

`justfile`, `control` (Debian control file template, `VERSION` substituted at build time), `mdview.1.md`
(man page source) and `snap/` all support the release pipeline in `.github/workflows/release.yml`,
which triggers on tag push, builds all platform targets + deb packages via `just`, and publishes a
draft GitHub release with `softprops/action-gh-release`. `.github/workflows/build.yml` runs the same
`just all` build on every push as a CI check (no tests to run).
