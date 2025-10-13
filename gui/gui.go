package gui

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed templates
var templatesFS embed.FS

var (
	Templates *template.Template
	LoginHTML []byte
)

func init() {
	// We need to strip the 'templates' prefix from the path for the template names.
	fs, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		panic(err)
	}
	Templates = template.Must(template.ParseFS(fs, "*.html"))

	LoginHTML, err = templatesFS.ReadFile("templates/login.html")
	if err != nil {
		panic(err)
	}
}