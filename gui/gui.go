package gui

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed frontend/dist frontend/src/style.css frontend/*.html frontend/app/app.html frontend/settings/settings.html frontend/web/web.html frontend/welcome/welcome.html
var FrontendFS embed.FS

var Templates *template.Template

func init() {
	fs, err := fs.Sub(FrontendFS, "frontend")
	if err != nil {
		panic(err)
	}
	Templates = template.Must(template.ParseFS(fs, "*.html", "*/*.html"))
}
