package main

import (
	"crypto/rand"
	"encoding/base64"
	_ "embed"
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
	"gitlab.com/golang-commonmark/markdown"
)

var appVersion string

//go:embed github-markdown.css
var style string

//go:embed template.html
var template string

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

	md := markdown.New(
		markdown.HTML(true),
		markdown.Nofollow(true),
		markdown.Tables(true),
		markdown.Typographer(true))

	markdownTokens := md.Parse(dat)
	
	// Convert relative image links to data URIs
	baseDir := filepath.Dir(inputFilename)
	processImageTokens(markdownTokens, baseDir)
	
	html := md.RenderTokensToString(markdownTokens)
	title := getTitle(markdownTokens)

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

	_, err = fmt.Fprintf(f, template, actualStyle, title, html)
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

func getTitle(tokens []markdown.Token) string {
	var result string
	if len(tokens) > 0 {
		for i := 0; i < len(tokens); i++ {
			if topLevelHeading, ok := tokens[i].(*markdown.HeadingOpen); ok {
				for j := i + 1; j < len(tokens); j++ {
					if token, ok := tokens[j].(*markdown.HeadingClose); ok && token.Lvl == topLevelHeading.Lvl {
						break
					}
					result += getText(tokens[j])
				}
				result = strings.TrimSpace(result)
				break
			}
		}
	}
	
	return result
}

func getText(token markdown.Token) string {
	switch token := token.(type) {
	case *markdown.Text:
		return token.Content
	case *markdown.Inline:
		result := ""
		for _, token := range token.Children {
			result += getText(token)
		}
		return result
	}
	return ""

}

func isSnap() bool {
	return os.Getenv("SNAP_USER_COMMON") != ""
}

// processImageTokens walks through markdown tokens and converts relative image paths to data URIs
func processImageTokens(tokens []markdown.Token, baseDir string) {
	for _, token := range tokens {
		switch t := token.(type) {
		case *markdown.Image:
			if isRelativePath(t.Src) {
				if dataURI := imageToDataURI(t.Src, baseDir); dataURI != "" {
					t.Src = dataURI
				}
			}
		case *markdown.HTMLInline:
			// Process inline HTML that may contain <img> tags
			t.Content = processHTMLImages(t.Content, baseDir)
		case *markdown.HTMLBlock:
			// Process block HTML that may contain <img> tags
			t.Content = processHTMLImages(t.Content, baseDir)
		case *markdown.Inline:
			// Recursively process child tokens
			if t.Children != nil {
				processImageTokens(t.Children, baseDir)
			}
		}
	}
}

// processHTMLImages processes HTML content and converts relative image src attributes to data URIs
func processHTMLImages(html string, baseDir string) string {
	// Regular expression to match <img> tags with src attributes
	// This handles various formats: src="path", src='path', src=path
	imgRegex := regexp.MustCompile(`(<img[^>]*\ssrc=)(['"]?)([^'"\s>]+)(['"]?)`)
	
	result := imgRegex.ReplaceAllStringFunc(html, func(match string) string {
		// Extract the parts using the regex
		parts := imgRegex.FindStringSubmatch(match)
		if len(parts) != 5 {
			return match
		}
		
		prefix := parts[1]      // "<img...src="
		openQuote := parts[2]   // " or ' or empty
		srcPath := parts[3]     // the actual path
		closeQuote := parts[4]  // " or ' or empty
		
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
		if err != nil || strings.HasPrefix(relPath, "..") {
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
