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
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/browser"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/mermaid"
)

var appVersion string

// imgSrcRegex matches <img> tags with src attributes
// Captures: 1=prefix, 2=opening quote, 3=src path, 4=closing quote
var imgSrcRegex = regexp.MustCompile(`(<img[^>]*\ssrc=)(["']?)([^"'\s>]+)(["']?)`)

// cdnScriptRegex matches Mermaid CDN script tags followed by initialization script
// Goldmark's mermaid extension inserts these tags that need to be replaced with embedded version
// Pattern allows for flexible whitespace between elements
var cdnScriptRegex = regexp.MustCompile(`<script\s+src\s*=\s*"https://cdn\.jsdelivr\.net/npm/mermaid[^"]*"\s*>\s*</script>\s*<script[^>]*>\s*mermaid\.initialize\s*\([^)]*\)\s*;\s*</script>`)

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
		os.Stderr.WriteString("Usage:\nmdview [options] <filename>\nFormats markdown and launches it in a browser.\nIf the environment variable MDVIEW_DIR is set, the temporary file will be written there.\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	dat, err := os.ReadFile(inputFilename)
	check(err)

	// Convert relative image links to data URIs in the markdown source
	baseDir := filepath.Dir(inputFilename)
	processedMarkdown := processMarkdownImages(string(dat), baseDir)
	processedBytes := []byte(processedMarkdown)

	// Create Goldmark markdown processor with extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,        // GitHub Flavored Markdown (includes tables)
			extension.Typographer, // Smart quotes, dashes, etc.
			&mermaid.Extender{
				NoScript: true,   // Don't add CDN script tags - we'll add our own embedded version
			},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Auto-generate heading IDs
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // Allow raw HTML (equivalent to markdown.HTML(true))
		),
	)

	// Parse markdown to AST
	doc := md.Parser().Parse(text.NewReader(processedBytes))
	
	// Extract title from AST
	title := getTitleFromAST(doc, processedBytes)
	
	// Render to HTML
	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, processedBytes, doc); err != nil {
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

	_, err = fmt.Fprintf(f, template, actualStyle, title, htmlContent)
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
			fmt.Println("https://github.com/mapitman/mdview?tab=readme-ov-file#installation\n")
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
			title = strings.TrimSpace(buf.String())
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

// processMarkdownImages processes markdown source and converts relative image paths to data URIs
func processMarkdownImages(markdown string, baseDir string) string {
	// Process markdown image syntax: ![alt](path)
	imgMarkdownRegex := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	markdown = imgMarkdownRegex.ReplaceAllStringFunc(markdown, func(match string) string {
		parts := imgMarkdownRegex.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		alt := parts[1]
		imgPath := parts[2]
		
		if isRelativePath(imgPath) {
			if dataURI := imageToDataURI(imgPath, baseDir); dataURI != "" {
				return fmt.Sprintf("![%s](%s)", alt, dataURI)
			}
		}
		return match
	})
	
	// Process HTML img tags in markdown
	markdown = processHTMLImages(markdown, baseDir)
	
	return markdown
}

// processHTMLImages processes HTML content and converts relative image src attributes to data URIs
func processHTMLImages(html string, baseDir string) string {
	// Use the package-level regex to match <img> tags with src attributes
	result := imgSrcRegex.ReplaceAllStringFunc(html, func(match string) string {
		// Extract the parts using the regex
		parts := imgSrcRegex.FindStringSubmatch(match)
		if len(parts) != 5 {
			return match
		}
		
		prefix := parts[1]      // "<img...src="
		openQuote := parts[2]   // " or ' or empty
		srcPath := parts[3]     // the actual path
		closeQuote := parts[4]  // " or ' or empty
		
		// If quotes don't match, return original (malformed HTML)
		if openQuote != closeQuote {
			return match
		}
		
		// Check if the path is relative
		if isRelativePath(srcPath) {
			if dataURI := imageToDataURI(srcPath, baseDir); dataURI != "" {
				return prefix + openQuote + dataURI + closeQuote
			}
		}
		
		return match
	})
	
	return result
}

// isRelativePath checks if a path is relative (not http://, https://, //, or absolute path)
func isRelativePath(path string) bool {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return false
	}
	if strings.HasPrefix(path, "//") {
		return false
	}
	if strings.HasPrefix(path, "data:") {
		return false
	}
	if filepath.IsAbs(path) {
		return false
	}
	return true
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
	
	// Determine MIME type based on file extension
	mimeType := getMimeType(cleanedPath)
	
	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(data)
	
	// Return data URI
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
}

// getMimeType returns the MIME type based on file extension
func getMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
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
	case ".bmp":
		return "image/bmp"
	case ".ico":
		return "image/x-icon"
	default:
		// For unknown extensions, log a warning but try with generic image type
		log.Printf("Warning: Unknown image extension %s for file %s, using image/* MIME type", ext, path)
		return "image/*"
	}
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
