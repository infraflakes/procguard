package auth

import (
	"fmt"
	"syscall"

	"golang.org/x/term"
	"procguard/internal/auth"
	"procguard/internal/config"
)

// Check performs a password check for a CLI command. If no password is set, it prompts to create one.
func Check() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("could not load config: %w", err)
	}

	if cfg.PasswordHash == "" {
		fmt.Println("No password set. Please create one.")
		fmt.Print("Enter New Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("could not read password: %w", err)
		}
		password := string(bytePassword)
		fmt.Println()

		fmt.Print("Confirm New Password: ")
		byteConfirm, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("could not read confirmation password: %w", err)
		}
		confirm := string(byteConfirm)
		fmt.Println()

		if password != confirm {
			return fmt.Errorf("passwords do not match")
		}

		hash, err := auth.HashPassword(password)
		if err != nil {
			return fmt.Errorf("could not hash password: %w", err)
		}

		cfg.PasswordHash = hash
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("could not save new password: %w", err)
		}

		fmt.Println("Password has been set.")
		return nil
	}

	// If password is set, prompt for it
	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("could not read password: %w", err)
	}
	password := string(bytePassword)
	fmt.Println()

	if !auth.CheckPasswordHash(password, cfg.PasswordHash) {
		return fmt.Errorf("invalid password")
	}

	return nil
}
