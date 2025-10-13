package gui

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed frontend
var FrontendFS embed.FS

var Templates *template.Template

func init() {
	fs, err := fs.Sub(FrontendFS, "frontend")
	if err != nil {
		panic(err)
	}
	Templates = template.Must(template.ParseFS(fs, "*.html", "*/*.html"))
}
