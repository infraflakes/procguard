package block

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/term"
	"net/http"
	"os"
	"procguard/internal/auth"
	"procguard/internal/config"
	"syscall"

	"github.com/spf13/cobra"
)

func CheckAuth(cmd *cobra.Command) {
	token, _ := cmd.Flags().GetString("token")
	if token != "" {
		if verifyToken(token) {
			return
		}
	}

	// If token is not provided or invalid, prompt for password
	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading password:", err)
		os.Exit(1)
	}
	password := string(bytePassword)
	fmt.Println()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load config:", err)
		os.Exit(1)
	}

	if !auth.CheckPasswordHash(password, cfg.PasswordHash) {
		fmt.Fprintln(os.Stderr, "Invalid password")
		os.Exit(1)
	}
}

func verifyToken(token string) bool {
	const defaultPort = "58141"
	guiAddress := "127.0.0.1:" + defaultPort
	guiUrl := "http://" + guiAddress

	requestBody, err := json.Marshal(map[string]string{
		"token": token,
	})
	if err != nil {
		return false
	}

	resp, err := http.Post(guiUrl+"/api/verify-token", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var result map[string]bool
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	return result["valid"]
}
