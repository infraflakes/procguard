package run

import (
	"fmt"
	"procguard/cmd/daemon"
	"procguard/internal/database"
	"procguard/internal/logger"
	"procguard/services/api"

	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a ProcGuard service (api or daemon)",
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Run the ProcGuard API server",
	Run: func(cmd *cobra.Command, args []string) {
		const defaultPort = "58141"
		addr := "127.0.0.1:" + defaultPort
		fmt.Println("Starting API server on http://" + addr)
		api.StartWebServer(addr)
	},
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run the ProcGuard daemon",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting ProcGuard daemon...")
		appLogger := logger.Get()
		db, err := database.InitDB()
		if err != nil {
			appLogger.Fatalf("Failed to initialize database: %v", err)
		}
		defer db.Close()

		daemon.Start(appLogger, db)
		// Keep the main goroutine alive
		select {}
	},
}

func init() {
	RunCmd.AddCommand(apiCmd, daemonCmd, nativeHostCmd)
}
