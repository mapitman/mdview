package main

import (
	"bytes"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	stdhtml "html"
	"io"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"go.abhg.dev/goldmark/mermaid"
)

var appVersion string

// imgTagRegex matches whole <img> tags, allowing > inside quoted attribute values
var imgTagRegex = regexp.MustCompile(`(?is)<img\b(?:"[^"]*"|'[^']*'|[^>"'])*>`)

// imgAttrRegex matches the src and srcset attributes of an <img> tag
// Captures: 1=name and equals sign (with leading whitespace), 2=double-quoted value,
// 3=single-quoted value, 4=unquoted value
var imgAttrRegex = regexp.MustCompile(`(?i)(\s(?:src|srcset)\s*=\s*)(?:"([^"]*)"|'([^']*)'|([^\s"'>]+))`)

// srcsetCandidateRegex splits one srcset candidate into whitespace, URL and descriptor
var srcsetCandidateRegex = regexp.MustCompile(`^(\s*)(\S+)([\s\S]*)$`)

// uriSchemeRegex matches a URI scheme such as http:, file: or mailto:
var uriSchemeRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]+:`)

//go:embed github-markdown.css
var style string

//go:embed template.html
var template string

//go:embed mermaid.min.js
var mermaidJS string

func main() {
	// Warnings go to stderr without a timestamp, like the messages from check()
	log.SetFlags(0)
	log.SetPrefix("mdview: ")

	var outfilePtr = flag.String("o", "", "Output filename. (Optional)")
	var versionPtr = flag.Bool("version", false, "Prints mdview version.")
	var helpPtr = flag.Bool("help", false, "Prints mdview help message.")
	var barePtr = flag.Bool("bare", false, "Bare HTML with no style applied.")
	flag.BoolVar(versionPtr, "v", false, "Prints mdview version.")
	flag.BoolVar(helpPtr, "h", false, "Prints mdview help message.")
	flag.BoolVar(barePtr, "b", false, "Bare HTML with no style applied.")

	flag.Usage = func() { printUsage(os.Stderr) }
	args, err := parseFlags(flag.CommandLine, os.Args[1:])
	if err != nil {
		os.Exit(2)
	}

	if *versionPtr {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	if *helpPtr {
		printUsage(os.Stdout)
		os.Exit(0)
	}

	if len(args) == 0 {
		printUsage(os.Stderr)
		os.Exit(1)
	}

	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "mdview accepts one filename, but received %d: %s\n\n", len(args), strings.Join(args, " "))
		printUsage(os.Stderr)
		os.Exit(1)
	}

	inputFilename := args[0]

	// A filename of - means read the markdown from standard input. Relative image
	// paths are then resolved against the working directory, as there is no input
	// file to anchor them to.
	var dat []byte
	baseDir := "."

	if inputFilename == "-" {
		if stdinIsTerminal() {
			fmt.Fprintln(os.Stderr, "mdview: no markdown on standard input. Pipe markdown into mdview or pass a filename.")
			os.Exit(1)
		}
		dat, err = io.ReadAll(os.Stdin)
		check(err)
	} else {
		dat, err = os.ReadFile(inputFilename)
		check(err)
		baseDir = filepath.Dir(inputFilename)
	}

	md := newMarkdown(baseDir)

	// Parse markdown to AST
	doc := md.Parser().Parse(text.NewReader(dat))

	// Extract title from AST
	title := getTitleFromAST(doc, dat)

	// Render to HTML
	var buf bytes.Buffer
	check(md.Renderer().Render(&buf, dat, doc))
	htmlContent := buf.String()

	// Add embedded Mermaid.js and initialization script when diagrams are present
	if hasMermaidBlock(doc) {
		htmlContent = embedMermaidScript(htmlContent)
	}

	outfilePath := *outfilePtr
	if outfilePath == "" {
		tempDir := getTempDir()
		removeStaleTempFiles(tempDir, staleTempFileAge)
		outfilePath = tempFileName(tempDir, "mdview", ".html")
	}

	actualStyle := style
	if *barePtr {
		actualStyle = ""
	}

	check(writeOutput(outfilePath, actualStyle, stdhtml.EscapeString(title), htmlContent))
	browser.Stderr = nil
	browser.Stdout = nil
	err = browser.OpenFile(outfilePath)
	check(err)
}

// printUsage writes the usage text and the option defaults to w
func printUsage(w io.Writer) {
	fmt.Fprint(w, "Usage:\nmdview [options] <filename>\nFormats markdown and launches it in a browser.\nUse - as the filename to read markdown from standard input.\nIf the environment variable MDVIEW_DIR is set, the temporary file will be written there.\n")
	flag.CommandLine.SetOutput(w)
	flag.PrintDefaults()
}

// parseFlags parses args and returns the positional arguments. Unlike FlagSet.Parse
// alone, it accepts flags before, between and after the positional arguments.
// Everything after a -- argument is positional.
func parseFlags(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		consumed := len(args) - fs.NArg()
		if consumed > 0 && args[consumed-1] == "--" {
			return append(positional, fs.Args()...), nil
		}
		if fs.NArg() == 0 {
			return positional, nil
		}
		positional = append(positional, fs.Arg(0))
		args = fs.Args()[1:]
	}
}

// stdinIsTerminal reports whether standard input is an interactive terminal
func stdinIsTerminal() bool {
	info, err := os.Stdin.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// writeOutput writes the HTML document to path and reports any write or close error
func writeOutput(path string, style string, title string, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(f, template, style, title, content); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func tempFileName(dir, prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(dir, prefix+hex.EncodeToString(randBytes)+suffix)
}

// staleTempFileAge is how old a generated temp file must be before mdview removes it
const staleTempFileAge = 24 * time.Hour

// tempFileRegex matches the names produced by tempFileName, so cleanup never touches other files
var tempFileRegex = regexp.MustCompile(`^mdview[0-9a-f]{32}\.html$`)

// removeStaleTempFiles deletes generated temp files in dir that are older than maxAge.
// The browser reads the current file after mdview exits, so only older files are removed.
func removeStaleTempFiles(dir string, maxAge time.Duration) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !tempFileRegex.MatchString(entry.Name()) {
			continue
		}
		if info, err := entry.Info(); err == nil && time.Since(info.ModTime()) > maxAge {
			os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
}

func getTempDir() string {
	if os.Getenv("MDVIEW_DIR") != "" {
		var tempDir = os.Getenv(("MDVIEW_DIR"))
		check(os.MkdirAll(tempDir, 0700))
		return tempDir
	}

	if isSnap() {
		var tmpdir = filepath.Join(os.Getenv("HOME"), "mdview-temp")
		check(os.MkdirAll(tmpdir, 0700))
		return tmpdir
	}

	return os.TempDir()
}

// check prints a failure to standard error and exits with status 1
func check(e error) {
	if e == nil {
		return
	}

	permission := errors.Is(e, fs.ErrPermission)
	notExist := errors.Is(e, fs.ErrNotExist)

	// Name the path, because the failing file can be the input, the output or the output directory
	target := "the file"
	var pathErr *fs.PathError
	if errors.As(e, &pathErr) {
		target = pathErr.Path
	}

	if permission {
		fmt.Fprintf(os.Stderr, "There was a permission error accessing %s.\n", target)
	}

	if notExist {
		fmt.Fprintf(os.Stderr, "mdview was unable to find %s.\n", target)
	}

	// The Snap sandbox only blocks file access, so show its hint for those errors only
	if isSnap() && (permission || notExist) {
		fmt.Fprintln(os.Stderr, "Since mdview was installed as a Snap, it can only access files in your HOME directory.")
		fmt.Fprintln(os.Stderr, "If you need to use it with files outside of your HOME directory, choose a different installation method.")
		fmt.Fprintln(os.Stderr, "https://github.com/mapitman/mdview?tab=readme-ov-file#installation")
		fmt.Fprintln(os.Stderr)
	}

	fmt.Fprintln(os.Stderr, "mdview:", e)
	os.Exit(1)
}

func getTitleFromAST(doc ast.Node, source []byte) string {
	var title string
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if heading, ok := n.(*ast.Heading); ok && heading.Level == 1 {
			// Extract all text from heading and its children recursively
			var buf bytes.Buffer
			extractText(heading, source, &buf)
			title = stdhtml.UnescapeString(strings.TrimSpace(buf.String()))
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return title
}

// extractText recursively extracts text from a node and its children
func extractText(node ast.Node, source []byte, buf *bytes.Buffer) {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Text:
			buf.Write(n.Segment.Value(source))
		case *ast.String:
			buf.Write(n.Value)
		default:
			// Recursively extract text from other node types (emphasis, strong, links, etc.)
			extractText(child, source, buf)
		}
	}
}

func isSnap() bool {
	return os.Getenv("SNAP_USER_COMMON") != ""
}

// newMarkdown creates the Goldmark processor. Relative image references, in both
// markdown syntax and raw HTML, are embedded as data URIs while rendering.
func newMarkdown(baseDir string) goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,         // GitHub Flavored Markdown (includes tables)
			extension.Typographer, // Smart quotes, dashes, etc.
			&mermaid.Extender{
				NoScript: true, // Don't add CDN script tags - we'll add our own embedded version
			},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Auto-generate heading IDs
			parser.WithASTTransformers(util.Prioritized(&imageTransformer{baseDir: baseDir}, 100)),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // Allow raw HTML (equivalent to markdown.HTML(true))
			renderer.WithNodeRenderers(util.Prioritized(&htmlImageRenderer{baseDir: baseDir}, 100)),
		),
	)
}

// imageTransformer embeds relative markdown images (inline and reference-style).
// Working on the AST leaves images in code spans and code blocks untouched.
type imageTransformer struct {
	baseDir string
}

func (t *imageTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if img, ok := n.(*ast.Image); ok && entering {
			if dataURI := embedImage(string(img.Destination), t.baseDir); dataURI != "" {
				img.Destination = []byte(dataURI)
			}
		}
		return ast.WalkContinue, nil
	})
}

// htmlImageRenderer renders raw HTML like the default renderer with WithUnsafe,
// but embeds relative <img> sources first.
type htmlImageRenderer struct {
	baseDir string
}

func (r *htmlImageRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
}

func (r *htmlImageRenderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.HTMLBlock)
	if entering {
		var block []byte
		for i := 0; i < n.Lines().Len(); i++ {
			line := n.Lines().At(i)
			block = append(block, line.Value(source)...)
		}
		html.DefaultWriter.SecureWrite(w, []byte(processHTMLImages(string(block), r.baseDir)))
	} else if n.HasClosure() {
		closure := n.ClosureLine
		html.DefaultWriter.SecureWrite(w, []byte(processHTMLImages(string(closure.Value(source)), r.baseDir)))
	}
	return ast.WalkContinue, nil
}

func (r *htmlImageRenderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	n := node.(*ast.RawHTML)
	var raw []byte
	for i := 0; i < n.Segments.Len(); i++ {
		segment := n.Segments.At(i)
		raw = append(raw, segment.Value(source)...)
	}
	html.DefaultWriter.SecureWrite(w, []byte(processHTMLImages(string(raw), r.baseDir)))
	return ast.WalkSkipChildren, nil
}

// embedImage returns a data URI for a relative image reference, or "" when the
// reference is not a relative path or the file cannot be embedded.
func embedImage(src string, baseDir string) string {
	if !isRelativePath(src) {
		return ""
	}
	// Drop any query or fragment, then decode percent-encoding
	if i := strings.IndexAny(src, "?#"); i >= 0 {
		src = src[:i]
	}
	if src == "" {
		return ""
	}
	if decoded, err := url.PathUnescape(src); err == nil {
		src = decoded
	}
	return imageToDataURI(src, baseDir)
}

// processHTMLImages converts relative src and srcset values of <img> tags to data URIs
func processHTMLImages(htmlSource string, baseDir string) string {
	return imgTagRegex.ReplaceAllStringFunc(htmlSource, func(tag string) string {
		return imgAttrRegex.ReplaceAllStringFunc(tag, func(attr string) string {
			parts := imgAttrRegex.FindStringSubmatch(attr)
			prefix := parts[1]
			value, quote := parts[4], ""
			if rest := attr[len(prefix):]; rest[0] == '"' || rest[0] == '\'' {
				value, quote = rest[1:len(rest)-1], rest[:1]
			}

			if strings.EqualFold(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(prefix), "=")), "srcset") {
				return prefix + quote + embedSrcset(value, baseDir) + quote
			}
			if dataURI := embedImage(value, baseDir); dataURI != "" {
				return prefix + quote + dataURI + quote
			}
			return attr
		})
	})
}

// embedSrcset converts the relative URLs in a srcset value, keeping descriptors intact.
// A value that already contains a data URI is left alone, as its commas make splitting unsafe.
func embedSrcset(srcset string, baseDir string) string {
	if strings.Contains(srcset, "data:") {
		return srcset
	}
	candidates := strings.Split(srcset, ",")
	for i, candidate := range candidates {
		parts := srcsetCandidateRegex.FindStringSubmatch(candidate)
		if parts == nil {
			continue
		}
		if dataURI := embedImage(parts[2], baseDir); dataURI != "" {
			candidates[i] = parts[1] + dataURI + parts[3]
		}
	}
	return strings.Join(candidates, ",")
}

// isRelativePath checks if a path is relative (no URI scheme, not //, and not an absolute path)
func isRelativePath(path string) bool {
	if path == "" || strings.HasPrefix(path, "//") || uriSchemeRegex.MatchString(path) {
		return false
	}
	return !filepath.IsAbs(path)
}

// maxImageSize is the largest image file that is embedded
const maxImageSize = 10 * 1024 * 1024 // 10MB

// maxParentLevels is how many directory levels above the base directory an image may be
const maxParentLevels = 3

// imageToDataURI reads an image file and converts it to a base64 data URI.
// It returns "" and logs a warning when the file must not or cannot be embedded.
func imageToDataURI(imagePath string, baseDir string) string {
	// Check the extension first so that non-image files are never opened
	mimeType := getMimeType(imagePath)
	if mimeType == "" {
		log.Printf("Warning: %s does not have a supported image extension, skipping", imagePath)
		return ""
	}

	// Resolve symlinks in both paths so that a link cannot point outside the allowed area
	resolvedBase, err := filepath.EvalSymlinks(baseDir)
	if err != nil {
		log.Printf("Warning: Invalid base directory %s: %v", baseDir, err)
		return ""
	}
	resolvedBase, err = filepath.Abs(resolvedBase)
	if err != nil {
		log.Printf("Warning: Invalid base directory %s: %v", baseDir, err)
		return ""
	}
	resolvedPath, err := filepath.EvalSymlinks(filepath.Join(resolvedBase, imagePath))
	if err != nil {
		log.Printf("Warning: Unable to resolve image file %s: %v", imagePath, err)
		return ""
	}

	// Allow images in the base directory, below it, or a few levels above it
	// (for markdown repos that keep images in a sibling directory)
	relPath, err := filepath.Rel(resolvedBase, resolvedPath)
	if err != nil {
		log.Printf("Warning: Unable to determine relative path for %s: %v", imagePath, err)
		return ""
	}
	parentLevels := 0
	for _, component := range strings.Split(filepath.ToSlash(relPath), "/") {
		if component != ".." {
			break
		}
		parentLevels++
	}
	if parentLevels > maxParentLevels {
		log.Printf("Warning: Image path %s goes too many levels above base directory", imagePath)
		return ""
	}

	// Read at most one byte past the limit, so a file that grows after the check is still capped
	f, err := os.Open(resolvedPath)
	if err != nil {
		log.Printf("Warning: Unable to read image file %s: %v", resolvedPath, err)
		return ""
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		log.Printf("Warning: Unable to stat image file %s: %v", resolvedPath, err)
		return ""
	}
	if !fileInfo.Mode().IsRegular() {
		log.Printf("Warning: Image file %s is not a regular file", resolvedPath)
		return ""
	}

	data, err := io.ReadAll(io.LimitReader(f, maxImageSize+1))
	if err != nil {
		log.Printf("Warning: Unable to read image file %s: %v", resolvedPath, err)
		return ""
	}
	if len(data) > maxImageSize {
		log.Printf("Warning: Image file %s is too large (max %d bytes)", resolvedPath, maxImageSize)
		return ""
	}

	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
}

// getMimeType returns the MIME type for a supported image extension, or "" for any other file.
// The extension allowlist keeps non-image files (keys, configs) out of the output.
func getMimeType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".avif":
		return "image/avif"
	case ".bmp":
		return "image/bmp"
	case ".tif", ".tiff":
		return "image/tiff"
	case ".ico":
		return "image/x-icon"
	}
	return ""
}

// hasMermaidBlock reports whether the document contains a mermaid code block.
// Checking the AST ignores raw HTML that merely mentions class="mermaid".
func hasMermaidBlock(doc ast.Node) bool {
	found := false
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == mermaid.Kind {
			found = true
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return found
}

// embedMermaidScript adds the embedded Mermaid.js library to the HTML content
// Since we set NoScript: true on the Mermaid extender, we need to manually add the script
func embedMermaidScript(htmlContent string) string {
	// Escape any </script> tags (with closing >) inside the mermaid.js code
	// to prevent premature script closure. The standard way is to escape the forward slash.
	escapedMermaidJS := strings.ReplaceAll(mermaidJS, "</script>", "<\\/script>")

	// Add the embedded Mermaid.js and initialization at the end of the content
	// Initialize mermaid with theme detection
	initScript := `
    var isDarkMode = window.matchMedia('(prefers-color-scheme: dark)').matches;
    mermaid.initialize({
      startOnLoad: true,
      theme: isDarkMode ? 'dark' : 'default',
      flowchart: {
        useMaxWidth: true,
        htmlLabels: true,
        curve: 'linear'
      },
      securityLevel: 'strict'
    });
  `
	inlineScript := fmt.Sprintf("<script>%s</script><script>%s</script>", escapedMermaidJS, initScript)

	return htmlContent + inlineScript
}
