package static

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed all:out
var staticFS embed.FS

var StaticFS, _ = fs.Sub(staticFS, "out")

func HasFrontend() bool {
	f, err := StaticFS.Open("index.html")
	if err != nil {
		return false
	}
	f.Close()

	entries, err := fs.ReadDir(StaticFS, "_next/static/chunks")
	if err != nil {
		return false
	}

	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".js") {
			return true
		}
	}

	return false
}
