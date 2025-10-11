package api

import (
	_ "embed"
)

//go:embed templates/dashboard.html
var dashboardHTML []byte

//go:embed templates/login.html
var loginHTML []byte
