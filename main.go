package main

import (
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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

	if os.Getenv("SNAP_USER_COMMON") != "" {
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
		panic(e)
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
