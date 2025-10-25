package gui

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed frontend/dist frontend/src/style.css frontend/*.html frontend/app/app.html frontend/settings/settings.html frontend/web/web.html frontend/welcome/welcome.html
var FrontendFS embed.FS

// Templates holds the parsed HTML templates for the web UI.
var Templates *template.Template

// The init function is executed on package initialization.
// It parses the HTML templates embedded in the FrontendFS.
func init() {
	fs, err := fs.Sub(FrontendFS, "frontend")
	if err != nil {
		panic(err)
	}
	Templates = template.Must(template.ParseFS(fs, "*.html", "*/*.html"))
}
