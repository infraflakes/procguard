//go:build !linux && !darwin && !windows

package cmd

import "fmt"

func openBrowser(url string) {
	fmt.Println("Please open your browser and navigate to:", url)
}
