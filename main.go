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
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	var outfilePtr = flag.String("o", "", "Output filename. (Optional)")
	var versionPtr = flag.Bool("version", false, "Prints mdview version.")
	var helpPtr = flag.Bool("help", false, "Prints mdview help message.")
	var barePtr = flag.Bool("bare", false, "Bare HTML with no style applied.")
	flag.BoolVar(versionPtr, "v", false, "Prints mdview version.")
	flag.BoolVar(helpPtr, "h", false, "Prints mdview help message.")
	flag.BoolVar(barePtr, "b", false, "Bare HTML with no style applied.")

	flag.Parse()
	inputFilename := flag.Arg(0)

	if *versionPtr {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	if inputFilename == "" || *helpPtr {
		os.Stderr.WriteString("Usage:\nmdview [options] <filename>\nFormats markdown and launches it in a browser.\nUse - as the filename to read markdown from standard input.\nIf the environment variable MDVIEW_DIR is set, the temporary file will be written there.\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// A filename of - means read the markdown from standard input. Relative image
	// paths are then resolved against the working directory, as there is no input
	// file to anchor them to.
	var dat []byte
	var err error
	baseDir := "."

	if inputFilename == "-" {
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
	if err := md.Renderer().Render(&buf, dat, doc); err != nil {
		log.Fatal(err)
	}
	htmlContent := buf.String()
	
	// Add embedded Mermaid.js and initialization script when diagrams are present
	htmlContent = embedMermaidScript(htmlContent)

	outfilePath := *outfilePtr
	if outfilePath == "" {
		outfilePath = tempFileName("mdview", ".html")
	}

	f, err := os.Create(outfilePath)
	check(err)
	defer f.Close()
	var actualStyle string
	if *barePtr {
		actualStyle = ""
	} else {
		actualStyle = style
	}

	_, err = fmt.Fprintf(f, template, actualStyle, stdhtml.EscapeString(title), htmlContent)
	check(err)
	f.Sync()
	browser.Stderr = nil
	browser.Stdout = nil
	err = browser.OpenFile(outfilePath)
	check(err)
}

func tempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(getTempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}

func getTempDir() string {
	if os.Getenv("MDVIEW_DIR") != "" {
		var tempDir = os.Getenv(("MDVIEW_DIR"))
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			err = os.Mkdir(tempDir, 0700)
			check(err)
		}

		return tempDir
	}

	if isSnap() {
		var tmpdir = os.Getenv("HOME") + "/mdview-temp"
		if _, err := os.Stat(tmpdir); os.IsNotExist(err) {
			err = os.Mkdir(tmpdir, 0700)
			check(err)
		}
		return tmpdir
	}

	return os.TempDir()
}

func check(e error) {
	if e != nil {
		if errors.Is(e, fs.ErrPermission) {
			fmt.Println("There was a permission error accessing the file.")
		}

		if errors.Is(e, fs.ErrNotExist) {
			fmt.Println("mdview was unable to find the file.")
		}

		if isSnap() {
			fmt.Println("Since mdview was installed as a Snap, it can only access files in your HOME directory.")
			fmt.Println("If you need to use it with files outside of your HOME directory, choose a different installation method.")
			fmt.Println("https://github.com/mapitman/mdview?tab=readme-ov-file#installation")
			fmt.Println()
		}

		log.Fatal(e)
	}
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

// imageToDataURI reads an image file and converts it to a base64 data URI
func imageToDataURI(imagePath string, baseDir string) string {
	// Resolve the full path relative to the markdown file
	fullPath := filepath.Join(baseDir, imagePath)
	
	// Clean and validate the path to prevent path traversal attacks
	cleanedPath, err := filepath.Abs(fullPath)
	if err != nil {
		log.Printf("Warning: Invalid image path %s: %v", fullPath, err)
		return ""
	}
	
	// Ensure the resolved path is within or relative to the base directory
	cleanedBase, err := filepath.Abs(baseDir)
	if err != nil {
		log.Printf("Warning: Invalid base directory %s: %v", baseDir, err)
		return ""
	}
	
	// Check if the cleaned path starts with the base directory or is a reasonable relative reference
	// We allow accessing parent directories for flexibility with markdown repos
	if !strings.HasPrefix(cleanedPath, cleanedBase) {
		relPath, err := filepath.Rel(cleanedBase, cleanedPath)
		if err != nil {
			log.Printf("Warning: Unable to determine relative path for %s: %v", imagePath, err)
			return ""
		}
		
		// If the path goes outside the base directory, check parent traversal limits
		if strings.HasPrefix(relPath, "..") {
			// Allow up to 3 levels of parent directory traversal for flexibility
			// Count the number of ".." path components
			components := strings.Split(filepath.ToSlash(relPath), "/")
			parentLevels := 0
			for _, component := range components {
				if component == ".." {
					parentLevels++
				}
			}
			if parentLevels > 3 {
				log.Printf("Warning: Image path %s goes too many levels above base directory", imagePath)
				return ""
			}
		}
	}
	
	// Check file size before reading (limit to 10MB to prevent memory issues)
	fileInfo, err := os.Stat(cleanedPath)
	if err != nil {
		log.Printf("Warning: Unable to stat image file %s: %v", cleanedPath, err)
		return ""
	}
	
	const maxSize = 10 * 1024 * 1024 // 10MB
	if fileInfo.Size() > maxSize {
		log.Printf("Warning: Image file %s is too large (%d bytes, max %d bytes)", cleanedPath, fileInfo.Size(), maxSize)
		return ""
	}
	
	// Read the image file
	data, err := os.ReadFile(cleanedPath)
	if err != nil {
		log.Printf("Warning: Unable to read image file %s: %v", cleanedPath, err)
		return ""
	}
	
	mimeType := getMimeType(cleanedPath, data)
	if mimeType == "" {
		log.Printf("Warning: File %s is not a recognized image type, skipping", cleanedPath)
		return ""
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(data)
	
	// Return data URI
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
}

// getMimeType returns the image MIME type for a file, or "" when it is not an image.
// It trusts the file extension first, then sniffs the content.
func getMimeType(path string, data []byte) string {
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
	if detected := http.DetectContentType(data); strings.HasPrefix(detected, "image/") {
		return detected
	}
	return ""
}

// embedMermaidScript adds the embedded Mermaid.js library to the HTML content
// Since we set NoScript: true on the Mermaid extender, we need to manually add the script
func embedMermaidScript(htmlContent string) string {
	// Check if there are any mermaid diagrams in the content
	// The Goldmark mermaid extension uses class="mermaid" for diagram blocks
	if !strings.Contains(htmlContent, `class="mermaid"`) {
		return htmlContent // No mermaid diagrams, don't add the script
	}
	
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
