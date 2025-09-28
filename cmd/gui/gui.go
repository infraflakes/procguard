package gui

import (
	"fmt"
	_ "embed"

	"github.com/spf13/cobra"
)

//go:embed dashboard.html
var dashboardHTML []byte

//go:embed login.html
var loginHTML []byte

var GuiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Run the web-based GUI",
	Run:   runGUI,
}

func init() {}

func runGUI(cmd *cobra.Command, args []string) {
	const defaultPort = "58141"
	addr := "127.0.0.1:" + defaultPort
	fmt.Println("Starting GUI on http://" + addr)
	StartWebServer(addr)
}
