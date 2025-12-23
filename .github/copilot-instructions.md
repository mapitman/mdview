# Copilot Instructions for mdview

## Project Overview

mdview is a command-line tool written in Go that formats Markdown files and launches them in a web browser. It converts Markdown to HTML with GitHub-style CSS and supports various features including image embedding, custom output paths, and bare HTML output.

## Technology Stack

- **Language**: Go 1.21.1
- **Main Dependencies**:
  - `github.com/pkg/browser` - for opening files in the default browser
  - `gitlab.com/golang-commonmark/markdown` - for markdown parsing and rendering
- **Build Tool**: Make
- **Documentation**: Pandoc (for man page generation)

## Project Structure

- `main.go` - Single-file application containing all core logic
- `github-markdown.css` - Embedded GitHub-style CSS for rendered output
- `template.html` - Embedded HTML template for output
- `mdview.1.md` - Man page in Markdown format (converted to `mdview.1` via pandoc)
- `Makefile` - Build automation for multiple platforms
- `.github/workflows/` - CI/CD workflows for building and releasing

## Build Commands

### Building for Development
```bash
# Build for current platform
go build -o mdview .

# Build for specific platform with version
VERSION=v1.0.0 make linux        # Linux (amd64, arm64, i386)
VERSION=v1.0.0 make windows      # Windows (amd64)
VERSION=v1.0.0 make darwin       # macOS (amd64, arm64)
VERSION=v1.0.0 make freebsd      # FreeBSD (amd64)
VERSION=v1.0.0 make all          # All platforms
```

### Building Packages
```bash
VERSION=v1.0.0 make deb          # Debian packages (amd64, arm64)
make snap                         # Snap package
```

### Testing
Currently, there is no formal test suite in this project.

### Linting
```bash
go vet ./...
go fmt ./...
```

## Code Style and Conventions

### General Go Conventions
- Follow standard Go formatting (use `go fmt`)
- Use standard Go idioms and practices
- Keep code simple and readable
- Use descriptive variable names

### Project-Specific Conventions

1. **Version Handling**:
   - Version is injected at build time via `-ldflags "-X main.appVersion=$(VERSION)"`
   - VERSION variable should not include 'v' prefix in Makefile operations
   - Strip 'v' prefix with `${VERSION#v}` when needed

2. **Error Handling**:
   - Use the `check()` function for error handling
   - Provide user-friendly error messages for common errors (permission, file not found)
   - Include special handling for Snap installations with helpful messages

3. **Embedded Resources**:
   - CSS and HTML template are embedded using `//go:embed` directives
   - Keep embedded resources in the root directory

4. **Flag Handling**:
   - Support both short (`-v`, `-h`, `-b`) and long (`-version`, `-help`, `-bare`) flag forms
   - Use the `flag` package for command-line argument parsing

5. **File Handling**:
   - Support custom temp directory via `MDVIEW_DIR` environment variable
   - Handle Snap installations specially (restricted to HOME directory)
   - Use proper file permissions (0700 for directories)

6. **Image Processing**:
   - Convert relative image paths to data URIs for portability
   - Process both Markdown image syntax and HTML `<img>` tags
   - Use regex for HTML image processing

## Important Implementation Details

1. **Browser Integration**:
   - Silence browser stdout/stderr to avoid cluttering output
   - Use `browser.OpenFile()` after writing HTML

2. **Temporary Files**:
   - Generate random hex-based filenames for temp files
   - Respect `MDVIEW_DIR` environment variable
   - Handle Snap sandboxing restrictions

3. **Markdown Processing**:
   - Enable HTML, tables, and typographer features
   - Extract first heading as page title
   - Process tokens to handle embedded images

## Platform Support

mdview is built for multiple platforms:
- Linux: amd64, arm64, i386
- Windows: amd64
- macOS (Darwin): amd64, arm64
- FreeBSD: amd64

## Packaging

The project supports multiple package formats:
- **Debian/Ubuntu**: `.deb` packages via `dpkg-deb`
- **Snap**: Snapcraft packages (note: sandboxing restrictions apply)
- **Tarballs**: `.tar.gz` archives for Linux, macOS, FreeBSD
- **ZIP**: `.zip` archives for Windows

## Release Process

Releases are automated via GitHub Actions (`.github/workflows/release.yml`):
1. Triggered on tag push
2. Builds for all platforms
3. Creates Debian packages
4. Generates release artifacts
5. Uploads to GitHub Releases

## Dependencies

When adding new dependencies:
- Keep dependencies minimal
- Prefer standard library when possible
- Update `go.mod` and `go.sum` properly
- Test cross-platform builds after adding dependencies

## Documentation

- Keep README.md up to date with usage examples
- Update man page (`mdview.1.md`) when changing command-line flags
- Man page is generated from Markdown using pandoc

## Common Tasks

### Adding a New Flag
1. Add flag variable using `flag.String()`, `flag.Bool()`, etc.
2. Add both short and long forms using `flag.BoolVar()` or similar
3. Update help text in error message
4. Update man page (`mdview.1.md`)
5. Update README.md examples

### Modifying HTML Output
1. Edit `template.html` for structure changes
2. Edit `github-markdown.css` for style changes
3. Both files are embedded at build time

### Adding Platform Support
1. Add new target in Makefile
2. Follow existing pattern for `bin/<os>-<arch>/mdview`
3. Create appropriate archive (tar.gz for Unix-like, zip for Windows)
4. Update README.md installation section

## CI/CD

- **Build workflow** (`.github/workflows/build.yml`): Runs on all pushes
- **Release workflow** (`.github/workflows/release.yml`): Runs on tag pushes
- Both workflows use `make` targets for building

## Things to Avoid

- Don't add test files or test dependencies unless implementing a proper test suite
- Don't modify the build process without testing all platforms
- Don't change embedded resources without considering existing users
- Don't break backward compatibility with command-line flags
- Don't add unnecessary dependencies
- Avoid making changes that would affect cross-platform builds
